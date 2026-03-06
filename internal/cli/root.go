package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyaoi/gitshelf/internal/paths"
	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

type commandContext struct {
	rootOverride string
	rootDir      string
	showID       bool
}

func NewRootCommand(version string) *cobra.Command {
	ctx := &commandContext{}

	cmd := &cobra.Command{
		Use:           "shelf",
		Short:         "Git-backed lightweight task shelf",
		Long:          "shelf is a lightweight CLI tool for managing tasks and links in .shelf/",
		SilenceUsage:  true,
		SilenceErrors: false,
		Version:       version,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Name() == "init" || isCompletionCommand(cmd) {
				return nil
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("作業ディレクトリの取得に失敗しました: %w", err)
			}

			rootDir, err := shelf.ResolveShelfRoot(ctx.rootOverride, cwd)
			if err != nil {
				if errors.Is(err, shelf.ErrShelfNotFound) {
					return errors.New(".shelf が見つかりません。`shelf init` または `shelf init --global` を実行してください")
				}
				return err
			}
			ctx.rootDir = rootDir
			return nil
		},
	}

	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.PersistentFlags().StringVar(&ctx.rootOverride, "root", "", "Directory that contains .shelf")
	cmd.PersistentFlags().BoolVarP(&ctx.showID, "show-id", "i", false, "Show task IDs in list/tree/interactive labels")

	cmd.AddCommand(newInitCommand(ctx))
	cmd.AddCommand(newCompletionCommand())
	cmd.AddCommand(newAddCommand(ctx))
	cmd.AddCommand(newLsCommand(ctx))
	cmd.AddCommand(newViewCommand(ctx))
	cmd.AddCommand(newPresetCommand(ctx))
	cmd.AddCommand(newNextCommand(ctx))
	cmd.AddCommand(newAgendaCommand(ctx))
	cmd.AddCommand(newTodayCommand(ctx))
	cmd.AddCommand(newTreeCommand(ctx))
	cmd.AddCommand(newShowCommand(ctx))
	cmd.AddCommand(newExplainCommand(ctx))
	cmd.AddCommand(newEditCommand(ctx))
	cmd.AddCommand(newSetCommand(ctx))
	cmd.AddCommand(newSnoozeCommand(ctx))
	cmd.AddCommand(newArchiveCommand(ctx))
	cmd.AddCommand(newUnarchiveCommand(ctx))
	cmd.AddCommand(newMvCommand(ctx))
	cmd.AddCommand(newDoneCommand(ctx))
	cmd.AddCommand(newStartCommand(ctx))
	cmd.AddCommand(newBlockCommand(ctx))
	cmd.AddCommand(newCancelCommand(ctx))
	cmd.AddCommand(newReopenCommand(ctx))
	cmd.AddCommand(newLinkCommand(ctx))
	cmd.AddCommand(newUnlinkCommand(ctx))
	cmd.AddCommand(newLinksCommand(ctx))
	cmd.AddCommand(newDepsCommand(ctx))
	cmd.AddCommand(newExportCommand(ctx))
	cmd.AddCommand(newImportCommand(ctx))
	cmd.AddCommand(newUndoCommand(ctx))
	cmd.AddCommand(newRedoCommand(ctx))
	cmd.AddCommand(newHistoryCommand(ctx))
	cmd.AddCommand(newDoctorCommand(ctx))

	return cmd
}

func isCompletionCommand(cmd *cobra.Command) bool {
	for current := cmd; current != nil; current = current.Parent() {
		if current.Name() == "completion" {
			return true
		}
	}
	return false
}

