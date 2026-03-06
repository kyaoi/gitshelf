package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestCLIInitAddDoctorFlow(t *testing.T) {
	root := t.TempDir()

	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if _, err := executeCLI(t, "add", "--root", root, "--title", "integration task"); err != nil {
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

	if _, err := executeCLI(t, "doctor", "--root", root); err != nil {
		t.Fatalf("doctor failed: %v", err)
	}
}

func TestCLILsFilters(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	tasks := []shelf.AddTaskInput{
		{Title: "todo-open", Kind: "todo", Status: "open"},
		{Title: "todo-in-progress", Kind: "todo", Status: "in_progress"},
		{Title: "memo-open", Kind: "memo", Status: "open"},
		{Title: "todo-done", Kind: "todo", Status: "done"},
		{Title: "todo-cancelled", Kind: "todo", Status: "cancelled"},
	}
	for _, input := range tasks {
		if _, err := shelf.AddTask(root, input); err != nil {
			t.Fatalf("add task failed: %v", err)
		}
	}

	output, err := executeCLI(t, "ls", "--root", root, "--kind", "todo", "--status", "open")
	if err != nil {
		t.Fatalf("ls kind/status failed: %v", err)
	}
	if !strings.Contains(output, "todo-open") || strings.Contains(output, "memo-open") || strings.Contains(output, "todo-done") {
		t.Fatalf("unexpected output for kind+status filter: %s", output)
	}

	output, err = executeCLI(t, "ls", "--root", root, "--not-status", "done", "--not-status", "cancelled")
	if err != nil {
		t.Fatalf("ls not-status failed: %v", err)
	}
	if strings.Contains(output, "todo-done") || strings.Contains(output, "todo-cancelled") {
		t.Fatalf("unexpected output for not-status filter: %s", output)
	}

	output, err = executeCLI(t, "ls", "--root", root, "--status", "open", "--status", "in_progress")
	if err != nil {
		t.Fatalf("ls multi-status failed: %v", err)
	}
	if !strings.Contains(output, "todo-open") || !strings.Contains(output, "todo-in-progress") || strings.Contains(output, "todo-done") {
		t.Fatalf("unexpected output for multi-status filter: %s", output)
	}
}

func TestCLILsUnknownFilterValues(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "task"}); err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	if _, err := executeCLI(t, "ls", "--root", root, "--status", "unknown"); err == nil || !strings.Contains(err.Error(), "unknown status") {
		t.Fatalf("expected unknown status error, got: %v", err)
	}
	if _, err := executeCLI(t, "ls", "--root", root, "--kind", "unknown"); err == nil || !strings.Contains(err.Error(), "unknown kind") {
		t.Fatalf("expected unknown kind error, got: %v", err)
	}
}

func TestCLIEditWithIDUsesEditor(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "edit me"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	t.Setenv("VISUAL", "true")
	t.Setenv("EDITOR", "")
	if _, err := executeCLI(t, "edit", "--root", root, task.ID); err != nil {
		t.Fatalf("edit failed: %v", err)
	}
}

func TestCLIEditReturnsEditorExitError(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "edit fail"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	t.Setenv("VISUAL", "false")
	t.Setenv("EDITOR", "")
	if _, err := executeCLI(t, "edit", "--root", root, task.ID); err == nil || !strings.Contains(err.Error(), "editor exited with status") {
		t.Fatalf("expected editor exit error, got: %v", err)
	}
}

func executeCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewRootCommand("test")
	cmd.SetArgs(args)

	stdout := os.Stdout
	stderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe create failed: %v", err)
	}
	os.Stdout = w
	os.Stderr = w

	execErr := cmd.Execute()
	_ = w.Close()
	os.Stdout = stdout
	os.Stderr = stderr

	data, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatalf("pipe read failed: %v", readErr)
	}
	return string(data), execErr
}
