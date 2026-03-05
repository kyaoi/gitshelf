package cli

import (
	"errors"
	"fmt"
	"os"

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

	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newStubCommand("add"))
	cmd.AddCommand(newStubCommand("ls"))
	cmd.AddCommand(newStubCommand("tree"))
	cmd.AddCommand(newStubCommand("show"))
	cmd.AddCommand(newStubCommand("set"))
	cmd.AddCommand(newStubCommand("mv"))
	cmd.AddCommand(newStubCommand("link"))
	cmd.AddCommand(newStubCommand("unlink"))
	cmd.AddCommand(newStubCommand("links"))
	cmd.AddCommand(newStubCommand("doctor"))

	return cmd
}

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize .shelf in the current directory",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("not implemented yet")
		},
	}
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
