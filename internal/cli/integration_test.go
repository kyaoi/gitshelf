package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCLIInitAddDoctorFlow(t *testing.T) {
	root := t.TempDir()

	cmd := NewRootCommand("test")
	cmd.SetArgs([]string{"init", "--root", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cmd = NewRootCommand("test")
	cmd.SetArgs([]string{"add", "--root", root, "--title", "integration task"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	tasksDir := filepath.Join(root, ".shelf", "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		t.Fatalf("failed to read tasks directory: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one task file")
	}

	cmd = NewRootCommand("test")
	cmd.SetArgs([]string{"doctor", "--root", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
}
