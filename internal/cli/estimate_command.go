package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newEstimateCommand(ctx *commandContext) *cobra.Command {
	var (
		setEstimate   string
		setSpent      string
		addSpent      string
		clearEstimate bool
		clearSpent    bool
		asJSON        bool
	)

	cmd := &cobra.Command{
		Use:   "estimate <id>",
		Short: "Show or update estimate/spent work for a task",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "見積もり対象のタスクを選択", nil, true)
			if err != nil {
				return err
			}
			task, err := shelf.EnsureTaskExists(ctx.rootDir, id)
			if err != nil {
				return err
			}

			hasUpdate := clearEstimate || clearSpent || strings.TrimSpace(setEstimate) != "" || strings.TrimSpace(setSpent) != "" || strings.TrimSpace(addSpent) != ""
			if !hasUpdate {
				return printEstimate(task, asJSON)
			}

			input := shelf.SetTaskInput{}
			if clearEstimate {
				zero := 0
				input.EstimateMin = &zero
			}
			if clearSpent {
				zero := 0
				input.SpentMin = &zero
			}
			if strings.TrimSpace(setEstimate) != "" {
				mins, err := shelf.ParseWorkDurationMinutes(setEstimate)
				if err != nil {
					return err
				}
				input.EstimateMin = &mins
			}
			if strings.TrimSpace(setSpent) != "" {
				mins, err := shelf.ParseWorkDurationMinutes(setSpent)
				if err != nil {
					return err
				}
				input.SpentMin = &mins
			}
			if strings.TrimSpace(addSpent) != "" {
				mins, err := shelf.ParseWorkDurationMinutes(addSpent)
				if err != nil {
					return err
				}
				total := task.SpentMin + mins
				input.SpentMin = &total
			}

			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "estimate"); err != nil {
					return err
				}
				updated, err := shelf.SetTask(ctx.rootDir, id, input)
				if err != nil {
					return err
				}
				return printEstimate(updated, asJSON)
			})
		},
	}

	cmd.Flags().StringVar(&setEstimate, "set", "", "Set estimate duration (e.g. 2h30m)")
	cmd.Flags().StringVar(&setSpent, "spent", "", "Set spent duration (e.g. 45m)")
	cmd.Flags().StringVar(&addSpent, "add-spent", "", "Add spent duration")
	cmd.Flags().BoolVar(&clearEstimate, "clear-estimate", false, "Clear estimate")
	cmd.Flags().BoolVar(&clearSpent, "clear-spent", false, "Clear spent")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func printEstimate(task shelf.Task, asJSON bool) error {
	remaining := task.EstimateMin - task.SpentMin
	if remaining < 0 {
		remaining = 0
	}
	if asJSON {
		payload := map[string]any{
			"id":         task.ID,
			"title":      task.Title,
			"estimate":   task.EstimateMin,
			"spent":      task.SpentMin,
			"remaining":  remaining,
			"timer_open": strings.TrimSpace(task.TimerStart) != "",
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	fmt.Printf("%s\n", task.Title)
	fmt.Printf("estimate=%s spent=%s remaining=%s timer=%t\n",
		shelf.FormatWorkMinutes(task.EstimateMin),
		shelf.FormatWorkMinutes(task.SpentMin),
		shelf.FormatWorkMinutes(remaining),
		strings.TrimSpace(task.TimerStart) != "",
	)
	return nil
}
