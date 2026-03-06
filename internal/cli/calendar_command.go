package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

type calendarDay struct {
	Date  string       `json:"date"`
	Tasks []shelf.Task `json:"tasks"`
}

func newCalendarCommand(ctx *commandContext) *cobra.Command {
	var (
		start    string
		days     int
		statuses []string
		asJSON   bool
	)

	cmd := &cobra.Command{
		Use:   "calendar",
		Short: "Show due tasks in a weekly calendar view",
		Example: "  shelf calendar\n" +
			"  shelf calendar --start 2026-03-09\n" +
			"  shelf calendar --status open --status blocked --json",
		RunE: func(_ *cobra.Command, _ []string) error {
			startDate, err := resolveCalendarStart(start)
			if err != nil {
				return err
			}
			if days <= 0 {
				return fmt.Errorf("--days must be > 0")
			}
			selectedStatuses := toStatuses(statuses)
			if len(selectedStatuses) == 0 {
				selectedStatuses = []shelf.Status{"open", "in_progress", "blocked"}
			}

			tasks, err := shelf.ListTasks(ctx.rootDir, shelf.TaskFilter{
				Statuses: selectedStatuses,
				Limit:    0,
			})
			if err != nil {
				return err
			}

			calendar := buildCalendarDays(tasks, startDate, days)
			if asJSON {
				data, err := json.MarshalIndent(calendar, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}
			if days > 7 {
				if !interactive.IsTTY() {
					return errors.New("calendar の 8 日以上表示はTTYが必要です。--json を使うか --days を 7 以下にしてください")
				}
				return runCalendarTUI(calendar, ctx.showID)
			}

			fmt.Printf("Week of %s\n", startDate.Format("2006-01-02"))
			for _, day := range calendar {
				parsed, _ := time.Parse("2006-01-02", day.Date)
				fmt.Println(uiHeading(parsed.Format("Mon 2006-01-02")))
				if len(day.Tasks) == 0 {
					fmt.Println(uiMuted("  (none)"))
					continue
				}
				for _, task := range day.Tasks {
					label := uiPrimary(task.Title)
					if ctx.showID {
						label = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(task.ID)), label)
					}
					fmt.Printf("  - %s (%s/%s)\n", label, uiKind(task.Kind), uiStatus(task.Status))
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&start, "start", "", "Week start date (YYYY-MM-DD|today|tomorrow). Defaults to current week Monday")
	cmd.Flags().IntVar(&days, "days", 7, "Number of days to render")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Include status (repeatable)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func resolveCalendarStart(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return startOfWeek(time.Now().Local()), nil
	}
	normalized, err := shelf.NormalizeDueOn(value)
	if err != nil {
		return time.Time{}, err
	}
	parsed, err := time.ParseInLocation("2006-01-02", normalized, time.Now().Location())
	if err != nil {
		return time.Time{}, err
	}
	return startOfWeek(parsed), nil
}

func startOfWeek(value time.Time) time.Time {
	day := value
	for day.Weekday() != time.Monday {
		day = day.AddDate(0, 0, -1)
	}
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
}

func buildCalendarDays(tasks []shelf.Task, startDate time.Time, days int) []calendarDay {
	grouped := make(map[string][]shelf.Task, days)
	for _, task := range tasks {
		if strings.TrimSpace(task.DueOn) == "" {
			continue
		}
		grouped[task.DueOn] = append(grouped[task.DueOn], task)
	}
	rows := make([]calendarDay, 0, days)
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i).Format("2006-01-02")
		rows = append(rows, calendarDay{
			Date:  date,
			Tasks: grouped[date],
		})
	}
	return rows
}
