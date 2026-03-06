package cli

import (
	"encoding/json"
	"fmt"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newTodayCommand(ctx *commandContext) *cobra.Command {
	var (
		view            string
		includeArchived bool
		onlyArchived    bool
		format          string
		kinds           []string
		statuses        []string
		notKinds        []string
		notStatuses     []string
		asJSON          bool
	)

	cmd := &cobra.Command{
		Use:   "today",
		Short: "Show overdue and today tasks",
		Example: "  shelf today\n" +
			"  shelf today --view active\n" +
			"  shelf today --json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateFormat(format, []string{"compact", "detail"}); err != nil {
				return err
			}
			preset, err := resolveTaskView(ctx.rootDir, view)
			if err != nil {
				return err
			}

			filter := shelf.TaskFilter{
				Kinds:           toKinds(kinds),
				Statuses:        toStatuses(statuses),
				NotKinds:        toKinds(notKinds),
				NotStatuses:     toStatuses(notStatuses),
				IncludeArchived: includeArchived,
				OnlyArchived:    onlyArchived,
				Limit:           0,
			}
			if !cmd.Flags().Changed("status") && len(preset.Statuses) == 0 && len(preset.NotStatuses) == 0 {
				filter.Statuses = []shelf.Status{"open", "in_progress", "blocked"}
			}
			filter = mergeTaskFilterWithView(filter, preset, map[string]bool{
				"limit": true,
			})
			tasks, err := shelf.ListTasks(ctx.rootDir, filter)
			if err != nil {
				return err
			}

			buckets := buildAgendaBuckets(tasks, 0)
			if asJSON {
				payload := map[string]any{
					"overdue": buckets.Overdue,
					"today":   buckets.Today,
				}
				data, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			printRows := func(title string, rows []shelf.Task) {
				fmt.Println(uiHeading(title + ":"))
				if len(rows) == 0 {
					fmt.Println(uiMuted("  (none)"))
					return
				}
				for _, task := range rows {
					label := uiPrimary(task.Title)
					if ctx.showID {
						label = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(task.ID)), uiPrimary(task.Title))
					}
					if format == "detail" {
						fmt.Printf("  %s kind=%s status=%s due=%s repeat=%s\n", label, uiKind(task.Kind), uiStatus(task.Status), uiDue(task.DueOn), task.RepeatEvery)
						continue
					}
					fmt.Printf("  %s (%s/%s) due=%s\n", label, uiKind(task.Kind), uiStatus(task.Status), uiDue(task.DueOn))
				}
			}

			printRows("Overdue", buckets.Overdue)
			printRows("Today", buckets.Today)
			return nil
		},
	}

	cmd.Flags().StringVar(&view, "view", "", "Apply built-in or config view")
	cmd.Flags().BoolVar(&includeArchived, "include-archived", false, "Include archived tasks")
	cmd.Flags().BoolVar(&onlyArchived, "only-archived", false, "Include only archived tasks")
	cmd.Flags().StringVar(&format, "format", "compact", "Output format: compact|detail")
	cmd.Flags().StringArrayVar(&kinds, "kind", nil, "Include kind (repeatable)")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Include status (repeatable)")
	cmd.Flags().StringArrayVar(&notKinds, "not-kind", nil, "Exclude kind (repeatable)")
	cmd.Flags().StringArrayVar(&notStatuses, "not-status", nil, "Exclude status (repeatable)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}
