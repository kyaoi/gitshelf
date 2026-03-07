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
