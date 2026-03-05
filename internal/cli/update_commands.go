package cli

import (
	"errors"
	"fmt"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newSetCommand(ctx *commandContext) *cobra.Command {
	var (
		title      string
		kind       string
		status     string
		parent     string
		body       string
		appendBody string
	)

	cmd := &cobra.Command{
		Use:   "set <id>",
		Short: "Update task fields",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "更新するタスクを選択", nil)
			if err != nil {
				return err
			}

			input := shelf.SetTaskInput{}
			if cmd.Flags().Changed("title") {
				input.Title = &title
			}
			if cmd.Flags().Changed("kind") {
				k := shelf.Kind(kind)
				input.Kind = &k
			}
			if cmd.Flags().Changed("status") {
				s := shelf.Status(status)
				input.Status = &s
			}
			if cmd.Flags().Changed("parent") {
				input.Parent = &parent
			}
			if cmd.Flags().Changed("body") {
				input.Body = &body
			}
			if cmd.Flags().Changed("append-body") {
				input.AppendBody = &appendBody
			}

			if input.Title == nil && input.Kind == nil && input.Status == nil && input.Parent == nil && input.Body == nil && input.AppendBody == nil {
				if interactive.IsTTY() {
					cfg, err := shelf.LoadConfig(ctx.rootDir)
					if err != nil {
						return err
					}
					statusOptions := make([]interactive.Option, 0, len(cfg.Statuses))
					for _, s := range cfg.Statuses {
						statusOptions = append(statusOptions, interactive.Option{
							Value:      string(s),
							Label:      string(s),
							SearchText: string(s),
						})
					}
					selected, err := interactive.Select("Status を選択してください", statusOptions)
					if err != nil {
						return err
					}
					s := shelf.Status(selected.Value)
					input.Status = &s
				} else {
					return errors.New("更新対象がありません。--title/--kind/--status/--parent/--body/--append-body を指定してください")
				}
			}

			task, err := shelf.SetTask(ctx.rootDir, id, input)
			if err != nil {
				return err
			}
			fmt.Printf("Updated: [%s] %s\n", shelf.ShortID(task.ID), task.Title)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&kind, "kind", "", "New kind")
	cmd.Flags().StringVar(&status, "status", "", "New status")
	cmd.Flags().StringVar(&parent, "parent", "", "New parent task ID or root")
	cmd.Flags().StringVar(&body, "body", "", "Replace body")
	cmd.Flags().StringVar(&appendBody, "append-body", "", "Append text to body")
	return cmd
}

func newMvCommand(ctx *commandContext) *cobra.Command {
	var parent string
	cmd := &cobra.Command{
		Use:   "mv <id>",
		Short: "Move task under a new parent",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "移動するタスクを選択", nil)
			if err != nil {
				return err
			}

			resolvedParent := parent
			if !cmd.Flags().Changed("parent") {
				resolvedParent, err = selectParentIfMissing(ctx, id, parent)
				if err != nil {
					return err
				}
			}

			task, err := shelf.SetTask(ctx.rootDir, id, shelf.SetTaskInput{
				Parent: &resolvedParent,
			})
			if err != nil {
				return err
			}
			parentLabel := "root"
			if task.Parent != "" {
				parentLabel = shelf.ShortID(task.Parent)
			}
			fmt.Printf("Moved: [%s] parent=%s\n", shelf.ShortID(task.ID), parentLabel)
			return nil
		},
	}

	cmd.Flags().StringVar(&parent, "parent", "", "New parent task ID or root")
	return cmd
}

func newDoneCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done <id>",
		Short: "Shortcut to set --status done",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "done にするタスクを選択", func(task shelf.Task) bool {
				return task.Status != shelf.Status("done")
			})
			if err != nil {
				return err
			}

			done := shelf.Status("done")
			task, err := shelf.SetTask(ctx.rootDir, id, shelf.SetTaskInput{
				Status: &done,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Done: [%s] %s\n", shelf.ShortID(task.ID), task.Title)
			return nil
		},
	}
	return cmd
}
