package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newPresetCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preset",
		Short: "Manage output presets",
	}
	cmd.AddCommand(newPresetListCommand(ctx))
	cmd.AddCommand(newPresetShowCommand(ctx))
	cmd.AddCommand(newPresetSetCommand(ctx))
	cmd.AddCommand(newPresetDeleteCommand(ctx))
	return cmd
}

func newPresetListCommand(ctx *commandContext) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List output presets",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			names := make([]string, 0, len(cfg.OutputPresets))
			for name := range cfg.OutputPresets {
				names = append(names, name)
			}
			sort.Strings(names)

			if asJSON {
				data, err := json.MarshalIndent(cfg.OutputPresets, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}
			if len(names) == 0 {
				fmt.Println(uiMuted("(none)"))
				return nil
			}
			for _, name := range names {
				p := cfg.OutputPresets[name]
				fmt.Printf("%s command=%s format=%s view=%s limit=%d\n", name, p.Command, p.Format, p.View, p.Limit)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newPresetShowCommand(ctx *commandContext) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show an output preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			preset, ok := cfg.OutputPresets[name]
			if !ok {
				return fmt.Errorf("unknown output preset: %s", name)
			}
			if asJSON {
				data, err := json.MarshalIndent(map[string]any{
					"name":   name,
					"preset": preset,
				}, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}
			fmt.Printf("name: %s\n", name)
			fmt.Printf("command: %s\n", preset.Command)
			fmt.Printf("format: %s\n", preset.Format)
			fmt.Printf("view: %s\n", preset.View)
			fmt.Printf("limit: %d\n", preset.Limit)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newPresetSetCommand(ctx *commandContext) *cobra.Command {
	var (
		command string
		format  string
		view    string
		limit   int
	)
	cmd := &cobra.Command{
		Use:   "set <name>",
		Short: "Create or replace an output preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if !cmd.Flags().Changed("command") {
				return fmt.Errorf("--command is required")
			}
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			cfg.OutputPresets[name] = shelf.OutputPreset{
				Command: strings.TrimSpace(command),
				Format:  strings.TrimSpace(format),
				View:    strings.TrimSpace(view),
				Limit:   limit,
			}
			if err := shelf.SaveConfig(ctx.rootDir, cfg); err != nil {
				return err
			}
			fmt.Printf("Saved preset: %s\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&command, "command", "", "Target command: ls|tree|next|agenda|today")
	cmd.Flags().StringVar(&format, "format", "", "Default format for command")
	cmd.Flags().StringVar(&view, "view", "", "Default view")
	cmd.Flags().IntVar(&limit, "limit", 0, "Default limit")
	return cmd
}

func newPresetDeleteCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm"},
		Short:   "Delete an output preset",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			if _, ok := cfg.OutputPresets[name]; !ok {
				return fmt.Errorf("unknown output preset: %s", name)
			}
			delete(cfg.OutputPresets, name)
			if err := shelf.SaveConfig(ctx.rootDir, cfg); err != nil {
				return err
			}
			fmt.Printf("Deleted preset: %s\n", name)
			return nil
		},
	}
	return cmd
}
