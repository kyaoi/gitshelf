package cli

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestConfigCopyPresetSetCommandPersistsPreset(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	out, err := executeCLI(
		t,
		"--root", root,
		"config", "copy-preset", "set",
		"--name", "subtree-path",
		"--scope", "subtree",
		"--subtree-style", "tree",
		"--template", "{{path}}\n{{subtree}}",
		"--join-with", "\n\n",
	)
	if err != nil {
		t.Fatalf("config copy-preset set failed: %v", err)
	}
	if !strings.Contains(out, "Saved copy preset: subtree-path") {
		t.Fatalf("unexpected command output: %s", out)
	}

	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if len(cfg.Commands.Cockpit.CopyPresets) != 1 {
		t.Fatalf("unexpected copy presets: %+v", cfg.Commands.Cockpit.CopyPresets)
	}
	got := cfg.Commands.Cockpit.CopyPresets[0]
	if got.Name != "subtree-path" || got.Scope != shelf.CopyPresetScopeSubtree || got.SubtreeStyle != shelf.CopySubtreeStyleTree || got.Template != "{{path}}\n{{subtree}}" || got.JoinWith != "\n\n" {
		t.Fatalf("unexpected persisted preset: %+v", got)
	}
}

func TestConfigCopyPresetSetCommandUpdatesExistingPreset(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if _, err := cfg.UpsertCopyPreset(shelf.CopyPreset{
		Name:         "subtree-path",
		Scope:        shelf.CopyPresetScopeSubtree,
		SubtreeStyle: shelf.CopySubtreeStyleIndented,
		Template:     "{{path}}\n{{subtree}}",
		JoinWith:     "\n\n",
	}); err != nil {
		t.Fatalf("seed preset failed: %v", err)
	}
	if err := shelf.SaveConfig(root, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	out, err := executeCLI(
		t,
		"--root", root,
		"config", "copy-preset", "set",
		"--name", "subtree-path",
		"--scope", "task",
		"--subtree-style", "tree",
		"--template", "{{title}}",
	)
	if err != nil {
		t.Fatalf("config copy-preset set failed: %v", err)
	}
	if !strings.Contains(out, "Updated copy preset: subtree-path") {
		t.Fatalf("unexpected command output: %s", out)
	}

	updated, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("reload config failed: %v", err)
	}
	if len(updated.Commands.Cockpit.CopyPresets) != 1 {
		t.Fatalf("unexpected copy preset count: %+v", updated.Commands.Cockpit.CopyPresets)
	}
	got := updated.Commands.Cockpit.CopyPresets[0]
	if got.Scope != shelf.CopyPresetScopeTask || got.SubtreeStyle != shelf.CopySubtreeStyleTree || got.Template != "{{title}}" || got.JoinWith != "" {
		t.Fatalf("unexpected updated preset: %+v", got)
	}
}

