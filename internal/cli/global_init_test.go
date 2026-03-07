package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kyaoi/gitshelf/internal/paths"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestCLIInitGlobalCreatesConfigAndShelf(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmp, "data"))

	cmd := NewRootCommand("test")
	cmd.SetArgs([]string{"init", "--global"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init --global failed: %v", err)
	}

	cfg, err := paths.LoadGlobalConfig()
	if err != nil {
		t.Fatalf("load global config failed: %v", err)
	}
	expectedRoot, _ := filepath.Abs(filepath.Join(tmp, "data", "gitshelf"))
	if cfg.DefaultRoot != expectedRoot {
		t.Fatalf("expected global root %q, got %q", expectedRoot, cfg.DefaultRoot)
	}

	if _, err := shelf.LoadConfig(cfg.DefaultRoot); err != nil {
		t.Fatalf("global shelf config should exist: %v", err)
	}
}

func TestNoLocalAndNoGlobalShowsGuidance(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmp, "data"))
	t.Chdir(filepath.Join(tmp))

	cmd := NewRootCommand("test")
	cmd.SetArgs([]string{"ls"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "shelf init --global") {
		t.Fatalf("expected guidance message, got: %v", err)
	}
}
