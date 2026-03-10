package cli

import (
	"time"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

type dailyCockpitOutputMode string

const (
	dailyCockpitOutputText dailyCockpitOutputMode = "text"
	dailyCockpitOutputJSON dailyCockpitOutputMode = "json"
	dailyCockpitOutputTUI  dailyCockpitOutputMode = "tui"
)

var (
	dailyCockpitIsTTY    = interactive.IsTTY
	runCalendarModeTUIFn = runCalendarModeTUI
)

func resolveReviewOutputMode(isTTY bool, asJSON bool, plain bool) dailyCockpitOutputMode {
	if asJSON {
		return dailyCockpitOutputJSON
	}
	if isTTY && !plain {
		return dailyCockpitOutputTUI
	}
	return dailyCockpitOutputText
}

func resolveTodayOutputMode(isTTY bool, asJSON bool, plain bool, carryOver bool) dailyCockpitOutputMode {
	if asJSON {
		return dailyCockpitOutputJSON
	}
	if carryOver {
		return dailyCockpitOutputText
	}
	if isTTY && !plain {
		return dailyCockpitOutputTUI
	}
	return dailyCockpitOutputText
}

func resolveTreeOutputMode(isTTY bool, asJSON bool, plain bool) dailyCockpitOutputMode {
	if asJSON {
		return dailyCockpitOutputJSON
	}
	if isTTY && !plain {
		return dailyCockpitOutputTUI
	}
	return dailyCockpitOutputText
}

func resolveDailyCockpitRange(rootDir string) (time.Time, int, error) {
	cfg, err := shelf.LoadConfig(rootDir)
	if err != nil {
		return time.Time{}, 0, err
	}
	start := startOfWeek(time.Now().Local())
	return resolveCalendarRange(start, 0, 0, 0, cfg.Commands.Calendar, false, false, false)
}

func activeStatusFilter() []shelf.Status {
	return []shelf.Status{"open", "in_progress", "blocked"}
}

func runDefaultCockpit(ctx *commandContext) error {
	if !dailyCockpitIsTTY() {
		return nil
	}
	cfg, err := shelf.LoadConfig(ctx.rootDir)
	if err != nil {
		return err
	}
	start := startOfWeek(time.Now().Local())
	startDate, dayCount, err := resolveCalendarRange(start, 0, 0, 0, cfg.Commands.Calendar, false, false, false)
	if err != nil {
		return err
	}
	statuses := activeStatusFilter()
	if err := runCalendarModeTUIFn(ctx.rootDir, startDate, dayCount, statuses, calendarTUIOptions{
		Mode:   calendarModeCalendar,
		ShowID: ctx.showID,
		Filter: shelf.TaskFilter{
			Statuses: statuses,
			Limit:    0,
		},
	}); err != nil {
		return err
	}
	settings, err := resolvePostExitGitSettings(ctx, cfg)
	if err != nil {
		return err
	}
	return runPostExitGitAction(ctx.rootDir, settings)
}
