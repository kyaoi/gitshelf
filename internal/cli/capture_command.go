package cli

import (
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newCaptureCommand(ctx *commandContext) *cobra.Command {
	var (
		title string
		body  string
		tags  []string
		due   string
	)

	cmd := &cobra.Command{
		Use:   "capture [title...]",
		Short: "Quickly capture a task into inbox/open",
		Example: "  shelf capture \"Investigate incident\"\n" +
			"  shelf capture --title \"Write idea\" --tag idea --body \"raw memo\"",
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if c.Flags().Changed("title") && len(args) > 0 {
				return fmt.Errorf("--title と引数titleは同時に指定できません")
			}

			resolvedTitle := strings.TrimSpace(title)
			if resolvedTitle == "" {
				resolvedTitle = strings.TrimSpace(strings.Join(args, " "))
			}
			if resolvedTitle == "" {
				if !interactive.IsTTY() {
					return fmt.Errorf("<title> を指定してください")
				}
				inputTitle, err := interactive.PromptText("Title を入力してください")
				if err != nil {
					return err
				}
				resolvedTitle = strings.TrimSpace(inputTitle)
			}
			if resolvedTitle == "" {
				return fmt.Errorf("title は必須です")
			}

			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			if err := cfg.ValidateKind("inbox"); err != nil {
				return fmt.Errorf("capture requires kind \"inbox\": %w", err)
			}
			if err := cfg.ValidateStatus("open"); err != nil {
				return fmt.Errorf("capture requires status \"open\": %w", err)
			}

			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "capture"); err != nil {
					return err
				}
				task, err := shelf.AddTask(ctx.rootDir, shelf.AddTaskInput{
					Title:  resolvedTitle,
					Kind:   "inbox",
					Status: "open",
					Tags:   parseTagFlagValues(tags),
					DueOn:  strings.TrimSpace(due),
					Body:   body,
				})
				if err != nil {
					return err
				}
				fmt.Printf("Captured: [%s] %s\n", shelf.ShortID(task.ID), task.Title)
				fmt.Printf("ID: %s\n", task.ID)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Task title")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Task tag (repeatable)")
	cmd.Flags().StringVar(&due, "due", "", "Task due date (YYYY-MM-DD|today|tomorrow|+Nd|-Nd|next-week|this-week|mon..sun|next-mon..next-sun|in N days)")
	cmd.Flags().StringVar(&body, "body", "", "Task body")
	return cmd
}
