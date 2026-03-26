package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kyaoi/gitshelf/internal/paths"
	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

type commandContext struct {
	rootOverride string
	rootDir      string
	showID       bool
	gitOnExit    string
	gitMessage   string
}

func NewRootCommand(version string) *cobra.Command {
	ctx := &commandContext{}

	cmd := &cobra.Command{
		Use:           "shelf",
		Short:         "Git-backed task shelf with a Cockpit workspace",
		Long:          "shelf is a lightweight CLI tool for managing tasks and links in .shelf/. In TTY, running `shelf` opens the Cockpit workspace.",
		SilenceUsage:  true,
		SilenceErrors: false,
		Version:       version,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Name() == "init" || isCompletionCommand(cmd) {
				return nil
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			rootDir, err := shelf.ResolveShelfRoot(ctx.rootOverride, cwd)
			if err != nil {
				if errors.Is(err, shelf.ErrShelfNotFound) {
					return errors.New(".shelf not found. Run `shelf init` or `shelf init --global`.")
				}
				return err
			}
			ctx.rootDir = rootDir
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !dailyCockpitIsTTY() {
				return cmd.Help()
			}
			return runDefaultCockpit(ctx)
		},
	}

	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.PersistentFlags().StringVar(&ctx.rootOverride, "root", "", "Directory that contains .shelf")
	cmd.PersistentFlags().BoolVarP(&ctx.showID, "show-id", "i", false, "Show task IDs in list/tree/interactive labels")
	cmd.PersistentFlags().StringVar(&ctx.gitOnExit, "git-on-exit", "", "Run git action after Cockpit exits: none|commit|commit_push")
	cmd.PersistentFlags().StringVar(&ctx.gitMessage, "git-message", "", "Commit message used when --git-on-exit creates a commit")

	cmd.AddCommand(newInitCommand(ctx))
	cmd.AddCommand(newCompletionCommand())
	cmd.AddCommand(newConfigCommand(ctx))
	cmd.AddCommand(newCockpitCommand(ctx))
	cmd.AddCommand(newCalendarCommand(ctx))
	cmd.AddCommand(newBoardCommand(ctx))
	cmd.AddCommand(newReviewCommand(ctx))
	cmd.AddCommand(newLsCommand(ctx))
	cmd.AddCommand(newLinkCommand(ctx))
	cmd.AddCommand(newLinksCommand(ctx))
	cmd.AddCommand(newNextCommand(ctx))
	cmd.AddCommand(newNowCommand(ctx))
	cmd.AddCommand(newShowCommand(ctx))
	cmd.AddCommand(newTreeCommand(ctx))
	cmd.AddCommand(newUnlinkCommand(ctx))

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
					return fmt.Errorf("failed to get working directory: %w", err)
				}
				targetDir = cwd
			}

			result, err := shelf.Initialize(targetDir, force)
			if err != nil {
				return err
			}
			switch {
			case result.ConfigForced:
				fmt.Printf("Initialized: %s (rewrote config.toml)\n", result.ShelfDir)
			case result.ConfigCreated:
				fmt.Printf("Initialized: %s\n", result.ShelfDir)
			default:
				fmt.Printf("Already initialized: %s\n", result.ShelfDir)
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
		normalized, err := shelf.NormalizeRootDir(rootOverride)
		if err != nil {
			return err
		}
		defaultRoot = normalized
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

	fmt.Printf("Global config: %s\n", globalPath)
	fmt.Printf("default_root: %s\n", defaultRoot)
	switch {
	case result.ConfigForced:
		fmt.Printf("Initialized: %s (rewrote config.toml)\n", result.ShelfDir)
	case result.ConfigCreated:
		fmt.Printf("Initialized: %s\n", result.ShelfDir)
	default:
		fmt.Printf("Already initialized: %s\n", result.ShelfDir)
	}
	return nil
}
