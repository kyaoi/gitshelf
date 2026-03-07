package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newLsCommand(ctx *commandContext) *cobra.Command {
	var (
		kinds           []string
		statuses        []string
		tags            []string
		notKinds        []string
		notStatuses     []string
		notTags         []string
		includeArchived bool
		onlyArchived    bool
		format          string
		ready           bool
		depsBlocked     bool
		dueBefore       string
		dueAfter        string
		overdue         bool
		noDue           bool
		asJSON          bool
		parent          string
		limit           int
		search          string
	)

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List tasks",
		Example: "  shelf ls\n" +
			"  shelf ls --kind todo --status open --status in_progress\n" +
			"  shelf ls --tag backend --not-tag wip\n" +
			"  shelf ls --ready --overdue\n" +
			"  shelf ls --json",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := validateFormat(format, []string{"compact", "detail", "kanban"}); err != nil {
				return err
			}

			filter := shelf.TaskFilter{
				Kinds:           toKinds(kinds),
				Statuses:        toStatuses(statuses),
				Tags:            parseTagFlagValues(tags),
				NotKinds:        toKinds(notKinds),
				NotStatuses:     toStatuses(notStatuses),
				NotTags:         parseTagFlagValues(notTags),
				IncludeArchived: includeArchived,
				OnlyArchived:    onlyArchived,
				ReadyOnly:       ready,
				DepsBlocked:     depsBlocked,
				DueBefore:       dueBefore,
				DueAfter:        dueAfter,
				Overdue:         overdue,
				NoDue:           noDue,
				Parent:          parent,
				Limit:           limit,
				Search:          search,
			}

			tasks, err := shelf.ListTasks(ctx.rootDir, filter)
			if err != nil {
				return err
			}
			allTasks, err := shelf.NewTaskStore(ctx.rootDir).List()
			if err != nil {
				return err
			}
			titleByID := make(map[string]string, len(allTasks))
			for _, task := range allTasks {
				titleByID[task.ID] = task.Title
			}

			if asJSON {
				type lsItem struct {
					ID          string   `json:"id"`
					Title       string   `json:"title"`
					Kind        string   `json:"kind"`
					Status      string   `json:"status"`
					Tags        []string `json:"tags,omitempty"`
					DueOn       string   `json:"due_on,omitempty"`
					RepeatEvery string   `json:"repeat_every,omitempty"`
					ArchivedAt  string   `json:"archived_at,omitempty"`
					Parent      string   `json:"parent,omitempty"`
					ParentTitle string   `json:"parent_title,omitempty"`
				}
				items := make([]lsItem, 0, len(tasks))
				for _, task := range tasks {
					parentTitle := ""
					if task.Parent != "" {
						parentTitle = titleByID[task.Parent]
					}
					items = append(items, lsItem{
						ID:          task.ID,
						Title:       task.Title,
						Kind:        string(task.Kind),
						Status:      string(task.Status),
						Tags:        task.Tags,
						DueOn:       task.DueOn,
						RepeatEvery: task.RepeatEvery,
						ArchivedAt:  task.ArchivedAt,
						Parent:      task.Parent,
						ParentTitle: parentTitle,
					})
				}
				data, err := json.MarshalIndent(items, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			if format == "kanban" {
				statusOrder := []shelf.Status{"open", "in_progress", "blocked", "done", "cancelled"}
				grouped := map[shelf.Status][]shelf.Task{}
				for _, task := range tasks {
					grouped[task.Status] = append(grouped[task.Status], task)
				}
				for _, status := range statusOrder {
					fmt.Println(uiHeading(string(status) + ":"))
					rows := grouped[status]
					if len(rows) == 0 {
						fmt.Println(uiMuted("  (none)"))
						continue
					}
					for _, task := range rows {
						label := uiPrimary(task.Title)
						if ctx.showID {
							label = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(task.ID)), uiPrimary(task.Title))
						}
						dueText := uiMuted("-")
						if task.DueOn != "" {
							dueText = uiDue(task.DueOn)
						}
						fmt.Printf("  %s (%s) due=%s\n", label, uiKind(task.Kind), dueText)
					}
				}
				return nil
			}

			for _, task := range tasks {
				parentLabel := "root"
				if task.Parent != "" {
					if title, ok := titleByID[task.Parent]; ok {
						parentLabel = uiPrimary(title)
						if ctx.showID {
							parentLabel = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(task.Parent)), uiPrimary(title))
						}
					} else {
						parentLabel = uiMuted("(missing)")
					}
				} else {
					parentLabel = uiMuted(parentLabel)
				}
				label := uiPrimary(task.Title)
				if ctx.showID {
					label = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(task.ID)), uiPrimary(task.Title))
				}
				dueText := ""
				if task.DueOn != "" {
					dueText = fmt.Sprintf(" due=%s", uiDue(task.DueOn))
				}
				tagText := ""
				if len(task.Tags) > 0 {
					tagText = fmt.Sprintf(" tags=%s", strings.Join(task.Tags, ","))
				}
				archivedText := ""
				if task.ArchivedAt != "" {
					archivedText = " " + uiMuted("[archived]")
				}
				if format == "detail" {
					repeatText := "-"
					if task.RepeatEvery != "" {
						repeatText = task.RepeatEvery
					}
					fmt.Printf("%s kind=%s status=%s tags=%s due=%s repeat=%s archived_at=%q parent=%s\n", label, uiKind(task.Kind), uiStatus(task.Status), formatTagSummary(task.Tags), uiDue(task.DueOn), repeatText, task.ArchivedAt, parentLabel)
					continue
				}
				fmt.Printf("%s  (%s/%s)%s%s%s parent=%s\n", label, uiKind(task.Kind), uiStatus(task.Status), dueText, tagText, archivedText, parentLabel)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&kinds, "kind", nil, "Include kind (repeatable)")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Include status (repeatable)")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Include tag (repeatable)")
	cmd.Flags().StringArrayVar(&notKinds, "not-kind", nil, "Exclude kind (repeatable)")
	cmd.Flags().StringArrayVar(&notStatuses, "not-status", nil, "Exclude status (repeatable)")
	cmd.Flags().StringArrayVar(&notTags, "not-tag", nil, "Exclude tag (repeatable)")
	cmd.Flags().BoolVar(&includeArchived, "include-archived", false, "Include archived tasks")
	cmd.Flags().BoolVar(&onlyArchived, "only-archived", false, "Include only archived tasks")
	cmd.Flags().StringVar(&format, "format", "compact", "Output format: compact|detail|kanban")
	cmd.Flags().BoolVar(&ready, "ready", false, "Include only actionable tasks")
	cmd.Flags().BoolVar(&depsBlocked, "blocked-by-deps", false, "Include only tasks blocked by unresolved dependencies")
	cmd.Flags().StringVar(&dueBefore, "due-before", "", "Include only tasks due before this date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&dueAfter, "due-after", "", "Include only tasks due after this date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&overdue, "overdue", false, "Include only overdue tasks")
	cmd.Flags().BoolVar(&noDue, "no-due", false, "Include only tasks without due date")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&parent, "parent", "", "Filter by parent task ID or root")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of items")
	cmd.Flags().StringVar(&search, "search", "", "Search by title/body")
	return cmd
}

