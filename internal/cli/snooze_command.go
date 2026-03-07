package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

var byDaysPattern = regexp.MustCompile(`^(-?\d+)d$`)

func newSnoozeCommand(ctx *commandContext) *cobra.Command {
	var (
		by string
		to string
	)

	cmd := &cobra.Command{
		Use:   "snooze <id>",
		Short: "Adjust due date by relative days or absolute date",
		Example: "  shelf snooze 01ABCDEFG... --by 2d\n" +
			"  shelf snooze 01ABCDEFG... --to tomorrow\n" +
			"  shelf snooze --by -1d",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "期限を調整するタスクを選択", nil, true)
			if err != nil {
				return err
			}

			byChanged := cmd.Flags().Changed("by")
			toChanged := cmd.Flags().Changed("to")
			mode, err := resolveSnoozeMode(byChanged, toChanged, interactive.IsTTY())
			if err != nil {
				return err
			}
			value := ""
			switch mode {
			case snoozeModeBy:
				value = by
			case snoozeModeTo:
				value = to
			default:
				mode, value, err = promptSnoozeInputInteractive()
				if err != nil {
					return err
				}
			}

			task, err := shelf.EnsureTaskExists(ctx.rootDir, id)
			if err != nil {
				return err
			}

			var nextDue string
			if mode == snoozeModeTo {
				nextDue, err = shelf.NormalizeDueOn(value)
				if err != nil {
					return err
				}
			} else {
				nextDue, err = applyByDays(task.DueOn, value)
				if err != nil {
					return err
				}
			}

			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "snooze"); err != nil {
					return err
				}
				updated, err := shelf.SetTask(ctx.rootDir, id, shelf.SetTaskInput{
					DueOn: &nextDue,
				})
				if err != nil {
					return err
				}

				label := uiPrimary(updated.Title)
				if ctx.showID {
					label = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(updated.ID)), label)
				}
				fmt.Printf("Snoozed: %s due=%s\n", label, uiDue(updated.DueOn))
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&by, "by", "", "Shift due date by relative days (e.g. 2d, -1d)")
	cmd.Flags().StringVar(&to, "to", "", "Set due date directly (YYYY-MM-DD|today|tomorrow|+Nd|-Nd|next-week|this-week|mon..sun|next-mon..next-sun|in N days)")
	return cmd
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

func resolveSnoozeMode(byChanged, toChanged bool, interactiveEnabled bool) (snoozeMode, error) {
	switch {
	case byChanged && toChanged:
		return "", fmt.Errorf("--by か --to のどちらか一方を指定してください")
	case byChanged:
		return snoozeModeBy, nil
	case toChanged:
		return snoozeModeTo, nil
	case !interactiveEnabled:
		return "", fmt.Errorf("非TTYでは --by か --to を指定してください")
	default:
		return "", nil
	}
}

func promptSnoozeInputInteractive() (snoozeMode, string, error) {
	options := snoozeInteractiveOptions()
	selected, err := selectEnumOption("期限変更方法を選択してください", options)
	if err != nil {
		return "", "", err
	}
	preset, ok := snoozePresetByLabel(selected.Value)
	if !ok {
		return "", "", fmt.Errorf("unknown snooze preset: %s", selected.Value)
	}
	if !preset.NeedsInput {
		return preset.Mode, preset.Value, nil
	}

	switch preset.Mode {
	case snoozeModeBy:
		value, err := interactive.PromptText("日数を入力してください (<N>d, 例: 2d / -1d)")
		if err != nil {
			return "", "", err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return "", "", fmt.Errorf("--by の値が空です")
		}
		return snoozeModeBy, value, nil
	case snoozeModeTo:
		value, err := interactive.PromptText("新しい期限を入力してください (YYYY-MM-DD, today, tomorrow, +Nd, -Nd, next-week, mon..sun)")
		if err != nil {
			return "", "", err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return "", "", fmt.Errorf("--to の値が空です")
		}
		return snoozeModeTo, value, nil
	default:
		return "", "", fmt.Errorf("unknown snooze mode: %s", selected.Value)
	}
}

func snoozeInteractiveOptions() []interactive.Option {
	presets := snoozeInteractivePresets()
	options := make([]interactive.Option, 0, len(presets))
	for _, preset := range presets {
		search := fmt.Sprintf("%s %s %s", preset.Label, preset.Mode, preset.Value)
		options = append(options, interactive.Option{
			Value:      preset.Label,
			Label:      preset.Label,
			SearchText: search,
		})
	}
	return options
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

func snoozePresetByLabel(label string) (snoozePreset, bool) {
	for _, preset := range snoozeInteractivePresets() {
		if preset.Label == label {
			return preset, true
		}
	}
	return snoozePreset{}, false
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