func newInitCommand(ctx *commandContext) *cobra.Command {
	var (
		force  bool
		global bool
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize .shelf in the current directory",
		Example: "  shelf init\n" +
			"  shelf init --root /path/to/project\n" +
			"  shelf init --global",
		RunE: func(_ *cobra.Command, _ []string) error {
			if global {
				return runGlobalInit(ctx.rootOverride, force)
			}

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
	cmd.Flags().BoolVar(&global, "global", false, "Initialize global default root and write global config")
	return cmd
}

func runGlobalInit(rootOverride string, force bool) error {
	globalPath, err := paths.GlobalConfigPath()
	if err != nil {
		return err
	}

	var (
		defaultRoot       string
		existing          paths.GlobalConfig
		hasExistingConfig bool
	)
	if strings.TrimSpace(rootOverride) != "" {
		abs, err := filepath.Abs(rootOverride)
		if err != nil {
			return fmt.Errorf("--root の絶対パス解決に失敗しました: %w", err)
		}
		defaultRoot = abs
	} else {
		cfg, err := paths.LoadGlobalConfig()
		switch {
		case err == nil:
			hasExistingConfig = true
			existing = cfg
			defaultRoot = cfg.DefaultRoot
		case errors.Is(err, paths.ErrGlobalConfigNotFound):
			defaultRoot, err = paths.DefaultGlobalRoot()
			if err != nil {
				return err
			}
		default:
			return err
		}
	}

	shouldSaveGlobal := force || !hasExistingConfig
	if hasExistingConfig && strings.TrimSpace(existing.DefaultRoot) == "" {
		shouldSaveGlobal = true
	}
	if strings.TrimSpace(rootOverride) != "" && (!hasExistingConfig || existing.DefaultRoot != defaultRoot) {
		shouldSaveGlobal = true
	}
	if shouldSaveGlobal {
		if err := paths.SaveGlobalConfig(paths.GlobalConfig{
			DefaultRoot: defaultRoot,
		}); err != nil {
			return err
		}
	}

	result, err := shelf.Initialize(defaultRoot, force)
	if err != nil {
		return err
	}

	fmt.Printf("グローバル設定: %s\n", globalPath)
	fmt.Printf("default_root: %s\n", defaultRoot)
	switch {
	case result.ConfigForced:
		fmt.Printf("初期化しました: %s (config.toml を再生成)\n", result.ShelfDir)
	case result.ConfigCreated:
		fmt.Printf("初期化しました: %s\n", result.ShelfDir)
	default:
		fmt.Printf("既に初期化済みです: %s\n", result.ShelfDir)
	}
	return nil
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
		title       string
		kind        string
		status      string
		tags        []string
		due         string
		repeatEvery string
		parent      string
		body        string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new task",
		Example: "  shelf add --title \"Weekly Goal\"\n" +
			"  shelf add --title \"Write report\" --kind todo --status in_progress --tag backend --due 2026-03-31\n" +
			"  shelf add --title \"Research note\" --kind memo --parent root",
		RunE: func(_ *cobra.Command, _ []string) error {
			var input shelf.AddTaskInput
			if strings.TrimSpace(title) == "" {
				interactiveInput, err := resolveAddInputInteractive(ctx, body, status, due, repeatEvery, parseTagFlagValues(tags))
				if err != nil {
					return err
				}
				input = interactiveInput
			} else {
				input = shelf.AddTaskInput{
					Title:       title,
					Kind:        shelf.Kind(kind),
					Status:      shelf.Status(status),
					Tags:        parseTagFlagValues(tags),
					DueOn:       due,
					RepeatEvery: repeatEvery,
					Parent:      parent,
					Body:        body,
				}
			}

			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "add"); err != nil {
					return err
				}
				task, err := shelf.AddTask(ctx.rootDir, input)
				if err != nil {
					return err
				}

				fmt.Printf("Created: [%s] %s\n", shelf.ShortID(task.ID), task.Title)
				fmt.Printf("ID: %s\n", task.ID)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Task title")
	cmd.Flags().StringVar(&kind, "kind", "", "Task kind")
	cmd.Flags().StringVar(&status, "status", "", "Task status")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Task tag (repeatable, free input)")
	cmd.Flags().StringVar(&due, "due", "", "Task due date (YYYY-MM-DD|today|tomorrow|+Nd|-Nd|next-week|this-week|mon..sun|next-mon..next-sun|in N days)")
	cmd.Flags().StringVar(&repeatEvery, "repeat-every", "", "Repeat interval (<N>d|<N>w|<N>m|<N>y)")
	cmd.Flags().StringVar(&parent, "parent", "", "Parent task ID or root")
	cmd.Flags().StringVar(&body, "body", "", "Task body")
	return cmd
}
