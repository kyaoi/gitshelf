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
		years    int
		statuses []string
		asJSON   bool
	)

	cmd := &cobra.Command{
		Use:     "calendar",
		Aliases: []string{"cal"},
		Short:   "Open Focus in calendar mode",
		Long: "Open Focus in calendar mode.\n\n" +
			"If no explicit range flag is set, config [commands.calendar]\n" +
			"default_range_unit/default_days/default_months/default_years decides the range.",
		Example: "  shelf calendar\n" +
			"  shelf calendar --months 3\n" +
			"  shelf calendar --start 2026-03-09 --days 14\n" +
			"  shelf calendar --status open --status blocked --json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			startDate, err := resolveCalendarStart(start)
			if err != nil {
				return err
			}
			rangeStart, dayCount, err := resolveCalendarRange(startDate, days, months, years, cfg.Commands.Calendar, cmd.Flags().Changed("days"), cmd.Flags().Changed("months"), cmd.Flags().Changed("years"))
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

	cmd.Flags().StringVar(&start, "start", "", "Anchor date (YYYY-MM-DD|today|tomorrow). Defaults to current week Monday")
	cmd.Flags().IntVar(&days, "days", 0, "Render an explicit day range")
	cmd.Flags().IntVar(&months, "months", 0, "Render an explicit month range from the month containing --start")
	cmd.Flags().IntVar(&years, "years", 0, "Render an explicit year range from the year containing --start")
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

func resolveCalendarRange(startDate time.Time, days int, months int, years int, cfg shelf.CalendarCommandConfig, daysChanged bool, monthsChanged bool, yearsChanged bool) (time.Time, int, error) {
	changedCount := 0
	for _, changed := range []bool{daysChanged, monthsChanged, yearsChanged} {
		if changed {
			changedCount++
		}
	}
	if changedCount > 1 {
		return time.Time{}, 0, fmt.Errorf("--days / --months / --years はどれか1つだけ指定してください")
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
	if yearsChanged {
		if years <= 0 {
			return time.Time{}, 0, fmt.Errorf("--years must be > 0")
		}
		yearStart := time.Date(startDate.Year(), time.January, 1, 0, 0, 0, 0, startDate.Location())
		yearEndExclusive := yearStart.AddDate(years, 0, 0)
		dayCount := int(yearEndExclusive.Sub(yearStart).Hours() / 24)
		return yearStart, dayCount, nil
	}
	if !daysChanged && !monthsChanged && !yearsChanged {
		switch cfg.DefaultRangeUnit {
		case "months":
			months = cfg.DefaultMonths
			monthsChanged = true
		case "years":
			years = cfg.DefaultYears
			yearsChanged = true
		default:
			days = cfg.DefaultDays
			daysChanged = true
		}
		return resolveCalendarRange(startDate, days, months, years, cfg, daysChanged, monthsChanged, yearsChanged)
	}
	if days <= 0 {
		return time.Time{}, 0, fmt.Errorf("--days must be > 0")
	}
	return startDate, days, nil
}
