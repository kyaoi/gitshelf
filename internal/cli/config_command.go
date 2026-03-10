package cli

import (
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newConfigCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Update persisted shelf config values",
	}
	cmd.AddCommand(newConfigCopyPresetCommand(ctx))
	return cmd
}

func newConfigCopyPresetCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy-preset",
		Short: "Manage saved Cockpit copy presets",
	}
	cmd.AddCommand(newConfigCopyPresetSetCommand(ctx))
	return cmd
}

func newConfigCopyPresetSetCommand(ctx *commandContext) *cobra.Command {
	var (
		name         string
		scope        string
		subtreeStyle string
		template     string
		joinWith     string
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Create or update a saved Cockpit copy preset",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			preset := shelf.CopyPreset{
				Name:         strings.TrimSpace(name),
				Scope:        shelf.CopyPresetScope(strings.TrimSpace(scope)),
				SubtreeStyle: shelf.CopySubtreeStyle(strings.TrimSpace(subtreeStyle)),
				Template:     template,
				JoinWith:     joinWith,
			}
			updated, err := cfg.UpsertCopyPreset(preset)
			if err != nil {
				return err
			}
			if err := shelf.SaveConfig(ctx.rootDir, cfg); err != nil {
				return err
			}
			if updated {
				fmt.Printf("copy preset を更新しました: %s\n", preset.Name)
			} else {
				fmt.Printf("copy preset を保存しました: %s\n", preset.Name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Preset name")
	cmd.Flags().StringVar(&scope, "scope", "", "Preset scope: task|subtree")
	cmd.Flags().StringVar(&subtreeStyle, "subtree-style", string(shelf.CopySubtreeStyleIndented), "Subtree rendering style: indented|tree")
	cmd.Flags().StringVar(&template, "template", "", "Per-item copy template")
	cmd.Flags().StringVar(&joinWith, "join-with", "", "Join separator for multiple rendered items")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("scope")
	_ = cmd.MarkFlagRequired("template")
	return cmd
}
