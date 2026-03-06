package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

type agendaBuckets struct {
	Overdue  []shelf.Task `json:"overdue"`
	Today    []shelf.Task `json:"today"`
	Tomorrow []shelf.Task `json:"tomorrow"`
	Upcoming []shelf.Task `json:"upcoming"`
	Later    []shelf.Task `json:"later"`
	NoDue    []shelf.Task `json:"no_due"`
}

func newAgendaCommand(ctx *commandContext) *cobra.Command {
	var (
		view            string
		includeArchived bool
		onlyArchived    bool
		kinds           []string
		statuses        []string
		notKinds        []string
		notStatuses     []string
		days            int
		asJSON          bool
	)

	cmd := &cobra.Command{
		Use:   "agenda",
		Short: "Show due-oriented task agenda",
		Example: "  shelf agenda\n" +
			"  shelf agenda --days 14\n" +
			"  shelf agenda --view active --json",
		RunE: func(cmd *cobra.Command, _ []string) error {
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

			buckets := buildAgendaBuckets(tasks, days)
			if asJSON {
				data, err := json.MarshalIndent(buckets, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			printAgendaBucket := func(title string, rows []shelf.Task) {
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
					dueText := uiMuted("(none)")
					if task.DueOn != "" {
						dueText = uiDue(task.DueOn)
					}
					fmt.Printf("  %s (%s/%s) due=%s\n", label, uiKind(task.Kind), uiStatus(task.Status), dueText)
				}
			}

			printAgendaBucket("Overdue", buckets.Overdue)
			printAgendaBucket("Today", buckets.Today)
			printAgendaBucket("Tomorrow", buckets.Tomorrow)
			printAgendaBucket("Upcoming", buckets.Upcoming)
			printAgendaBucket("Later", buckets.Later)
			printAgendaBucket("No due", buckets.NoDue)
			return nil
		},
	}

	cmd.Flags().StringVar(&view, "view", "", "Apply built-in or config view")
	cmd.Flags().BoolVar(&includeArchived, "include-archived", false, "Include archived tasks")
	cmd.Flags().BoolVar(&onlyArchived, "only-archived", false, "Include only archived tasks")
	cmd.Flags().StringArrayVar(&kinds, "kind", nil, "Include kind (repeatable)")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Include status (repeatable)")
	cmd.Flags().StringArrayVar(&notKinds, "not-kind", nil, "Exclude kind (repeatable)")
	cmd.Flags().StringArrayVar(&notStatuses, "not-status", nil, "Exclude status (repeatable)")
	cmd.Flags().IntVar(&days, "days", 7, "Upcoming range in days")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func buildAgendaBuckets(tasks []shelf.Task, days int) agendaBuckets {
	today := time.Now().Local().Format("2006-01-02")
	tomorrow := time.Now().Local().AddDate(0, 0, 1).Format("2006-01-02")
	upcomingLimit := time.Now().Local().AddDate(0, 0, days).Format("2006-01-02")

	result := agendaBuckets{}
	for _, task := range tasks {
		switch {
		case task.DueOn == "":
			result.NoDue = append(result.NoDue, task)
		case task.DueOn < today:
			result.Overdue = append(result.Overdue, task)
		case task.DueOn == today:
			result.Today = append(result.Today, task)
		case task.DueOn == tomorrow:
			result.Tomorrow = append(result.Tomorrow, task)
		case task.DueOn <= upcomingLimit:
			result.Upcoming = append(result.Upcoming, task)
		default:
			result.Later = append(result.Later, task)
		}
	}

	sortByDueThenID := func(rows []shelf.Task) {
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].DueOn != rows[j].DueOn {
				return rows[i].DueOn < rows[j].DueOn
			}
			return rows[i].ID < rows[j].ID
		})
	}
	sortByDueThenID(result.Overdue)
	sortByDueThenID(result.Today)
	sortByDueThenID(result.Tomorrow)
	sortByDueThenID(result.Upcoming)
	sortByDueThenID(result.Later)
	sort.Slice(result.NoDue, func(i, j int) bool {
		return result.NoDue[i].ID < result.NoDue[j].ID
	})

	return result
}
