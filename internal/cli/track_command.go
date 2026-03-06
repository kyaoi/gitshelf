package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newTrackCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "track",
		Short: "Start, stop, and inspect task timers",
	}
	cmd.AddCommand(newTrackStartCommand(ctx))
	cmd.AddCommand(newTrackStopCommand(ctx))
	cmd.AddCommand(newTrackShowCommand(ctx))
	return cmd
}

func newTrackStartCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start <id>",
		Short: "Start a timer for a task",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "タイマー開始対象のタスクを選択", nil, true)
			if err != nil {
				return err
			}
			task, err := shelf.EnsureTaskExists(ctx.rootDir, id)
			if err != nil {
				return err
			}
			if strings.TrimSpace(task.TimerStart) != "" {
				return fmt.Errorf("timer is already running")
			}
			startedAt := time.Now().Local().Round(time.Second).Format(time.RFC3339)
			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "track-start"); err != nil {
					return err
				}
				updated, err := shelf.SetTask(ctx.rootDir, id, shelf.SetTaskInput{TimerStart: &startedAt})
				if err != nil {
					return err
				}
				fmt.Printf("Timer started: [%s] %s\n", shelf.ShortID(updated.ID), updated.Title)
				return nil
			})
		},
	}
	return cmd
}

func newTrackStopCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop <id>",
		Short: "Stop a timer and add elapsed time to spent",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "タイマー停止対象のタスクを選択", nil, true)
			if err != nil {
				return err
			}
			task, err := shelf.EnsureTaskExists(ctx.rootDir, id)
			if err != nil {
				return err
			}
			elapsed, err := shelf.ElapsedMinutesSince(task.TimerStart, time.Now().Local())
			if err != nil {
				return err
			}
			total := task.SpentMin + elapsed
			empty := ""
			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "track-stop"); err != nil {
					return err
				}
				updated, err := shelf.SetTask(ctx.rootDir, id, shelf.SetTaskInput{
					SpentMin:   &total,
					TimerStart: &empty,
				})
				if err != nil {
					return err
				}
				fmt.Printf("Timer stopped: [%s] +%s total=%s\n",
					shelf.ShortID(updated.ID),
					shelf.FormatWorkMinutes(elapsed),
					shelf.FormatWorkMinutes(updated.SpentMin),
				)
				return nil
			})
		},
	}
	return cmd
}

func newTrackShowCommand(ctx *commandContext) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "show [id]",
		Short: "Show running timers",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
				task, err := shelf.EnsureTaskExists(ctx.rootDir, args[0])
				if err != nil {
					return err
				}
				return printTrackStatus([]shelf.Task{task}, asJSON)
			}
			tasks, err := shelf.NewTaskStore(ctx.rootDir).List()
			if err != nil {
				return err
			}
			running := make([]shelf.Task, 0)
			for _, task := range tasks {
				if strings.TrimSpace(task.TimerStart) != "" {
					running = append(running, task)
				}
			}
			return printTrackStatus(running, asJSON)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func printTrackStatus(tasks []shelf.Task, asJSON bool) error {
	type row struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Spent     int    `json:"spent"`
		TimerOpen bool   `json:"timer_open"`
		StartedAt string `json:"started_at,omitempty"`
	}
	rows := make([]row, 0, len(tasks))
	for _, task := range tasks {
		rows = append(rows, row{
			ID:        task.ID,
			Title:     task.Title,
			Spent:     task.SpentMin,
			TimerOpen: strings.TrimSpace(task.TimerStart) != "",
			StartedAt: task.TimerStart,
		})
	}
	if asJSON {
		data, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	if len(rows) == 0 {
		fmt.Println("(none)")
		return nil
	}
	for _, item := range rows {
		fmt.Printf("[%s] %s spent=%s running=%t\n",
			shelf.ShortID(item.ID),
			item.Title,
			shelf.FormatWorkMinutes(item.Spent),
			item.TimerOpen,
		)
	}
	return nil
}
