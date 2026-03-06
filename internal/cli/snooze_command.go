package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

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
			if byChanged == toChanged {
				return fmt.Errorf("--by か --to のどちらか一方を指定してください")
			}

			task, err := shelf.EnsureTaskExists(ctx.rootDir, id)
			if err != nil {
				return err
			}

			var nextDue string
			if toChanged {
				nextDue, err = shelf.NormalizeDueOn(to)
				if err != nil {
					return err
				}
			} else {
				nextDue, err = applyByDays(task.DueOn, by)
				if err != nil {
					return err
				}
			}

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
		},
	}

	cmd.Flags().StringVar(&by, "by", "", "Shift due date by relative days (e.g. 2d, -1d)")
	cmd.Flags().StringVar(&to, "to", "", "Set due date directly (YYYY-MM-DD|today|tomorrow)")
	return cmd
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