func newNextCommand(ctx *commandContext) *cobra.Command {
	var (
		limit  int
		asJSON bool
	)

	cmd := &cobra.Command{
		Use:   "next",
		Short: "List actionable tasks (ready to work on)",
		Example: "  shelf next\n" +
			"  shelf next --limit 20",
		RunE: func(_ *cobra.Command, _ []string) error {
			filter := shelf.TaskFilter{Limit: 0}
			tasks, err := shelf.ListTasks(ctx.rootDir, filter)
			if err != nil {
				return err
			}
			readiness, err := shelf.BuildTaskReadiness(ctx.rootDir)
			if err != nil {
				return err
			}

			titleByID := make(map[string]string, len(tasks))
			for _, task := range tasks {
				titleByID[task.ID] = task.Title
			}

			if asJSON {
				type nextItem struct {
					ID          string `json:"id"`
					Title       string `json:"title"`
					Kind        string `json:"kind"`
					Status      string `json:"status"`
					DueOn       string `json:"due_on,omitempty"`
					RepeatEvery string `json:"repeat_every,omitempty"`
					ArchivedAt  string `json:"archived_at,omitempty"`
					Parent      string `json:"parent,omitempty"`
					ParentTitle string `json:"parent_title,omitempty"`
				}
				items := make([]nextItem, 0)
				for _, task := range tasks {
					info, ok := readiness[task.ID]
					if !ok || !info.Ready {
						continue
					}
					parentTitle := ""
					if task.Parent != "" {
						parentTitle = titleByID[task.Parent]
					}
					items = append(items, nextItem{
						ID:          task.ID,
						Title:       task.Title,
						Kind:        string(task.Kind),
						Status:      string(task.Status),
						DueOn:       task.DueOn,
						RepeatEvery: task.RepeatEvery,
						ArchivedAt:  task.ArchivedAt,
						Parent:      task.Parent,
						ParentTitle: parentTitle,
					})
					if limit > 0 && len(items) >= limit {
						break
					}
				}
				data, err := json.MarshalIndent(items, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			count := 0
			for _, task := range tasks {
				info, ok := readiness[task.ID]
				if !ok || !info.Ready {
					continue
				}
				parentLabel := uiMuted("root")
				if task.Parent != "" {
					if title, ok := titleByID[task.Parent]; ok {
						parentLabel = uiPrimary(title)
						if ctx.showID {
							parentLabel = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(task.Parent)), uiPrimary(title))
						}
					} else {
						parentLabel = uiMuted("(missing)")
					}
				}
				label := uiPrimary(task.Title)
				if ctx.showID {
					label = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(task.ID)), uiPrimary(task.Title))
				}
				dueText := ""
				if task.DueOn != "" {
					dueText = fmt.Sprintf(" due=%s", uiDue(task.DueOn))
				}
				fmt.Printf("%s  (%s/%s)%s parent=%s\n", label, uiKind(task.Kind), uiStatus(task.Status), dueText, parentLabel)
				count++
				if limit > 0 && count >= limit {
					break
				}
			}
			if count == 0 {
				fmt.Println(uiMuted("(none)"))
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of items")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newTreeCommand(ctx *commandContext) *cobra.Command {
	var flags cockpitLaunchFlags

	cmd := &cobra.Command{
		Use:     "tree",
		Aliases: []string{"tr"},
		Short:   "Open Cockpit in tree mode",
		Example: "  shelf tree\n" +
			"  shelf tree --kind todo --not-status done --tag backend\n" +
			"  shelf tree --months 3 --status open",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !dailyCockpitIsTTY() {
				return fmt.Errorf("tree はTTYが必要です")
			}
			return runCockpitLaunch(ctx, cmd, calendarModeTree, flags)
		},
	}

	addCockpitLaunchFlags(cmd, &flags)
	return cmd
}

