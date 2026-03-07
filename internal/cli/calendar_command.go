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
		months   int
		statuses []string
		asJSON   bool
	)

	cmd := &cobra.Command{
		Use:   "calendar",
		Short: "Show due tasks in a calendar TUI",
		Example: "  shelf calendar\n" +
			"  shelf calendar --start 2026-03-09\n" +
			"  shelf calendar --status open --status blocked --json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			startDate, err := resolveCalendarStart(start)
			if err != nil {
				return err
			}
			rangeStart, dayCount, err := resolveCalendarRange(startDate, days, months, cmd.Flags().Changed("days"), cmd.Flags().Changed("months"))
			if err != nil {
				return err
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

			calendar := buildCalendarDays(tasks, rangeStart, dayCount)
			if asJSON {
				data, err := json.MarshalIndent(calendar, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}
			if !interactive.IsTTY() {
				return errors.New("calendar はTTYが必要です。非TTYでは --json を使ってください")
			}
			return runCalendarTUI(ctx.rootDir, rangeStart, dayCount, selectedStatuses, ctx.showID)
		},
	}

	cmd.Flags().StringVar(&start, "start", "", "Week start date (YYYY-MM-DD|today|tomorrow). Defaults to current week Monday")
	cmd.Flags().IntVar(&days, "days", 7, "Number of days to render")
	cmd.Flags().IntVar(&months, "months", 0, "Number of whole months to render from the month containing --start")
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

func resolveCalendarRange(startDate time.Time, days int, months int, daysChanged bool, monthsChanged bool) (time.Time, int, error) {
	if daysChanged && monthsChanged {
		return time.Time{}, 0, fmt.Errorf("--days と --months は同時に指定できません")
	}
	if monthsChanged {
		if months <= 0 {
			return time.Time{}, 0, fmt.Errorf("--months must be > 0")
		}
		monthStart := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
		monthEndExclusive := monthStart.AddDate(0, months, 0)
		dayCount := int(monthEndExclusive.Sub(monthStart).Hours() / 24)
		return monthStart, dayCount, nil
	}
	if days <= 0 {
		return time.Time{}, 0, fmt.Errorf("--days must be > 0")
	}
	return startDate, days, nil
}
