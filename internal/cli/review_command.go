package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

type reviewItem struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Kind        string   `json:"kind"`
	Status      string   `json:"status"`
	DueOn       string   `json:"due_on,omitempty"`
	Parent      string   `json:"parent,omitempty"`
	ParentTitle string   `json:"parent_title,omitempty"`
	BlockedBy   []string `json:"blocked_by,omitempty"`
}

type reviewPayload struct {
	Inbox   []reviewItem `json:"inbox"`
	Overdue []reviewItem `json:"overdue"`
	Today   []reviewItem `json:"today"`
	Blocked []reviewItem `json:"blocked"`
	Ready   []reviewItem `json:"ready"`
}

func newReviewCommand(ctx *commandContext) *cobra.Command {
	var (
		limit  int
		asJSON bool
		plain  bool
	)
	cmd := &cobra.Command{
		Use:     "review",
		Aliases: []string{"rv"},
		Short:   "Show a daily review of inbox, due, blocked, and ready tasks",
		Example: "  shelf review\n" +
			"  shelf review --limit 10\n" +
			"  shelf review --plain\n" +
			"  shelf review --json",
		RunE: func(_ *cobra.Command, _ []string) error {
			if resolveReviewOutputMode(dailyCockpitIsTTY(), asJSON, plain) == dailyCockpitOutputTUI {
				startDate, dayCount, err := resolveDailyCockpitRange(ctx.rootDir)
				if err != nil {
					return err
				}
				filter := shelf.TaskFilter{
					Statuses: activeStatusFilter(),
					Limit:    0,
				}
				return runCalendarModeTUIFn(ctx.rootDir, startDate, dayCount, filter.Statuses, calendarTUIOptions{
					Mode:         calendarModeReview,
					ShowID:       ctx.showID,
					SectionLimit: limit,
					Filter:       filter,
				})
			}
			payload, err := buildReviewPayload(ctx.rootDir, limit)
			if err != nil {
				return err
			}
			if asJSON {
				data, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf(
				"%s inbox=%d overdue=%d today=%d blocked=%d ready=%d\n",
				uiHeading("review:"),
				len(payload.Inbox),
				len(payload.Overdue),
				len(payload.Today),
				len(payload.Blocked),
				len(payload.Ready),
			)
			printReviewSection(ctx, "Inbox", payload.Inbox)
			printReviewSection(ctx, "Overdue", payload.Overdue)
			printReviewSection(ctx, "Today", payload.Today)
			printReviewSection(ctx, "Blocked", payload.Blocked)
			printReviewSection(ctx, "Ready", payload.Ready)
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 5, "Maximum items per section (0 means unlimited)")
	cmd.Flags().BoolVar(&plain, "plain", false, "Force plain text output even on TTY")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func buildReviewPayload(rootDir string, limit int) (reviewPayload, error) {
	tasks, err := shelf.NewTaskStore(rootDir).List()
	if err != nil {
		return reviewPayload{}, err
	}
	readiness, err := shelf.BuildTaskReadiness(rootDir)
	if err != nil {
		return reviewPayload{}, err
	}

	titleByID := make(map[string]string, len(tasks))
	for _, task := range tasks {
		titleByID[task.ID] = task.Title
	}

	today := time.Now().Local().Format("2006-01-02")
	payload := reviewPayload{
		Inbox:   []reviewItem{},
		Overdue: []reviewItem{},
		Today:   []reviewItem{},
		Blocked: []reviewItem{},
		Ready:   []reviewItem{},
	}
	for _, task := range tasks {
		if task.ArchivedAt != "" || !isReviewActiveStatus(task.Status) {
			continue
		}
		parentTitle := ""
		if task.Parent != "" {
			parentTitle = titleByID[task.Parent]
		}
		info := readiness[task.ID]
		item := reviewItem{
			ID:          task.ID,
			Title:       task.Title,
			Kind:        string(task.Kind),
			Status:      string(task.Status),
			DueOn:       task.DueOn,
			Parent:      task.Parent,
			ParentTitle: parentTitle,
			BlockedBy:   reviewBlockedBy(task, info, titleByID),
		}

		if task.Kind == "inbox" && task.Status == "open" {
			payload.Inbox = append(payload.Inbox, item)
		}
		if task.DueOn != "" && task.DueOn < today {
			payload.Overdue = append(payload.Overdue, item)
		}
		if task.DueOn == today {
			payload.Today = append(payload.Today, item)
		}
		if task.Status == "blocked" || info.BlockedByDeps {
			payload.Blocked = append(payload.Blocked, item)
		}
		if task.Kind != "inbox" && info.Ready {
			payload.Ready = append(payload.Ready, item)
		}
	}

	sortReviewItems(payload.Inbox)
	sortReviewItems(payload.Overdue)
	sortReviewItems(payload.Today)
	sortReviewItems(payload.Blocked)
	sortReviewItems(payload.Ready)
	payload.Inbox = limitReviewItems(payload.Inbox, limit)
	payload.Overdue = limitReviewItems(payload.Overdue, limit)
	payload.Today = limitReviewItems(payload.Today, limit)
	payload.Blocked = limitReviewItems(payload.Blocked, limit)
	payload.Ready = limitReviewItems(payload.Ready, limit)
	return payload, nil
}

func isReviewActiveStatus(status shelf.Status) bool {
	return status == "open" || status == "in_progress" || status == "blocked"
}

func reviewBlockedBy(task shelf.Task, info shelf.TaskReadiness, titleByID map[string]string) []string {
	reasons := make([]string, 0, 2)
	if task.Status == "blocked" {
		reasons = append(reasons, "status=blocked")
	}
	if len(info.UnresolvedDependsOn) > 0 {
		labels := make([]string, 0, len(info.UnresolvedDependsOn))
		for _, depID := range info.UnresolvedDependsOn {
			if title := strings.TrimSpace(titleByID[depID]); title != "" {
				labels = append(labels, title)
				continue
			}
			labels = append(labels, depID)
		}
		reasons = append(reasons, "depends_on: "+strings.Join(labels, ", "))
	}
	return reasons
}

func sortReviewItems(items []reviewItem) {
	sort.Slice(items, func(i, j int) bool {
		leftDue := strings.TrimSpace(items[i].DueOn)
		rightDue := strings.TrimSpace(items[j].DueOn)
		switch {
		case leftDue == "" && rightDue != "":
			return false
		case leftDue != "" && rightDue == "":
			return true
		case leftDue != rightDue:
			return leftDue < rightDue
		default:
			return items[i].ID < items[j].ID
		}
	})
}

func limitReviewItems(items []reviewItem, limit int) []reviewItem {
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

func printReviewSection(ctx *commandContext, title string, items []reviewItem) {
	fmt.Println(uiHeading(title + ":"))
	if len(items) == 0 {
		fmt.Println(uiMuted("  (none)"))
		return
	}
	for _, item := range items {
		label := uiPrimary(item.Title)
		if ctx.showID {
			label = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(item.ID)), uiPrimary(item.Title))
		}
		dueText := uiMuted("-")
		if item.DueOn != "" {
			dueText = uiDue(item.DueOn)
		}
		parentText := uiMuted("root")
		if item.ParentTitle != "" {
			parentText = uiPrimary(item.ParentTitle)
		}
		blockedText := ""
		if len(item.BlockedBy) > 0 {
			blockedText = fmt.Sprintf(" blocked_by=%s", uiMuted(strings.Join(item.BlockedBy, "; ")))
		}
		fmt.Printf("  %s (%s/%s) due=%s parent=%s%s\n", label, uiKind(shelf.Kind(item.Kind)), uiStatus(shelf.Status(item.Status)), dueText, parentText, blockedText)
	}
}
