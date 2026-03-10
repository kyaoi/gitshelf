package cli

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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
		"board", "calendar", "cockpit", "completion", "config", "init", "link", "links", "ls", "next", "now", "review", "show", "tree", "unlink",
	} {
		if !strings.Contains(out, "\n  "+name+"\t") && !strings.Contains(out, "\n  "+name+" ") {
			t.Fatalf("help should contain %q: %s", name, out)
		}
	}
	for _, name := range []string{
		"add", "archive", "capture", "deps", "doctor", "edit", "export", "github", "import", "triage", "view",
	} {
		if strings.Contains(out, "\n  "+name+"\t") || strings.Contains(out, "\n  "+name+" ") {
			t.Fatalf("help should not contain removed command %q: %s", name, out)
		}
	}
}

func TestCLIRemovedCommandsReturnUnknown(t *testing.T) {
	for _, name := range []string{"add", "edit", "capture", "triage", "doctor", "github", "view", "import", "export"} {
		_, err := executeCLI(t, name)
		if err == nil || !strings.Contains(err.Error(), "unknown command") {
			t.Fatalf("%s should be unknown, got %v", name, err)
		}
	}
}

func TestCLIShowDisplaysInspectorStyleTaskDetails(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:       "Child",
		Kind:        "todo",
		Status:      "in_progress",
		Parent:      parent.ID,
		Tags:        []string{"backend", "cli"},
		DueOn:       "2026-03-15",
		RepeatEvery: "1w",
		Body:        "first line\nsecond line",
	})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}
	dependency, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Dependency", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add dependency failed: %v", err)
	}
	if err := shelf.LinkTasks(root, task.ID, dependency.ID, "depends_on"); err != nil {
		t.Fatalf("link outbound failed: %v", err)
	}
	if err := shelf.LinkTasks(root, parent.ID, task.ID, "related"); err != nil {
		t.Fatalf("link inbound failed: %v", err)
	}

	out, err := executeCLI(t, "show", "--root", root, task.ID)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	for _, want := range []string{
		"Task: root > Parent > Child",
		"Kind: todo",
		"Status: in_progress",
		"Tags: backend, cli",
		"Parent: root > Parent",
		"Body:",
		"  first line",
		"Outbound:",
		"Inbound:",
		"--depends_on--> root > Dependency",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected show output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestCLIShowJSONIncludesPathBodyAndLinks(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Task", Body: "note"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}
	peer, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Peer"})
	if err != nil {
		t.Fatalf("add peer failed: %v", err)
	}
	if err := shelf.LinkTasks(root, peer.ID, task.ID, "related"); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	out, err := executeCLI(t, "show", "--root", root, task.ID, "--json")
	if err != nil {
		t.Fatalf("show --json failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("parse show json failed: %v\n%s", err, out)
	}
	if payload["path"] != "root > Task" || payload["body"] != "note" {
		t.Fatalf("unexpected show payload: %#v", payload)
	}
	if payload["file"] != filepath.Join(shelf.TasksDir(root), task.ID+".md") {
		t.Fatalf("expected show json to include file path, got: %#v", payload)
	}
	inbound, ok := payload["inbound"].([]any)
	if !ok || len(inbound) != 1 {
		t.Fatalf("unexpected inbound payload: %#v", payload["inbound"])
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
	first, ok := items[0]["path"].(string)
	if !ok || !strings.HasPrefix(first, "root > ") {
		t.Fatalf("expected next json items to include path, got: %#v", items)
	}
	if _, ok := items[0]["file"].(string); !ok {
		t.Fatalf("expected next json items to include file, got: %#v", items)
	}
}

func TestCLILsJSONIncludesPathAndTreeFormat(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	child, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Child", Parent: parent.ID})
	if err != nil {
		t.Fatalf("add child failed: %v", err)
	}

	jsonOut, err := executeCLI(t, "ls", "--root", root, "--json")
	if err != nil {
		t.Fatalf("ls --json failed: %v", err)
	}
	var items []map[string]any
	if err := json.Unmarshal([]byte(jsonOut), &items); err != nil {
		t.Fatalf("parse ls json failed: %v\n%s", err, jsonOut)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 tasks, got %d: %s", len(items), jsonOut)
	}
	var childPath string
	var childFile string
	var parentPath string
	for _, item := range items {
		if item["title"] == "Child" {
			childPath, _ = item["path"].(string)
			childFile, _ = item["file"].(string)
			parentPath, _ = item["parent_path"].(string)
		}
	}
	if childPath != "root > Parent > Child" {
		t.Fatalf("unexpected child path: %q", childPath)
	}
	if childFile != filepath.Join(shelf.TasksDir(root), child.ID+".md") {
		t.Fatalf("unexpected child file: %q", childFile)
	}
	if parentPath != "root > Parent" {
		t.Fatalf("unexpected parent path: %q", parentPath)
	}

	out, err := executeCLI(t, "ls", "--root", root, "--format", "tree")
	if err != nil {
		t.Fatalf("ls --format tree failed: %v", err)
	}
	if !strings.Contains(out, "Parent") || !strings.Contains(out, "Child") || !strings.Contains(out, "└─") {
		t.Fatalf("unexpected tree output: %s", out)
	}
}

func TestCLILsPresetNowUsesReadyDefaultsButAllowsOverride(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	ready, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Ready", Status: "open"})
	if err != nil {
		t.Fatalf("add ready failed: %v", err)
	}
	blocker, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Blocker", Status: "open"})
	if err != nil {
		t.Fatalf("add blocker failed: %v", err)
	}
	blocked, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Blocked", Status: "open"})
	if err != nil {
		t.Fatalf("add blocked failed: %v", err)
	}
	if err := shelf.LinkTasks(root, blocked.ID, blocker.ID, "depends_on"); err != nil {
		t.Fatalf("link failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Done", Status: "done"}); err != nil {
		t.Fatalf("add done failed: %v", err)
	}

	out, err := executeCLI(t, "ls", "--root", root, "--preset", "now")
	if err != nil {
		t.Fatalf("ls --preset now failed: %v", err)
	}
	if !strings.Contains(out, ready.Title) || strings.Contains(out, blocked.Title) || strings.Contains(out, "Done") {
		t.Fatalf("unexpected now preset output: %s", out)
	}

	out, err = executeCLI(t, "ls", "--root", root, "--preset", "now", "--status", "open")
	if err != nil {
		t.Fatalf("ls --preset now --status open failed: %v", err)
	}
	if !strings.Contains(out, blocked.Title) || !strings.Contains(out, ready.Title) || strings.Contains(out, "Done") {
		t.Fatalf("expected explicit status to override preset defaults: %s", out)
	}
}

func TestCLILsPresetReviewAndBoardApplyReadSideDefaults(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Open", Status: "open"}); err != nil {
		t.Fatalf("add open failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Done", Status: "done"}); err != nil {
		t.Fatalf("add done failed: %v", err)
	}

	reviewOut, err := executeCLI(t, "ls", "--root", root, "--preset", "review")
	if err != nil {
		t.Fatalf("ls --preset review failed: %v", err)
	}
	if !strings.Contains(reviewOut, "kind=") || strings.Contains(reviewOut, "Done") {
		t.Fatalf("unexpected review preset output: %s", reviewOut)
	}

	boardOut, err := executeCLI(t, "ls", "--root", root, "--preset", "board")
	if err != nil {
		t.Fatalf("ls --preset board failed: %v", err)
	}
	for _, want := range []string{"open:", "done:", "Open", "Done"} {
		if !strings.Contains(boardOut, want) {
			t.Fatalf("expected board preset output to contain %q, got:\n%s", want, boardOut)
		}
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