func printTreeNode(node shelf.TreeNode, prefix string, isLast bool, showID bool, format string) {
	branch := "├─ "
	nextPrefix := prefix + "│  "
	if isLast {
		branch = "└─ "
		nextPrefix = prefix + "   "
	}
	if prefix == "" {
		branch = ""
	}

	label := uiPrimary(node.Task.Title)
	if showID {
		label = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(node.Task.ID)), uiPrimary(node.Task.Title))
	}
	dueText := ""
	if node.Task.DueOn != "" {
		dueText = fmt.Sprintf(" due=%s", uiDue(node.Task.DueOn))
	}
	tagText := ""
	if len(node.Task.Tags) > 0 {
		tagText = fmt.Sprintf(" tags=%s", strings.Join(node.Task.Tags, ","))
	}
	if format == "detail" {
		repeatText := "-"
		if node.Task.RepeatEvery != "" {
			repeatText = node.Task.RepeatEvery
		}
		fmt.Printf("%s%s%s kind=%s status=%s tags=%s due=%s repeat=%s archived_at=%q\n", uiMuted(prefix), uiMuted(branch), label, uiKind(node.Task.Kind), uiStatus(node.Task.Status), formatTagSummary(node.Task.Tags), uiDue(node.Task.DueOn), repeatText, node.Task.ArchivedAt)
	} else {
		fmt.Printf("%s%s%s (%s/%s)%s%s\n", uiMuted(prefix), uiMuted(branch), label, uiKind(node.Task.Kind), uiStatus(node.Task.Status), dueText, tagText)
	}
	for i, child := range node.Children {
		printTreeNode(child, nextPrefix, i == len(node.Children)-1, showID, format)
	}
}

func toKinds(values []string) []shelf.Kind {
	kinds := make([]shelf.Kind, len(values))
	for i, value := range values {
		kinds[i] = shelf.Kind(value)
	}
	return kinds
}

func toStatuses(values []string) []shelf.Status {
	statuses := make([]shelf.Status, len(values))
	for i, value := range values {
		statuses[i] = shelf.Status(value)
	}
	return statuses
}

func buildTaskPath(task shelf.Task, byID map[string]shelf.Task) string {
	titles := []string{task.Title}
	current := task.Parent
	seen := map[string]struct{}{}
	for current != "" {
		if _, ok := seen[current]; ok {
			titles = append([]string{"(cycle)"}, titles...)
			break
		}
		seen[current] = struct{}{}

		parent, ok := byID[current]
		if !ok {
			titles = append([]string{"(missing)"}, titles...)
			break
		}
		titles = append([]string{parent.Title}, titles...)
		current = parent.Parent
	}
	return "root > " + strings.Join(titles, " > ")
}
