package cli

import (
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
		"--template", "{{path}}\n{{subtree}}",
		"--join-with", "\n\n",
	)
	if err != nil {
		t.Fatalf("config copy-preset set failed: %v", err)
	}
	if !strings.Contains(out, "copy preset を保存しました: subtree-path") {
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
	if got.Name != "subtree-path" || got.Scope != shelf.CopyPresetScopeSubtree || got.Template != "{{path}}\n{{subtree}}" || got.JoinWith != "\n\n" {
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
		Name:     "subtree-path",
		Scope:    shelf.CopyPresetScopeSubtree,
		Template: "{{path}}\n{{subtree}}",
		JoinWith: "\n\n",
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
		"--template", "{{title}}",
	)
	if err != nil {
		t.Fatalf("config copy-preset set failed: %v", err)
	}
	if !strings.Contains(out, "copy preset を更新しました: subtree-path") {
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
	if got.Scope != shelf.CopyPresetScopeTask || got.Template != "{{title}}" || got.JoinWith != "" {
		t.Fatalf("unexpected updated preset: %+v", got)
	}
}
