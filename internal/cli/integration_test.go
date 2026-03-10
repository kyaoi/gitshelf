package cli

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestCLIHelpShowsMinimalCockpitSurface(t *testing.T) {
	out, err := executeCLI(t, "--help")
	if err != nil {
		t.Fatalf("help failed: %v", err)
	}
	for _, name := range []string{
		"board", "calendar", "cockpit", "completion", "init", "link", "links", "ls", "next", "now", "review", "tree", "unlink",
	} {
		if !strings.Contains(out, "\n  "+name+"\t") && !strings.Contains(out, "\n  "+name+" ") {
			t.Fatalf("help should contain %q: %s", name, out)
		}
	}
	for _, name := range []string{
		"add", "archive", "capture", "deps", "doctor", "edit", "export", "github", "import", "show", "triage", "view",
	} {
		if strings.Contains(out, "\n  "+name+"\t") || strings.Contains(out, "\n  "+name+" ") {
			t.Fatalf("help should not contain removed command %q: %s", name, out)
		}
	}
}

func TestCLIRemovedCommandsReturnUnknown(t *testing.T) {
	for _, name := range []string{"add", "show", "edit", "capture", "triage", "doctor", "github", "view", "import", "export"} {
		_, err := executeCLI(t, name)
		if err == nil || !strings.Contains(err.Error(), "unknown command") {
			t.Fatalf("%s should be unknown, got %v", name, err)
		}
	}
}

func TestCLILauncherHelpMatchesCockpitOnlySurface(t *testing.T) {
	for _, args := range [][]string{
		{"calendar", "--help"},
		{"tree", "--help"},
		{"review", "--help"},
		{"now", "--help"},
	} {
		out, err := executeCLI(t, args...)
		if err != nil {
			t.Fatalf("%v help failed: %v", args, err)
		}
		if strings.Contains(out, "--plain") || strings.Contains(out, "--json") || strings.Contains(out, "--carry-over") {
			t.Fatalf("%v help should not mention legacy launcher flags: %s", args, out)
		}
		if !strings.Contains(out, "--days") || !strings.Contains(out, "--months") || !strings.Contains(out, "--years") {
			t.Fatalf("%v help should expose shared cockpit range flags: %s", args, out)
		}
	}
}

func TestCLIInitAndLsFilters(t *testing.T) {
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

func TestCLILinkCommands(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	from, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "From", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add from failed: %v", err)
	}
	to, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "To", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add to failed: %v", err)
	}

	if _, err := executeCLI(t, "link", "--root", root, "--from", from.ID, "--to", to.ID, "--type", "depends_on"); err != nil {
		t.Fatalf("link failed: %v", err)
	}
	output, err := executeCLI(t, "links", "--root", root, from.ID)
	if err != nil {
		t.Fatalf("links failed: %v", err)
	}
	if !strings.Contains(output, "--depends_on-->") || !strings.Contains(output, "root > From") || !strings.Contains(output, "root > To") {
		t.Fatalf("unexpected links output: %s", output)
	}
	if strings.Contains(output, from.ID) || strings.Contains(output, to.ID) {
		t.Fatalf("expected IDs to stay hidden by default, got: %s", output)
	}
	if _, err := executeCLI(t, "unlink", "--root", root, "--from", from.ID, "--to", to.ID, "--type", "depends_on"); err != nil {
		t.Fatalf("unlink failed: %v", err)
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

func TestCLINextListsReadyTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	ready, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "ready task", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add ready task failed: %v", err)
	}
	blocker, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "blocker", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add blocker failed: %v", err)
	}
	blocked, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "blocked task", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add blocked task failed: %v", err)
	}
	if err := shelf.LinkTasks(root, blocked.ID, blocker.ID, "depends_on"); err != nil {
		t.Fatalf("add link failed: %v", err)
	}

	out, err := executeCLI(t, "next", "--root", root)
	if err != nil {
		t.Fatalf("next failed: %v", err)
	}
	if !strings.Contains(out, ready.Title) || strings.Contains(out, blocked.Title) {
		t.Fatalf("unexpected next output: %s", out)
	}

	jsonOut, err := executeCLI(t, "next", "--root", root, "--json")
	if err != nil {
		t.Fatalf("next --json failed: %v", err)
	}
	var items []map[string]any
	if err := json.Unmarshal([]byte(jsonOut), &items); err != nil {
		t.Fatalf("parse next json failed: %v\n%s", err, jsonOut)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 ready tasks, got %d: %s", len(items), jsonOut)
	}
}

func TestCLICompletionBash(t *testing.T) {
	out, err := executeCLI(t, "completion", "bash")
	if err != nil {
		t.Fatalf("completion failed: %v", err)
	}
	if !strings.Contains(out, "shelf") {
		t.Fatalf("unexpected completion output: %s", out)
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
