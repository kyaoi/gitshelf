package shelf

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kyaoi/gitshelf/internal/paths"
)

func TestResolveShelfRootWithOverride(t *testing.T) {
	tmp := t.TempDir()
	if _, err := Initialize(tmp, false); err != nil {
		t.Fatal(err)
	}

	root, err := ResolveShelfRoot(tmp, tmp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if root != tmp {
		t.Fatalf("expected root %q, got %q", tmp, root)
	}
}

func TestResolveShelfRootByWalkingUp(t *testing.T) {
	tmp := t.TempDir()
	if _, err := Initialize(tmp, false); err != nil {
		t.Fatal(err)
	}

	child := filepath.Join(tmp, "a", "b")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}

	root, err := ResolveShelfRoot("", child)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if root != tmp {
		t.Fatalf("expected root %q, got %q", tmp, root)
	}
}

func TestResolveShelfRootNotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))
	_, err := ResolveShelfRoot("", tmp)
	if !errors.Is(err, ErrShelfNotFound) {
		t.Fatalf("expected ErrShelfNotFound, got %v", err)
	}
}

func TestResolveShelfRootFallsBackToGlobalDefaultRoot(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))

	globalRoot := filepath.Join(tmp, "global-store")
	if _, err := Initialize(globalRoot, false); err != nil {
		t.Fatalf("initialize global root failed: %v", err)
	}
	if err := paths.SaveGlobalConfig(paths.GlobalConfig{DefaultRoot: globalRoot}); err != nil {
		t.Fatalf("save global config failed: %v", err)
	}

	unrelatedCwd := filepath.Join(tmp, "workspace")
	if err := os.MkdirAll(unrelatedCwd, 0o755); err != nil {
		t.Fatal(err)
	}

	root, err := ResolveShelfRoot("", unrelatedCwd)
	if err != nil {
		t.Fatalf("expected fallback root, got error: %v", err)
	}
	if root != globalRoot {
		t.Fatalf("expected %q, got %q", globalRoot, root)
	}
}
