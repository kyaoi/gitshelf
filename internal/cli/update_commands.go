package cli

import (
	"errors"
	"fmt"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newSetCommand(ctx *commandContext) *cobra.Command {
	var (
		title      string
		kind       string
		state      string
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
			if cmd.Flags().Changed("state") {
				s := shelf.State(state)
				input.State = &s
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

			if input.Title == nil && input.Kind == nil && input.State == nil && input.Parent == nil && input.Body == nil && input.AppendBody == nil {
				return errors.New("更新対象がありません。--title/--kind/--state/--parent/--body/--append-body を指定してください")
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
	cmd.Flags().StringVar(&state, "state", "", "New state")
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
		Short: "Shortcut to set --state done",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "done にするタスクを選択", func(task shelf.Task) bool {
				return task.State != shelf.State("done")
			})
			if err != nil {
				return err
			}

			done := shelf.State("done")
			task, err := shelf.SetTask(ctx.rootDir, id, shelf.SetTaskInput{
				State: &done,
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
