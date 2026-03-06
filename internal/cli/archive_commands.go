package cli

import (
	"fmt"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newArchiveCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "archive <id>",
		Short:   "Archive a task (set archived_at)",
		Example: "  shelf archive 01ABCDEFG...\n  shelf archive",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "archive するタスクを選択", nil, true)
			if err != nil {
				return err
			}
			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "archive"); err != nil {
					return err
				}
				now := time.Now().Local().Round(time.Second).Format(time.RFC3339)
				task, err := shelf.SetTask(ctx.rootDir, id, shelf.SetTaskInput{
					ArchivedAt: &now,
				})
				if err != nil {
					return err
				}
				fmt.Printf("Archived: [%s] %s\n", shelf.ShortID(task.ID), task.Title)
				return nil
			})
		},
	}
	return cmd
}

func newUnarchiveCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unarchive <id>",
		Short:   "Unarchive a task (clear archived_at)",
		Example: "  shelf unarchive 01ABCDEFG...\n  shelf unarchive",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "unarchive するタスクを選択", nil, true)
			if err != nil {
				return err
			}
			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "unarchive"); err != nil {
					return err
				}
				empty := ""
				task, err := shelf.SetTask(ctx.rootDir, id, shelf.SetTaskInput{
					ArchivedAt: &empty,
				})
				if err != nil {
					return err
				}
				fmt.Printf("Unarchived: [%s] %s\n", shelf.ShortID(task.ID), task.Title)
				return nil
			})
		},
	}
	return cmd
}
