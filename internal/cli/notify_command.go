package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newNotifyCommand(ctx *commandContext) *cobra.Command {
	var (
		command string
		dryRun  bool
	)

	cmd := &cobra.Command{
		Use:   "notify",
		Short: "Run a local notification command for due tasks",
		Example: "  shelf notify --command 'notify-send gitshelf \"$SHELF_TASK_TITLE\"'\n" +
			"  shelf notify --dry-run",
		RunE: func(_ *cobra.Command, _ []string) error {
			tasks, err := shelf.ListTasks(ctx.rootDir, shelf.TaskFilter{
				Statuses: []shelf.Status{"open", "in_progress", "blocked"},
				Limit:    0,
			})
			if err != nil {
				return err
			}
			today := time.Now().Local().Format("2006-01-02")
			dueTasks := make([]shelf.Task, 0)
			for _, task := range tasks {
				if strings.TrimSpace(task.DueOn) == "" || task.DueOn > today {
					continue
				}
				dueTasks = append(dueTasks, task)
			}
			if len(dueTasks) == 0 {
				fmt.Println("No due tasks to notify.")
				return nil
			}

			resolvedCommand, err := resolveNotifyCommand(command)
			if err != nil {
				return err
			}
			if dryRun {
				fmt.Printf("Would notify %d task(s) with: %s\n", len(dueTasks), resolvedCommand)
				return nil
			}

			sent := 0
			for _, task := range dueTasks {
				if err := runNotifyCommand(resolvedCommand, task); err != nil {
					return err
				}
				sent++
			}
			fmt.Printf("Sent notifications: %d\n", sent)
			return nil
		},
	}

	cmd.Flags().StringVar(&command, "command", "", "Shell command used for notifications")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show notification plan without running commands")
	return cmd
}

func resolveNotifyCommand(command string) (string, error) {
	if strings.TrimSpace(command) != "" {
		return strings.TrimSpace(command), nil
	}
	if env := strings.TrimSpace(os.Getenv("SHELF_NOTIFY_COMMAND")); env != "" {
		return env, nil
	}
	if path, err := exec.LookPath("notify-send"); err == nil {
		return path + " gitshelf \"$SHELF_TASK_TITLE [$SHELF_TASK_DUE_ON]\"", nil
	}
	return "", fmt.Errorf("notification command is required (--command or SHELF_NOTIFY_COMMAND)")
}

func runNotifyCommand(command string, task shelf.Task) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Env = append(os.Environ(),
		"SHELF_TASK_ID="+task.ID,
		"SHELF_TASK_SHORT_ID="+shelf.ShortID(task.ID),
		"SHELF_TASK_TITLE="+task.Title,
		"SHELF_TASK_KIND="+string(task.Kind),
		"SHELF_TASK_STATUS="+string(task.Status),
		"SHELF_TASK_DUE_ON="+task.DueOn,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
