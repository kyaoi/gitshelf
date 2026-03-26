package cli

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

var byDaysPattern = regexp.MustCompile(`^(-?\d+)d$`)

const selectorHelpText = "j/k: move  Enter: confirm  /: search  ?: help  q/Esc/Ctrl+C: cancel"

func resolveEditorCommand(lookupEnv func(string) (string, bool)) (string, error) {
	for _, key := range []string{"VISUAL", "EDITOR"} {
		value, ok := lookupEnv(key)
		if ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value), nil
		}
	}
	return "vi", nil
}

func normalizeEditorExecError(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf("editor exited with status %d", exitErr.ExitCode())
	}
	return fmt.Errorf("failed to start editor: %w", err)
}

func selectEnumOption(prompt string, options []interactive.Option) (interactive.Option, error) {
	return interactive.SelectWithConfig(interactive.SelectConfig{
		Prompt:            prompt,
		Options:           options,
		ShowPreview:       false,
		MaxRows:           15,
		HelpText:          selectorHelpText,
		SearchPlaceholder: "Search",
	})
}

func treeOptionsFromFilter(base shelf.TreeOptions, filter shelf.TaskFilter) (shelf.TreeOptions, error) {
	if filter.ReadyOnly || filter.DepsBlocked || filter.DueBefore != "" || filter.DueAfter != "" || filter.Overdue || filter.NoDue || filter.Parent != "" || filter.Search != "" || filter.Limit > 0 {
		return shelf.TreeOptions{}, fmt.Errorf("tree filter contains unsupported fields")
	}
	opts := base
	if len(opts.Kinds) == 0 && len(filter.Kinds) > 0 {
		opts.Kinds = filter.Kinds
	}
	if len(opts.Statuses) == 0 && len(filter.Statuses) > 0 {
		opts.Statuses = filter.Statuses
	}
	if len(opts.Tags) == 0 && len(filter.Tags) > 0 {
		opts.Tags = filter.Tags
	}
	if len(opts.NotKinds) == 0 && len(filter.NotKinds) > 0 {
		opts.NotKinds = filter.NotKinds
	}
	if len(opts.NotStatuses) == 0 && len(filter.NotStatuses) > 0 {
		opts.NotStatuses = filter.NotStatuses
	}
	if len(opts.NotTags) == 0 && len(filter.NotTags) > 0 {
		opts.NotTags = filter.NotTags
	}
	return opts, nil
}

type snoozeMode string

const (
	snoozeModeBy snoozeMode = "by"
	snoozeModeTo snoozeMode = "to"
)

type snoozePreset struct {
	Label      string
	Mode       snoozeMode
	Value      string
	NeedsInput bool
}

func snoozeInteractivePresets() []snoozePreset {
	return []snoozePreset{
		{Label: "Today", Mode: snoozeModeTo, Value: "today"},
		{Label: "Tomorrow", Mode: snoozeModeTo, Value: "tomorrow"},
		{Label: "By +1 day", Mode: snoozeModeBy, Value: "1d"},
		{Label: "By +3 days", Mode: snoozeModeBy, Value: "3d"},
		{Label: "By +7 days", Mode: snoozeModeBy, Value: "7d"},
		{Label: "Next week", Mode: snoozeModeTo, Value: "next-week"},
		{Label: "Next Monday", Mode: snoozeModeTo, Value: "next-mon"},
		{Label: "Custom by days", Mode: snoozeModeBy, NeedsInput: true},
		{Label: "Custom date token", Mode: snoozeModeTo, NeedsInput: true},
	}
}

func applyByDays(currentDue string, by string) (string, error) {
	match := byDaysPattern.FindStringSubmatch(strings.TrimSpace(by))
	if len(match) != 2 {
		return "", fmt.Errorf("invalid --by value: %s (expected <N>d)", by)
	}
	days, err := strconv.Atoi(match[1])
	if err != nil {
		return "", fmt.Errorf("invalid --by value: %s", by)
	}

	base := time.Now().Local()
	if strings.TrimSpace(currentDue) != "" {
		base, err = time.ParseInLocation("2006-01-02", currentDue, time.Local)
		if err != nil {
			return "", fmt.Errorf("invalid existing due_on: %s", currentDue)
		}
	}
	return base.AddDate(0, 0, days).Format("2006-01-02"), nil
}
