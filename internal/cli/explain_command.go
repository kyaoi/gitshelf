package cli

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

type filterExplain struct {
	Match   bool     `json:"match"`
	Reasons []string `json:"reasons,omitempty"`
}

func newExplainCommand(ctx *commandContext) *cobra.Command {
	var (
		view   string
		asJSON bool
	)

	cmd := &cobra.Command{
		Use:   "explain <id>",
		Short: "Explain why a task matches (or does not match) views/readiness",
		Example: "  shelf explain <id>\n" +
			"  shelf explain <id> --view active\n" +
			"  shelf explain <id> --json",
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "説明対象のタスクを選択", nil, true)
			if err != nil {
				return err
			}
			task, err := shelf.EnsureTaskExists(ctx.rootDir, id)
			if err != nil {
				return err
			}
			readinessMap, err := shelf.BuildTaskReadiness(ctx.rootDir)
			if err != nil {
				return err
			}
			readiness := readinessMap[task.ID]
			allTasks, err := shelf.NewTaskStore(ctx.rootDir).List()
			if err != nil {
				return err
			}
			byID := make(map[string]shelf.Task, len(allTasks))
			for _, item := range allTasks {
				byID[item.ID] = item
			}

			builtin := map[string]shelf.TaskFilter{}
			for _, name := range []string{"active", "ready", "blocked", "overdue"} {
				filter, ok := builtinTaskView(name)
				if !ok {
					continue
				}
				builtin[name] = filter
			}
			builtinResult := map[string]filterExplain{}
			for name, filter := range builtin {
				result, err := explainTaskByFilter(task, filter, readiness)
				if err != nil {
					return err
				}
				builtinResult[name] = result
			}

			defaultLS, err := explainTaskByFilter(task, shelf.TaskFilter{}, readiness)
			if err != nil {
				return err
			}
			nextFilter := shelf.TaskFilter{
				Statuses:  []shelf.Status{"open", "in_progress"},
				ReadyOnly: true,
			}
			defaultNext, err := explainTaskByFilter(task, nextFilter, readiness)
			if err != nil {
				return err
			}
			agendaFilter := shelf.TaskFilter{
				Statuses: []shelf.Status{"open", "in_progress", "blocked"},
			}
			defaultAgenda, err := explainTaskByFilter(task, agendaFilter, readiness)
			if err != nil {
				return err
			}
			todayFilter := shelf.TaskFilter{
				Statuses: []shelf.Status{"open", "in_progress", "blocked"},
			}
			defaultToday, err := explainTaskByFilter(task, todayFilter, readiness)
			if err != nil {
				return err
			}
			today := time.Now().Local().Format("2006-01-02")
			if task.DueOn != "" && task.DueOn > today {
				defaultToday.Match = false
				defaultToday.Reasons = append(defaultToday.Reasons, "due_on is after today")
			} else if task.DueOn == "" {
				defaultToday.Match = false
				defaultToday.Reasons = append(defaultToday.Reasons, "due_on is empty")
			}

			var requested *struct {
				Name   string        `json:"name"`
				Result filterExplain `json:"result"`
			}
			if strings.TrimSpace(view) != "" {
				filter, err := resolveTaskView(ctx.rootDir, view)
				if err != nil {
					return err
				}
				result, err := explainTaskByFilter(task, filter, readiness)
				if err != nil {
					return err
				}
				requested = &struct {
					Name   string        `json:"name"`
					Result filterExplain `json:"result"`
				}{
					Name:   view,
					Result: result,
				}
			}

			if asJSON {
				payload := map[string]any{
					"task": map[string]any{
						"id":           task.ID,
						"title":        task.Title,
						"kind":         task.Kind,
						"status":       task.Status,
						"due_on":       task.DueOn,
						"repeat_every": task.RepeatEvery,
						"archived_at":  task.ArchivedAt,
						"parent":       task.Parent,
					},
					"readiness": map[string]any{
						"ready":                 readiness.Ready,
						"blocked_by_deps":       readiness.BlockedByDeps,
						"unresolved_depends_on": readiness.UnresolvedDependsOn,
					},
					"builtin_views": builtinResult,
					"default_commands": map[string]any{
						"ls":     defaultLS,
						"next":   defaultNext,
						"agenda": defaultAgenda,
						"today":  defaultToday,
					},
				}
				if requested != nil {
					payload["requested_view"] = requested
				}
				data, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("Task: %s (%s/%s)\n", uiPrimary(task.Title), uiKind(task.Kind), uiStatus(task.Status))
			if task.DueOn != "" {
				fmt.Printf("due_on: %s\n", uiDue(task.DueOn))
			} else {
				fmt.Printf("due_on: %s\n", uiMuted("(none)"))
			}
			fmt.Println(uiHeading("Readiness:"))
			fmt.Printf("  ready=%t\n", readiness.Ready)
			if len(readiness.UnresolvedDependsOn) == 0 {
				fmt.Println("  unresolved depends_on: (none)")
			} else {
				labels := make([]string, 0, len(readiness.UnresolvedDependsOn))
				for _, depID := range readiness.UnresolvedDependsOn {
					labels = append(labels, taskLabelForLink(depID, byID, ctx.showID))
				}
				fmt.Printf("  unresolved depends_on: %s\n", strings.Join(labels, ", "))
			}

			fmt.Println(uiHeading("Built-in Views:"))
			for _, name := range []string{"active", "ready", "blocked", "overdue"} {
				result := builtinResult[name]
				printFilterExplain(name, result)
			}

			fmt.Println(uiHeading("Default Commands:"))
			printFilterExplain("ls", defaultLS)
			printFilterExplain("next", defaultNext)
			printFilterExplain("agenda", defaultAgenda)
			printFilterExplain("today", defaultToday)

			if requested != nil {
				fmt.Println(uiHeading("Requested View:"))
				printFilterExplain(requested.Name, requested.Result)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&view, "view", "", "Evaluate this built-in/custom view")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func printFilterExplain(name string, result filterExplain) {
	match := uiColor("match=false", "31")
	if result.Match {
		match = uiColor("match=true", "32")
	}
	fmt.Printf("  - %s: %s\n", name, match)
	for _, reason := range result.Reasons {
		fmt.Printf("      * %s\n", reason)
	}
}

func explainTaskByFilter(task shelf.Task, filter shelf.TaskFilter, readiness shelf.TaskReadiness) (filterExplain, error) {
	reasons := make([]string, 0)

	if filter.OnlyArchived {
		if task.ArchivedAt == "" {
			reasons = append(reasons, "task is not archived")
		}
	} else if !filter.IncludeArchived && task.ArchivedAt != "" {
		reasons = append(reasons, "archived tasks are excluded")
	}

	if len(filter.Kinds) > 0 && !slices.Contains(filter.Kinds, task.Kind) {
		reasons = append(reasons, fmt.Sprintf("kind %q is not in include kinds", task.Kind))
	}
	if len(filter.Statuses) > 0 && !slices.Contains(filter.Statuses, task.Status) {
		reasons = append(reasons, fmt.Sprintf("status %q is not in include statuses", task.Status))
	}
	if slices.Contains(filter.NotKinds, task.Kind) {
		reasons = append(reasons, fmt.Sprintf("kind %q is excluded", task.Kind))
	}
	if slices.Contains(filter.NotStatuses, task.Status) {
		reasons = append(reasons, fmt.Sprintf("status %q is excluded", task.Status))
	}

	if filter.ReadyOnly && !readiness.Ready {
		reasons = append(reasons, "task is not ready")
	}
	if filter.DepsBlocked && !readiness.BlockedByDeps {
		reasons = append(reasons, "task is not blocked by dependencies")
	}

	if filter.NoDue && task.DueOn != "" {
		reasons = append(reasons, "task has due_on")
	}
	if strings.TrimSpace(filter.DueBefore) != "" {
		dueBefore, err := shelf.NormalizeDueOn(filter.DueBefore)
		if err != nil {
			return filterExplain{}, err
		}
		if task.DueOn == "" || task.DueOn >= dueBefore {
			reasons = append(reasons, fmt.Sprintf("due_on is not before %s", dueBefore))
		}
	}
	if strings.TrimSpace(filter.DueAfter) != "" {
		dueAfter, err := shelf.NormalizeDueOn(filter.DueAfter)
		if err != nil {
			return filterExplain{}, err
		}
		if task.DueOn == "" || task.DueOn <= dueAfter {
			reasons = append(reasons, fmt.Sprintf("due_on is not after %s", dueAfter))
		}
	}
	if filter.Overdue {
		today := time.Now().Local().Format("2006-01-02")
		if task.DueOn == "" || task.DueOn >= today {
			reasons = append(reasons, "task is not overdue")
		}
	}

	if strings.TrimSpace(filter.Parent) != "" {
		parent := strings.TrimSpace(filter.Parent)
		if strings.EqualFold(parent, "root") {
			parent = ""
		}
		if task.Parent != parent {
			if parent == "" {
				reasons = append(reasons, "task parent is not root")
			} else {
				reasons = append(reasons, fmt.Sprintf("task parent is not %s", parent))
			}
		}
	}

	if search := strings.ToLower(strings.TrimSpace(filter.Search)); search != "" {
		target := strings.ToLower(task.Title + "\n" + task.Body)
		if !strings.Contains(target, search) {
			reasons = append(reasons, fmt.Sprintf("title/body does not match search %q", filter.Search))
		}
	}

	return filterExplain{
		Match:   len(reasons) == 0,
		Reasons: reasons,
	}, nil
}
