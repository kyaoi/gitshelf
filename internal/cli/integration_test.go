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

func TestCLITreeLsAndShowHideIDsAndShowHierarchy(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	child, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Child", Kind: "todo", Status: "open", Parent: parent.ID})
	if err != nil {
		t.Fatalf("add child failed: %v", err)
	}

	lsOutput, err := executeCLI(t, "ls", "--root", root)
	if err != nil {
		t.Fatalf("ls failed: %v", err)
	}
	if strings.Contains(lsOutput, "[") || strings.Contains(lsOutput, shelf.ShortID(parent.ID)) {
		t.Fatalf("ls should not display IDs: %s", lsOutput)
	}
	if !strings.Contains(lsOutput, "parent=root") || !strings.Contains(lsOutput, "parent=Parent") {
		t.Fatalf("ls should display parent title hierarchy hint: %s", lsOutput)
	}

	treeOutput, err := executeCLI(t, "tree", "--root", root)
	if err != nil {
		t.Fatalf("tree failed: %v", err)
	}
	if strings.Contains(treeOutput, "[") || strings.Contains(treeOutput, shelf.ShortID(parent.ID)) {
		t.Fatalf("tree should not display IDs: %s", treeOutput)
	}
	if !strings.Contains(treeOutput, "Parent (todo/open)") || !strings.Contains(treeOutput, "└─ Child (todo/open)") {
		t.Fatalf("unexpected tree output: %s", treeOutput)
	}

	lsWithID, err := executeCLI(t, "ls", "--root", root, "--show-id")
	if err != nil {
		t.Fatalf("ls --show-id failed: %v", err)
	}
	if !strings.Contains(lsWithID, "["+shelf.ShortID(parent.ID)+"] Parent") {
		t.Fatalf("ls --show-id should include IDs: %s", lsWithID)
	}

	treeWithID, err := executeCLI(t, "tree", "--root", root, "--show-id")
	if err != nil {
		t.Fatalf("tree --show-id failed: %v", err)
	}
	if !strings.Contains(treeWithID, "["+shelf.ShortID(parent.ID)+"] Parent") {
		t.Fatalf("tree --show-id should include IDs: %s", treeWithID)
	}

	treeWithIDShort, err := executeCLI(t, "tree", "--root", root, "-i")
	if err != nil {
		t.Fatalf("tree -i failed: %v", err)
	}
	if !strings.Contains(treeWithIDShort, "["+shelf.ShortID(parent.ID)+"] Parent") {
		t.Fatalf("tree -i should include IDs: %s", treeWithIDShort)
	}

	showOutput, err := executeCLI(t, "show", "--root", root, child.ID)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if !strings.Contains(showOutput, "Hierarchy:") || !strings.Contains(showOutput, "Path: root > Parent > Child") {
		t.Fatalf("show should include hierarchy path: %s", showOutput)
	}
	if !strings.Contains(showOutput, "Subtree:") || !strings.Contains(showOutput, "Child (todo/open)") {
		t.Fatalf("show should include subtree output: %s", showOutput)
	}
}

func TestCLIPreviewBodyFlagRemoved(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	rootHelp, err := executeCLI(t, "--help")
	if err != nil {
		t.Fatalf("root help failed: %v", err)
	}
	if strings.Contains(rootHelp, "preview-body") {
		t.Fatalf("root help should not include preview-body: %s", rootHelp)
	}

	treeHelp, err := executeCLI(t, "tree", "--help")
	if err != nil {
		t.Fatalf("tree help failed: %v", err)
	}
	if strings.Contains(treeHelp, "preview-body") {
		t.Fatalf("tree help should not include preview-body: %s", treeHelp)
	}

	addHelp, err := executeCLI(t, "add", "--help")
	if err != nil {
		t.Fatalf("add help failed: %v", err)
	}
	if strings.Contains(addHelp, "preview-body") {
		t.Fatalf("add help should not include preview-body: %s", addHelp)
	}

	if _, err := executeCLI(t, "tree", "--root", root, "--preview-body"); err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("expected unknown flag error for preview-body, got: %v", err)
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
