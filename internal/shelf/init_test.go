package shelf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitializeCreatesShelfLayout(t *testing.T) {
	root := t.TempDir()

	result, err := Initialize(root, false)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if !result.ConfigCreated {
		t.Fatal("expected config to be created")
	}

	for _, p := range []string{ShelfDir(root), TasksDir(root), EdgesDir(root), ConfigPath(root)} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("path does not exist: %s (%v)", p, err)
		}
	}
	if _, err := os.Stat(filepath.Join(ShelfDir(root), "templates")); !os.IsNotExist(err) {
		t.Fatalf("templates dir should not exist anymore: %v", err)
	}
}

func TestInitializeIsIdempotentWithoutForce(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("first initialize failed: %v", err)
	}

	cfgPath := ConfigPath(root)
	original, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("second initialize failed: %v", err)
	}

	after, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(original) {
		t.Fatal("config should not change without --force")
	}
}

func TestInitializeNormalizesShelfDirTarget(t *testing.T) {
	root := t.TempDir()

	result, err := Initialize(filepath.Join(root, ".shelf"), false)
	if err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if result.RootDir != root {
		t.Fatalf("expected root %q, got %q", root, result.RootDir)
	}
	if _, err := os.Stat(ConfigPath(root)); err != nil {
		t.Fatalf("config should exist at normalized root: %v", err)
	}
}

func TestInitializeRejectsPathInsideShelf(t *testing.T) {
	root := t.TempDir()

	_, err := Initialize(filepath.Join(root, ".shelf", "tasks"), false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInitializeCreatesConfiguredStorageDirs(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("first initialize failed: %v", err)
	}

	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	cfg.StorageRoot = "."
	if err := SaveConfig(root, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("second initialize failed: %v", err)
	}
	for _, p := range []string{
		filepath.Join(root, "tasks"),
		filepath.Join(root, "edges"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("path does not exist: %s (%v)", p, err)
		}
	}
}
