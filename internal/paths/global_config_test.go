package paths

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadGlobalConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))

	root := filepath.Join(tmp, "store")
	if err := SaveGlobalConfig(GlobalConfig{DefaultRoot: root}); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	cfg, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	expected, _ := filepath.Abs(root)
	if cfg.DefaultRoot != expected {
		t.Fatalf("expected %q, got %q", expected, cfg.DefaultRoot)
	}
}

func TestLoadGlobalConfigNotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))

	_, err := LoadGlobalConfig()
	if !errors.Is(err, ErrGlobalConfigNotFound) {
		t.Fatalf("expected ErrGlobalConfigNotFound, got %v", err)
	}
}
