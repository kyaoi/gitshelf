package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newEditCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [id]",
		Short: "Open a task file in $VISUAL/$EDITOR",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := resolveEditTaskID(ctx, args, interactive.IsTTY())
			if err != nil {
				return err
			}

			taskPath := filepath.Join(shelf.TasksDir(ctx.rootDir), id+".md")
			if _, err := os.Stat(taskPath); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("task が存在しません: %s", id)
				}
				return fmt.Errorf("task file の確認に失敗しました: %w", err)
			}

			editorCmd, err := resolveEditorCommand(os.LookupEnv)
			if err != nil {
				return err
			}
			return runEditorCommand(editorCmd, taskPath, os.Stdin, os.Stdout, os.Stderr)
		},
	}
	return cmd
}

func resolveEditTaskID(ctx *commandContext, args []string, isTTY bool) (string, error) {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return args[0], nil
	}
	if !isTTY {
		return "", errors.New("<id> を指定してください")
	}
	return selectTaskIDIfMissing(ctx, args, "編集するタスクを選択", nil, false)
}

func resolveEditorCommand(lookupEnv func(string) (string, bool)) (string, error) {
	for _, key := range []string{"VISUAL", "EDITOR"} {
		value, ok := lookupEnv(key)
		if ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value), nil
		}
	}
	return "vi", nil
}

func runEditorCommand(command string, taskPath string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	args := strings.Fields(strings.TrimSpace(command))
	if len(args) == 0 {
		return errors.New("editor command is empty")
	}
	args = append(args, taskPath)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("editor exited with status %d", exitErr.ExitCode())
		}
		return fmt.Errorf("failed to start editor %q: %w", command, err)
	}
	return nil
}
