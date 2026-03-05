package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

type commandContext struct {
	rootOverride string
	rootDir      string
}

func NewRootCommand(version string) *cobra.Command {
	ctx := &commandContext{}

	cmd := &cobra.Command{
		Use:           "shelf",
		Short:         "Git-backed lightweight task shelf",
		Long:          "shelf is a lightweight CLI tool for managing tasks and links in .shelf/",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Name() == "init" {
				return nil
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("作業ディレクトリの取得に失敗しました: %w", err)
			}

			rootDir, err := shelf.ResolveShelfRoot(ctx.rootOverride, cwd)
			if err != nil {
				if errors.Is(err, shelf.ErrShelfNotFound) {
					return errors.New(".shelf が見つかりません。先に `shelf init` を実行してください")
				}
				return err
			}
			ctx.rootDir = rootDir
			return nil
		},
	}

	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.PersistentFlags().StringVar(&ctx.rootOverride, "root", "", "Directory that contains .shelf")

	cmd.AddCommand(newInitCommand(ctx))
	cmd.AddCommand(newAddCommand(ctx))
	cmd.AddCommand(newLsCommand(ctx))
	cmd.AddCommand(newTreeCommand(ctx))
	cmd.AddCommand(newShowCommand(ctx))
	cmd.AddCommand(newSetCommand(ctx))
	cmd.AddCommand(newMvCommand(ctx))
	cmd.AddCommand(newDoneCommand(ctx))
	cmd.AddCommand(newLinkCommand(ctx))
	cmd.AddCommand(newUnlinkCommand(ctx))
	cmd.AddCommand(newLinksCommand(ctx))
	cmd.AddCommand(newStubCommand("doctor"))

	return cmd
}

func newInitCommand(ctx *commandContext) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize .shelf in the current directory",
		RunE: func(_ *cobra.Command, _ []string) error {
			targetDir := ctx.rootOverride
			if targetDir == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("作業ディレクトリの取得に失敗しました: %w", err)
				}
				targetDir = cwd
			}

			result, err := shelf.Initialize(targetDir, force)
			if err != nil {
				return err
			}
			switch {
			case result.ConfigForced:
				fmt.Printf("初期化しました: %s (config.toml を再生成)\n", result.ShelfDir)
			case result.ConfigCreated:
				fmt.Printf("初期化しました: %s\n", result.ShelfDir)
			default:
				fmt.Printf("既に初期化済みです: %s\n", result.ShelfDir)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite config.toml with default values")
	return cmd
}

func newStubCommand(name string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("%s command (not implemented yet)", name),
		RunE: func(_ *cobra.Command, _ []string) error {
			return fmt.Errorf("`shelf %s` は未実装です", name)
		},
	}
}

func newAddCommand(ctx *commandContext) *cobra.Command {
	var (
		title  string
		kind   string
		state  string
		parent string
		body   string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new task",
		RunE: func(_ *cobra.Command, _ []string) error {
			var input shelf.AddTaskInput
			if strings.TrimSpace(title) == "" {
				interactiveInput, err := resolveAddInputInteractive(ctx, body, state)
				if err != nil {
					return err
				}
				input = interactiveInput
			} else {
				input = shelf.AddTaskInput{
					Title:  title,
					Kind:   shelf.Kind(kind),
					State:  shelf.State(state),
					Parent: parent,
					Body:   body,
				}
			}

			task, err := shelf.AddTask(ctx.rootDir, input)
			if err != nil {
				return err
			}

			fmt.Printf("Created: [%s] %s\n", shelf.ShortID(task.ID), task.Title)
			fmt.Printf("ID: %s\n", task.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Task title")
	cmd.Flags().StringVar(&kind, "kind", "", "Task kind")
	cmd.Flags().StringVar(&state, "state", "", "Task state")
	cmd.Flags().StringVar(&parent, "parent", "", "Parent task ID or root")
	cmd.Flags().StringVar(&body, "body", "", "Task body")
	return cmd
}
