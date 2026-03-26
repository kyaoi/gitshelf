package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newConfigCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Update persisted shelf config values",
	}
	cmd.AddCommand(newConfigShowCommand(ctx))
	cmd.AddCommand(newConfigCopyPresetCommand(ctx))
	return cmd
}

func newConfigShowCommand(ctx *commandContext) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the effective shelf config",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			if asJSON {
				data, err := json.MarshalIndent(buildConfigShowPayload(ctx.rootDir, cfg), "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}
			printConfigShow(ctx.rootDir, cfg)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newConfigCopyPresetCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy-preset",
		Short: "Manage saved Cockpit copy presets",
	}
	cmd.AddCommand(newConfigCopyPresetListCommand(ctx))
	cmd.AddCommand(newConfigCopyPresetGetCommand(ctx))
	cmd.AddCommand(newConfigCopyPresetRemoveCommand(ctx))
	cmd.AddCommand(newConfigCopyPresetSetCommand(ctx))
	return cmd
}

func newConfigCopyPresetListCommand(ctx *commandContext) *cobra.Command {
	var (
		asJSON   bool
		format   string
		fields   string
		header   bool
		noHeader bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved Cockpit copy presets",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := validateFormat(format, []string{"compact", "tsv", "csv", "jsonl"}); err != nil {
				return err
			}
			if strings.TrimSpace(fields) != "" && format != "tsv" && format != "csv" {
				return fmt.Errorf("--fields requires --format tsv or csv")
			}
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			if asJSON {
				data, err := json.MarshalIndent(buildCopyPresetListPayload(cfg.Commands.Cockpit.CopyPresets), "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}
			records := buildCopyPresetListPayload(cfg.Commands.Cockpit.CopyPresets)
			if format == "jsonl" {
				text, err := renderJSONL(records)
				if err != nil {
					return err
				}
				fmt.Print(text)
				return nil
			}
			if format == "tsv" {
				selectedFields, err := resolveTSVFields(fields, defaultCopyPresetTabularFields(), allowedCopyPresetTabularFields())
				if err != nil {
					return err
				}
				includeHeader, err := resolveTabularHeader(format, header, noHeader)
				if err != nil {
					return err
				}
				if includeHeader {
					fmt.Println(strings.Join(selectedFields, "\t"))
				}
				for _, record := range records {
					fmt.Println(joinTSVFields(selectedFields, record.TSVFields()))
				}
				return nil
			}
			if format == "csv" {
				selectedFields, err := resolveTSVFields(fields, defaultCopyPresetTabularFields(), allowedCopyPresetTabularFields())
				if err != nil {
					return err
				}
				includeHeader, err := resolveTabularHeader(format, header, noHeader)
				if err != nil {
					return err
				}
				text, err := renderCSV(records, selectedFields, includeHeader)
				if err != nil {
					return err
				}
				fmt.Print(text)
				return nil
			}
			printCopyPresetList(cfg.Commands.Cockpit.CopyPresets)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&format, "format", "compact", "Output format: compact|tsv|csv|jsonl")
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated field names for --format tsv or csv")
	cmd.Flags().BoolVar(&header, "header", false, "Include a header row for tabular output")
	cmd.Flags().BoolVar(&noHeader, "no-header", false, "Omit the header row for tabular output")
	return cmd
}

func newConfigCopyPresetGetCommand(ctx *commandContext) *cobra.Command {
	var (
		asJSON   bool
		format   string
		fields   string
		header   bool
		noHeader bool
	)

	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Show one saved Cockpit copy preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if err := validateFormat(format, []string{"compact", "tsv", "csv", "jsonl"}); err != nil {
				return err
			}
			if strings.TrimSpace(fields) != "" && format != "tsv" && format != "csv" {
				return fmt.Errorf("--fields requires --format tsv or csv")
			}
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			name := strings.TrimSpace(args[0])
			preset, ok := cfg.FindCopyPreset(name)
			if !ok {
				return fmt.Errorf("copy preset not found: %s", name)
			}
			if asJSON {
				data, err := json.MarshalIndent(buildCopyPresetPayload(preset), "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}
			record := buildCopyPresetPayload(preset)
			if format == "jsonl" {
				text, err := renderJSONL([]copyPresetRecord{record})
				if err != nil {
					return err
				}
				fmt.Print(text)
				return nil
			}
			if format == "tsv" {
				selectedFields, err := resolveTSVFields(fields, defaultCopyPresetTabularFields(), allowedCopyPresetTabularFields())
				if err != nil {
					return err
				}
				includeHeader, err := resolveTabularHeader(format, header, noHeader)
				if err != nil {
					return err
				}
				if includeHeader {
					fmt.Println(strings.Join(selectedFields, "\t"))
				}
				fmt.Println(joinTSVFields(selectedFields, record.TSVFields()))
				return nil
			}
			if format == "csv" {
				selectedFields, err := resolveTSVFields(fields, defaultCopyPresetTabularFields(), allowedCopyPresetTabularFields())
				if err != nil {
					return err
				}
				includeHeader, err := resolveTabularHeader(format, header, noHeader)
				if err != nil {
					return err
				}
				text, err := renderCSV([]copyPresetRecord{record}, selectedFields, includeHeader)
				if err != nil {
					return err
				}
				fmt.Print(text)
				return nil
			}
			printCopyPreset(preset)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&format, "format", "compact", "Output format: compact|tsv|csv|jsonl")
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated field names for --format tsv or csv")
	cmd.Flags().BoolVar(&header, "header", false, "Include a header row for tabular output")
	cmd.Flags().BoolVar(&noHeader, "no-header", false, "Omit the header row for tabular output")
	return cmd
}

func newConfigCopyPresetRemoveCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <name>",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove one saved Cockpit copy preset",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			name := strings.TrimSpace(args[0])
			if !cfg.DeleteCopyPreset(name) {
				return fmt.Errorf("copy preset not found: %s", name)
			}
			if err := shelf.SaveConfig(ctx.rootDir, cfg); err != nil {
				return err
			}
			fmt.Printf("Deleted copy preset: %s\n", name)
			return nil
		},
	}
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
				fmt.Printf("Updated copy preset: %s\n", preset.Name)
			} else {
				fmt.Printf("Saved copy preset: %s\n", preset.Name)
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

type configShowPayload struct {
	RootDir       string                    `json:"root_dir"`
	ConfigPath    string                    `json:"config_path"`
	StorageRoot   string                    `json:"storage_root"`
	StorageDir    string                    `json:"storage_dir"`
	Kinds         []string                  `json:"kinds"`
	Statuses      []string                  `json:"statuses"`
	Tags          []string                  `json:"tags,omitempty"`
	DefaultKind   string                    `json:"default_kind"`
	DefaultStatus string                    `json:"default_status"`
	LinkTypes     configShowLinkTypes       `json:"link_types"`
	Commands      configShowCommandsPayload `json:"commands"`
}

type configShowLinkTypes struct {
	Names    []string `json:"names"`
	Blocking string   `json:"blocking"`
}

type configShowCommandsPayload struct {
	Calendar configShowCalendarPayload `json:"calendar"`
	Cockpit  configShowCockpitPayload  `json:"cockpit"`
}

type configShowCalendarPayload struct {
	DefaultRangeUnit string `json:"default_range_unit"`
	DefaultDays      int    `json:"default_days"`
	DefaultMonths    int    `json:"default_months"`
	DefaultYears     int    `json:"default_years"`
}

type configShowCockpitPayload struct {
	CopySeparator     string             `json:"copy_separator"`
	CopyPresets       []copyPresetRecord `json:"copy_presets,omitempty"`
	PostExitGitAction string             `json:"post_exit_git_action"`
	CommitMessage     string             `json:"commit_message"`
}

func buildConfigShowPayload(rootDir string, cfg shelf.Config) configShowPayload {
	storageDir, err := shelf.ResolveStorageRootDir(rootDir, cfg.StorageRoot)
	if err != nil {
		storageDir = filepath.Join(rootDir, cfg.StorageRoot)
	}
	return configShowPayload{
		RootDir:       rootDir,
		ConfigPath:    shelf.ConfigPath(rootDir),
		StorageRoot:   cfg.StorageRoot,
		StorageDir:    storageDir,
		Kinds:         kindStrings(cfg.Kinds),
		Statuses:      statusStrings(cfg.Statuses),
		Tags:          append([]string{}, cfg.Tags...),
		DefaultKind:   string(cfg.DefaultKind),
		DefaultStatus: string(cfg.DefaultStatus),
		LinkTypes: configShowLinkTypes{
			Names:    linkTypeStrings(cfg.LinkTypes.Names),
			Blocking: string(cfg.LinkTypes.Blocking),
		},
		Commands: configShowCommandsPayload{
			Calendar: configShowCalendarPayload{
				DefaultRangeUnit: cfg.Commands.Calendar.DefaultRangeUnit,
				DefaultDays:      cfg.Commands.Calendar.DefaultDays,
				DefaultMonths:    cfg.Commands.Calendar.DefaultMonths,
				DefaultYears:     cfg.Commands.Calendar.DefaultYears,
			},
			Cockpit: configShowCockpitPayload{
				CopySeparator:     cfg.Commands.Cockpit.CopySeparator,
				CopyPresets:       buildCopyPresetListPayload(cfg.Commands.Cockpit.CopyPresets),
				PostExitGitAction: cfg.Commands.Cockpit.PostExitGitAction,
				CommitMessage:     cfg.Commands.Cockpit.CommitMessage,
			},
		},
	}
}

func buildCopyPresetListPayload(presets []shelf.CopyPreset) []copyPresetRecord {
	items := make([]copyPresetRecord, 0, len(presets))
	for _, preset := range presets {
		items = append(items, buildCopyPresetPayload(preset))
	}
	return items
}

func buildCopyPresetPayload(preset shelf.CopyPreset) copyPresetRecord {
	return buildCopyPresetRecord(preset)
}

func printConfigShow(rootDir string, cfg shelf.Config) {
	payload := buildConfigShowPayload(rootDir, cfg)
	fmt.Printf("Config: %s\n", payload.ConfigPath)
	fmt.Printf("Root: %s\n", payload.RootDir)
	fmt.Printf("Storage: %s (%s)\n", payload.StorageRoot, payload.StorageDir)
	fmt.Printf("Kinds: %s\n", joinOrDash(payload.Kinds))
	fmt.Printf("Statuses: %s\n", joinOrDash(payload.Statuses))
	fmt.Printf("Tags: %s\n", joinOrDash(payload.Tags))
	fmt.Printf("Defaults: kind=%s status=%s\n", payload.DefaultKind, payload.DefaultStatus)
	fmt.Printf("Link Types: names=%s blocking=%s\n", joinOrDash(payload.LinkTypes.Names), payload.LinkTypes.Blocking)
	fmt.Printf("Calendar Defaults: unit=%s days=%d months=%d years=%d\n", payload.Commands.Calendar.DefaultRangeUnit, payload.Commands.Calendar.DefaultDays, payload.Commands.Calendar.DefaultMonths, payload.Commands.Calendar.DefaultYears)
	fmt.Printf("Cockpit: copy_separator=%q post_exit_git_action=%s commit_message=%q\n", payload.Commands.Cockpit.CopySeparator, payload.Commands.Cockpit.PostExitGitAction, payload.Commands.Cockpit.CommitMessage)
	printCopyPresetList(cfg.Commands.Cockpit.CopyPresets)
}

func printCopyPresetList(presets []shelf.CopyPreset) {
	fmt.Println("Copy Presets:")
	if len(presets) == 0 {
		fmt.Println("  (none)")
		return
	}
	for _, preset := range presets {
		preview := strings.ReplaceAll(preset.Template, "\n", `\n`)
		if len(preview) > 48 {
			preview = preview[:48] + "..."
		}
		fmt.Printf("  %s scope=%s subtree_style=%s template=%q\n", preset.Name, preset.Scope, preset.EffectiveSubtreeStyle(), preview)
	}
}

func printCopyPreset(preset shelf.CopyPreset) {
	fmt.Printf("Name: %s\n", preset.Name)
	fmt.Printf("Scope: %s\n", preset.Scope)
	fmt.Printf("Subtree Style: %s\n", preset.EffectiveSubtreeStyle())
	fmt.Printf("Template:\n%s\n", preset.Template)
	fmt.Printf("Join With: %q\n", preset.JoinWith)
}

func defaultCopyPresetTabularFields() []string {
	return []string{"name", "scope", "subtree_style", "template", "join_with"}
}

func allowedCopyPresetTabularFields() map[string]struct{} {
	return map[string]struct{}{
		"name": {}, "scope": {}, "subtree_style": {}, "template": {}, "join_with": {},
	}
}

func kindStrings(values []shelf.Kind) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

func statusStrings(values []shelf.Status) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

func linkTypeStrings(values []shelf.LinkType) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

func joinOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ",")
}
