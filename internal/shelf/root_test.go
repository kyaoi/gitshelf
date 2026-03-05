package shelf

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveShelfRootWithOverride(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, ShelfDirName), 0o755); err != nil {
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
	if err := os.Mkdir(filepath.Join(tmp, ShelfDirName), 0o755); err != nil {
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
	_, err := ResolveShelfRoot("", tmp)
	if !errors.Is(err, ErrShelfNotFound) {
		t.Fatalf("expected ErrShelfNotFound, got %v", err)
	}
}
