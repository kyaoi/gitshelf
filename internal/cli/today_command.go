package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newTodayCommand(ctx *commandContext) *cobra.Command {
	var (
		presetName      string
		view            string
		includeArchived bool
		onlyArchived    bool
		format          string
		plain           bool
		kinds           []string
		statuses        []string
		notKinds        []string
		notStatuses     []string
		carryOver       bool
		yes             bool
		asJSON          bool
	)

	cmd := &cobra.Command{
		Use:   "today",
		Short: "Show overdue and today tasks",
		Example: "  shelf today\n" +
			"  shelf today --view active\n" +
			"  shelf today --plain\n" +
			"  shelf today --carry-over --yes\n" +
			"  shelf today --json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			outputPreset, err := loadOutputPreset(ctx.rootDir, presetName, "today")
			if err != nil {
				return err
			}
			view = applyPresetString(view, cmd.Flags().Changed("view"), outputPreset.View)
			format = applyPresetString(format, cmd.Flags().Changed("format"), outputPreset.Format)

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
			if resolveTodayOutputMode(dailyCockpitIsTTY(), asJSON, plain, carryOver) == dailyCockpitOutputTUI {
				startDate, dayCount, err := resolveDailyCockpitRange(ctx.rootDir)
				if err != nil {
					return err
				}
				return runCalendarModeTUIFn(ctx.rootDir, startDate, dayCount, filter.Statuses, calendarTUIOptions{
					Mode:   calendarModeToday,
					ShowID: ctx.showID,
					Filter: filter,
				})
			}
			tasks, err := shelf.ListTasks(ctx.rootDir, filter)
			if err != nil {
				return err
			}

			buckets := buildAgendaBuckets(tasks, 0)
			carriedCount := 0
			if carryOver {
				today := time.Now().Local().Format("2006-01-02")
				targets := make([]shelf.Task, 0, len(buckets.Overdue))
				for _, task := range buckets.Overdue {
					if isCarryOverStatus(task.Status) {
						targets = append(targets, task)
					}
				}
				if len(targets) > 0 {
					if !yes {
						if !interactive.IsTTY() {
							return fmt.Errorf("non-TTY で --carry-over を使う場合は --yes が必要です")
						}
						confirm, err := selectEnumOption("期限切れタスクを今日に繰り上げますか？", []interactive.Option{
							{Value: "apply", Label: fmt.Sprintf("Apply (%d tasks)", len(targets))},
							{Value: "cancel", Label: "Cancel"},
						})
						if err != nil {
							return err
						}
						if confirm.Value != "apply" {
							return interactive.ErrCanceled
						}
					}

					if err := withWriteLock(ctx.rootDir, func() error {
						if err := prepareUndoSnapshot(ctx.rootDir, "today-carry-over"); err != nil {
							return err
						}
						for _, task := range targets {
							due := today
							if _, err := shelf.SetTask(ctx.rootDir, task.ID, shelf.SetTaskInput{
								DueOn: &due,
							}); err != nil {
								return err
							}
						}
						return nil
					}); err != nil {
						return err
					}
					carriedCount = len(targets)
					tasks, err = shelf.ListTasks(ctx.rootDir, filter)
					if err != nil {
						return err
					}
					buckets = buildAgendaBuckets(tasks, 0)
				}
			}
			if asJSON {
				payload := map[string]any{
					"overdue":            buckets.Overdue,
					"today":              buckets.Today,
					"carried_over_count": carriedCount,
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
			if carryOver {
				fmt.Printf("carried_over=%d\n", carriedCount)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&presetName, "preset", "", "Apply output preset for today")
	cmd.Flags().StringVar(&view, "view", "", "Apply built-in or config view")
	cmd.Flags().BoolVar(&includeArchived, "include-archived", false, "Include archived tasks")
	cmd.Flags().BoolVar(&onlyArchived, "only-archived", false, "Include only archived tasks")
	cmd.Flags().StringVar(&format, "format", "compact", "Output format: compact|detail")
	cmd.Flags().BoolVar(&plain, "plain", false, "Force plain text output even on TTY")
	cmd.Flags().StringArrayVar(&kinds, "kind", nil, "Include kind (repeatable)")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Include status (repeatable)")
	cmd.Flags().StringArrayVar(&notKinds, "not-kind", nil, "Exclude kind (repeatable)")
	cmd.Flags().StringArrayVar(&notStatuses, "not-status", nil, "Exclude status (repeatable)")
	cmd.Flags().BoolVar(&carryOver, "carry-over", false, "Move overdue active tasks due date to today")
	cmd.Flags().BoolVar(&yes, "yes", false, "Apply carry-over without confirmation")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func isCarryOverStatus(status shelf.Status) bool {
	return status == "open" || status == "in_progress" || status == "blocked"
}
