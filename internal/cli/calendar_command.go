package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

type calendarDay struct {
	Date  string       `json:"date"`
	Tasks []shelf.Task `json:"tasks"`
}

func newCalendarCommand(ctx *commandContext) *cobra.Command {
	var flags cockpitLaunchFlags

	cmd := &cobra.Command{
		Use:     "calendar",
		Aliases: []string{"cal"},
		Short:   "Open Cockpit in calendar mode",
		Long:    "Open Cockpit in calendar mode.",
		Example: "  shelf calendar\n" +
			"  shelf calendar --months 3\n" +
			"  shelf calendar --start 2026-03-09 --days 14\n" +
			"  shelf calendar --status open --status blocked",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !dailyCockpitIsTTY() {
				return errors.New("calendar requires a TTY")
			}
			return runCockpitLaunch(ctx, cmd, calendarModeCalendar, flags)
		},
	}

	addCockpitLaunchFlags(cmd, &flags)
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
	for day.Weekday() != time.Sunday {
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
		return time.Time{}, 0, fmt.Errorf("specify only one of --days / --months / --years")
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
