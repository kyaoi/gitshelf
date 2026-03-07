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