func TestConfigShowCommandPrintsEffectiveConfig(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	out, err := executeCLI(t, "--root", root, "config", "show")
	if err != nil {
		t.Fatalf("config show failed: %v", err)
	}
	for _, want := range []string{
		"Config: " + shelf.ConfigPath(root),
		"Storage: .shelf",
		"Defaults: kind=todo status=open",
		"Copy Presets:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestConfigShowCommandJSONIncludesCopyPresetPayload(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if _, err := cfg.UpsertCopyPreset(shelf.CopyPreset{
		Name:         "task-body",
		Scope:        shelf.CopyPresetScopeTask,
		SubtreeStyle: shelf.CopySubtreeStyleIndented,
		Template:     "{{title}}\n{{body}}",
	}); err != nil {
		t.Fatalf("seed preset failed: %v", err)
	}
	if err := shelf.SaveConfig(root, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	out, err := executeCLI(t, "--root", root, "config", "show", "--json")
	if err != nil {
		t.Fatalf("config show --json failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("parse json failed: %v\n%s", err, out)
	}
	commands, ok := payload["commands"].(map[string]any)
	if !ok {
		t.Fatalf("missing commands payload: %v", payload)
	}
	cockpit, ok := commands["cockpit"].(map[string]any)
	if !ok {
		t.Fatalf("missing cockpit payload: %v", commands)
	}
	presets, ok := cockpit["copy_presets"].([]any)
	if !ok || len(presets) != 1 {
		t.Fatalf("unexpected presets payload: %#v", cockpit["copy_presets"])
	}
}

func TestConfigCopyPresetListAndGetCommands(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	for _, preset := range []shelf.CopyPreset{
		{Name: "titles", Scope: shelf.CopyPresetScopeTask, Template: "{{title}}"},
		{Name: "subtree-path", Scope: shelf.CopyPresetScopeSubtree, SubtreeStyle: shelf.CopySubtreeStyleTree, Template: "{{path}}\n{{subtree}}", JoinWith: "\n\n"},
	} {
		if _, err := cfg.UpsertCopyPreset(preset); err != nil {
			t.Fatalf("seed preset failed: %v", err)
		}
	}
	if err := shelf.SaveConfig(root, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	out, err := executeCLI(t, "--root", root, "config", "copy-preset", "list")
	if err != nil {
		t.Fatalf("copy-preset list failed: %v", err)
	}
	if !strings.Contains(out, "titles scope=task") || !strings.Contains(out, "subtree-path scope=subtree") {
		t.Fatalf("unexpected list output: %s", out)
	}

	out, err = executeCLI(t, "--root", root, "config", "copy-preset", "get", "subtree-path")
	if err != nil {
		t.Fatalf("copy-preset get failed: %v", err)
	}
	for _, want := range []string{"Name: subtree-path", "Subtree Style: tree", "{{path}}"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected get output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestConfigCopyPresetGetCommandJSONAndRemove(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if _, err := cfg.UpsertCopyPreset(shelf.CopyPreset{
		Name:         "subtree-path",
		Scope:        shelf.CopyPresetScopeSubtree,
		SubtreeStyle: shelf.CopySubtreeStyleTree,
		Template:     "{{path}}\n{{subtree}}",
	}); err != nil {
		t.Fatalf("seed preset failed: %v", err)
	}
	if err := shelf.SaveConfig(root, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	out, err := executeCLI(t, "--root", root, "config", "copy-preset", "get", "subtree-path", "--json")
	if err != nil {
		t.Fatalf("copy-preset get --json failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("parse json failed: %v\n%s", err, out)
	}
	if payload["name"] != "subtree-path" || payload["scope"] != "subtree" {
		t.Fatalf("unexpected get payload: %#v", payload)
	}

	out, err = executeCLI(t, "--root", root, "config", "copy-preset", "rm", "subtree-path")
	if err != nil {
		t.Fatalf("copy-preset rm failed: %v", err)
	}
	if !strings.Contains(out, "Deleted copy preset: subtree-path") {
		t.Fatalf("unexpected rm output: %s", out)
	}

	updated, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("reload config failed: %v", err)
	}
	if len(updated.Commands.Cockpit.CopyPresets) != 0 {
		t.Fatalf("expected preset to be removed: %+v", updated.Commands.Cockpit.CopyPresets)
	}
}

func TestConfigCopyPresetListCommandCSVAndFields(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if _, err := cfg.UpsertCopyPreset(shelf.CopyPreset{
		Name:         "subtree-path",
		Scope:        shelf.CopyPresetScopeSubtree,
		SubtreeStyle: shelf.CopySubtreeStyleTree,
		Template:     "{{path}}\n{{subtree}}",
		JoinWith:     "\n\n",
	}); err != nil {
		t.Fatalf("seed preset failed: %v", err)
	}
	if err := shelf.SaveConfig(root, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	out, err := executeCLI(t, "--root", root, "config", "copy-preset", "list", "--format", "csv", "--fields", "name,scope,subtree_style", "--no-header")
	if err != nil {
		t.Fatalf("copy-preset list --format csv failed: %v", err)
	}
	rows, err := csv.NewReader(bytes.NewBufferString(out)).ReadAll()
	if err != nil {
		t.Fatalf("parse csv failed: %v\n%s", err, out)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d: %#v", len(rows), rows)
	}
	if rows[0][0] != "subtree-path" || rows[0][1] != "subtree" || rows[0][2] != "tree" {
		t.Fatalf("unexpected csv row: %#v", rows[0])
	}
}

func TestConfigCopyPresetGetCommandJSONL(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if _, err := cfg.UpsertCopyPreset(shelf.CopyPreset{
		Name:         "titles",
		Scope:        shelf.CopyPresetScopeTask,
		SubtreeStyle: shelf.CopySubtreeStyleIndented,
		Template:     "{{title}}",
	}); err != nil {
		t.Fatalf("seed preset failed: %v", err)
	}
	if err := shelf.SaveConfig(root, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	out, err := executeCLI(t, "--root", root, "config", "copy-preset", "get", "titles", "--format", "jsonl")
	if err != nil {
		t.Fatalf("copy-preset get --format jsonl failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 jsonl line, got %d: %q", len(lines), out)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &payload); err != nil {
		t.Fatalf("parse jsonl failed: %v\n%s", err, lines[0])
	}
	if payload["name"] != "titles" || payload["scope"] != "task" {
		t.Fatalf("unexpected jsonl payload: %#v", payload)
	}
}

func TestConfigCopyPresetFieldsRequireTabularFormat(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := executeCLI(t, "--root", root, "config", "copy-preset", "list", "--fields", "name"); err == nil || !strings.Contains(err.Error(), "--fields requires --format tsv or csv") {
		t.Fatalf("expected fields format error, got: %v", err)
	}
}
