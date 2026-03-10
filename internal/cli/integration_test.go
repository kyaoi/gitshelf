package cli

import (
	"bytes"
	"encoding/csv"
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

func TestCLIShowTSVFieldsIncludeBodyAndFile(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title: "Task",
		Body:  "first line\nsecond line",
		Tags:  []string{"focus"},
	})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	out, err := executeCLI(t, "show", "--root", root, task.ID, "--format", "tsv", "--fields", "id,title,body,file,tags")
	if err != nil {
		t.Fatalf("show --format tsv failed: %v", err)
	}
	fields := strings.Split(strings.TrimSpace(out), "\t")
	if len(fields) != 5 {
		t.Fatalf("expected 5 columns, got %d: %q", len(fields), out)
	}
	if fields[0] != task.ID || fields[1] != "Task" || fields[2] != "first line second line" || fields[3] != filepath.Join(shelf.TasksDir(root), task.ID+".md") || fields[4] != "focus" {
		t.Fatalf("unexpected show tsv fields: %#v", fields)
	}
}

func TestCLIShowCSVFormatIncludesHeaderAndQuotedBody(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Task", Body: "line 1\nline 2"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	out, err := executeCLI(t, "show", "--root", root, task.ID, "--format", "csv")
	if err != nil {
		t.Fatalf("show --format csv failed: %v", err)
	}
	rows, err := csv.NewReader(bytes.NewBufferString(out)).ReadAll()
	if err != nil {
		t.Fatalf("parse csv failed: %v\n%s", err, out)
	}
	if len(rows) != 2 {
		t.Fatalf("expected header + row, got %d: %#v", len(rows), rows)
	}
	if rows[0][0] != "id" || rows[1][0] != task.ID || rows[1][11] != "line 1\nline 2" {
		t.Fatalf("unexpected show csv rows: %#v", rows)
	}
}

func TestCLIShowCSVFieldsAndNoHeader(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Task", Tags: []string{"focus"}})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	out, err := executeCLI(t, "show", "--root", root, task.ID, "--format", "csv", "--fields", "title,file,tags", "--no-header")
	if err != nil {
		t.Fatalf("show --format csv failed: %v", err)
	}
	rows, err := csv.NewReader(bytes.NewBufferString(out)).ReadAll()
	if err != nil {
		t.Fatalf("parse csv failed: %v\n%s", err, out)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d: %#v", len(rows), rows)
	}
	if rows[0][0] != "Task" || rows[0][1] != filepath.Join(shelf.TasksDir(root), task.ID+".md") || rows[0][2] != "focus" {
		t.Fatalf("unexpected show csv rows: %#v", rows)
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

func TestCLILinksJSONIncludesPathAndFile(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	from, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "From", Parent: parent.ID})
	if err != nil {
		t.Fatalf("add from failed: %v", err)
	}
	to, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "To"})
	if err != nil {
		t.Fatalf("add to failed: %v", err)
	}
	if err := shelf.LinkTasks(root, from.ID, to.ID, "depends_on"); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	out, err := executeCLI(t, "links", "--root", root, from.ID, "--json")
	if err != nil {
		t.Fatalf("links --json failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("parse links json failed: %v\n%s", err, out)
	}
	taskPayload, ok := payload["task"].(map[string]any)
	if !ok {
		t.Fatalf("missing task payload: %#v", payload)
	}
	if taskPayload["path"] != "root > Parent > From" || taskPayload["file"] != filepath.Join(shelf.TasksDir(root), from.ID+".md") {
		t.Fatalf("unexpected task payload: %#v", taskPayload)
	}
	outbound, ok := payload["outbound"].([]any)
	if !ok || len(outbound) != 1 {
		t.Fatalf("unexpected outbound payload: %#v", payload["outbound"])
	}
	first, ok := outbound[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected outbound item: %#v", outbound[0])
	}
	if first["path"] != "root > To" || first["file"] != filepath.Join(shelf.TasksDir(root), to.ID+".md") {
		t.Fatalf("unexpected outbound item: %#v", first)
	}
}

func TestCLILinksTSVFieldsIncludeDirectionAndPaths(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	from, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "From"})
	if err != nil {
		t.Fatalf("add from failed: %v", err)
	}
	to, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "To"})
	if err != nil {
		t.Fatalf("add to failed: %v", err)
	}
	peer, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Peer"})
	if err != nil {
		t.Fatalf("add peer failed: %v", err)
	}
	if err := shelf.LinkTasks(root, from.ID, to.ID, "depends_on"); err != nil {
		t.Fatalf("add outbound link failed: %v", err)
	}
	if err := shelf.LinkTasks(root, peer.ID, from.ID, "related"); err != nil {
		t.Fatalf("add inbound link failed: %v", err)
	}

	out, err := executeCLI(t, "links", "--root", root, from.ID, "--format", "tsv", "--fields", "direction,type,other_id,other_path")
	if err != nil {
		t.Fatalf("links --format tsv failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 tsv rows, got %d: %q", len(lines), out)
	}
	if lines[0] != strings.Join([]string{"outbound", "depends_on", to.ID, "root > To"}, "\t") {
		t.Fatalf("unexpected outbound row: %q", lines[0])
	}
	if lines[1] != strings.Join([]string{"inbound", "related", peer.ID, "root > Peer"}, "\t") {
		t.Fatalf("unexpected inbound row: %q", lines[1])
	}
}

func TestCLILinksJSONLFormatUsesOneEdgePerLine(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	from, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "From"})
	if err != nil {
		t.Fatalf("add from failed: %v", err)
	}
	to, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "To"})
	if err != nil {
		t.Fatalf("add to failed: %v", err)
	}
	if err := shelf.LinkTasks(root, from.ID, to.ID, "depends_on"); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	out, err := executeCLI(t, "links", "--root", root, from.ID, "--format", "jsonl")
	if err != nil {
		t.Fatalf("links --format jsonl failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 edge line, got %d: %q", len(lines), out)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &payload); err != nil {
		t.Fatalf("parse jsonl failed: %v\n%s", err, lines[0])
	}
	if payload["direction"] != "outbound" || payload["type"] != "depends_on" {
		t.Fatalf("unexpected edge payload: %#v", payload)
	}
}

func TestCLILinksCSVFieldsAndHeader(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	from, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "From"})
	if err != nil {
		t.Fatalf("add from failed: %v", err)
	}
	to, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "To"})
	if err != nil {
		t.Fatalf("add to failed: %v", err)
	}
	if err := shelf.LinkTasks(root, from.ID, to.ID, "depends_on"); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	out, err := executeCLI(t, "links", "--root", root, from.ID, "--format", "csv", "--fields", "direction,type,other_file", "--header")
	if err != nil {
		t.Fatalf("links --format csv failed: %v", err)
	}
	rows, err := csv.NewReader(bytes.NewBufferString(out)).ReadAll()
	if err != nil {
		t.Fatalf("parse csv failed: %v\n%s", err, out)
	}
	if len(rows) != 2 {
		t.Fatalf("expected header + row, got %d: %#v", len(rows), rows)
	}
	if rows[0][0] != "direction" || rows[0][1] != "type" || rows[0][2] != "other_file" {
		t.Fatalf("unexpected header row: %#v", rows[0])
	}
	if rows[1][0] != "outbound" || rows[1][1] != "depends_on" || rows[1][2] != filepath.Join(shelf.TasksDir(root), to.ID+".md") {
		t.Fatalf("unexpected data row: %#v", rows[1])
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

func TestCLINextCSVHeaderOption(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Ready", Status: "open"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	out, err := executeCLI(t, "next", "--root", root, "--format", "csv", "--fields", "title,file", "--no-header")
	if err != nil {
		t.Fatalf("next --format csv failed: %v", err)
	}
	rows, err := csv.NewReader(bytes.NewBufferString(out)).ReadAll()
	if err != nil {
		t.Fatalf("parse csv failed: %v\n%s", err, out)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d: %#v", len(rows), rows)
	}
	if rows[0][0] != "Ready" || rows[0][1] != filepath.Join(shelf.TasksDir(root), task.ID+".md") {
		t.Fatalf("unexpected next csv row: %#v", rows[0])
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

func TestCLILsCSVFormatIncludesHeaderAndRows(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Task, One", Tags: []string{"focus"}})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	out, err := executeCLI(t, "ls", "--root", root, "--format", "csv", "--search", "Task")
	if err != nil {
		t.Fatalf("ls --format csv failed: %v", err)
	}
	rows, err := csv.NewReader(bytes.NewBufferString(out)).ReadAll()
	if err != nil {
		t.Fatalf("parse csv failed: %v\n%s", err, out)
	}
	if len(rows) != 2 {
		t.Fatalf("expected header + 1 row, got %d: %#v", len(rows), rows)
	}
	if rows[0][0] != "id" || rows[0][1] != "title" || rows[1][0] != task.ID || rows[1][1] != "Task, One" {
		t.Fatalf("unexpected csv rows: %#v", rows)
	}
}

func TestCLILsCSVFieldsAndNoHeader(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Task", Tags: []string{"focus"}})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	out, err := executeCLI(t, "ls", "--root", root, "--format", "csv", "--fields", "title,file,tags", "--no-header")
	if err != nil {
		t.Fatalf("ls --format csv failed: %v", err)
	}
	rows, err := csv.NewReader(bytes.NewBufferString(out)).ReadAll()
	if err != nil {
		t.Fatalf("parse csv failed: %v\n%s", err, out)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d: %#v", len(rows), rows)
	}
	if rows[0][0] != "Task" || rows[0][1] != filepath.Join(shelf.TasksDir(root), task.ID+".md") || rows[0][2] != "focus" {
		t.Fatalf("unexpected csv rows: %#v", rows)
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

func TestCLILsTSVFormatUsesStableColumns(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:       "Child\tTask",
		Parent:      parent.ID,
		Kind:        "todo",
		Status:      "open",
		DueOn:       "2026-03-20",
		RepeatEvery: "1w",
		Tags:        []string{"backend", "cli"},
	})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	out, err := executeCLI(t, "ls", "--root", root, "--format", "tsv", "--search", "Child")
	if err != nil {
		t.Fatalf("ls --format tsv failed: %v", err)
	}
	fields := strings.Split(strings.TrimSpace(out), "\t")
	if len(fields) != 12 {
		t.Fatalf("expected 12 tsv columns, got %d: %q", len(fields), out)
	}
	if fields[0] != task.ID || fields[1] != "Child Task" || fields[2] != "root > Parent > Child Task" {
		t.Fatalf("unexpected tsv columns: %#v", fields)
	}
	if fields[8] != parent.ID || fields[9] != "root > Parent" || fields[10] != "backend,cli" || fields[11] != filepath.Join(shelf.TasksDir(root), task.ID+".md") {
		t.Fatalf("unexpected tsv tail columns: %#v", fields)
	}
}

func TestCLINextTSVFormatUsesStableColumns(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title: "Ready",
		Kind:  "todo",
		Tags:  []string{"focus"},
	})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	out, err := executeCLI(t, "next", "--root", root, "--format", "tsv")
	if err != nil {
		t.Fatalf("next --format tsv failed: %v", err)
	}
	fields := strings.Split(strings.TrimSpace(out), "\t")
	if len(fields) != 11 {
		t.Fatalf("expected 11 tsv columns, got %d: %q", len(fields), out)
	}
	if fields[0] != task.ID || fields[1] != "Ready" || fields[2] != "root > Ready" || fields[10] != filepath.Join(shelf.TasksDir(root), task.ID+".md") {
		t.Fatalf("unexpected next tsv columns: %#v", fields)
	}
}

func TestCLILsTSVFieldsSelectAndReorderColumns(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Task", Tags: []string{"focus"}})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	out, err := executeCLI(t, "ls", "--root", root, "--format", "tsv", "--fields", "title,id,file,tags", "--search", "Task")
	if err != nil {
		t.Fatalf("ls --fields failed: %v", err)
	}
	fields := strings.Split(strings.TrimSpace(out), "\t")
	if len(fields) != 4 {
		t.Fatalf("expected 4 columns, got %d: %q", len(fields), out)
	}
	if fields[0] != "Task" || fields[1] != task.ID || fields[2] != filepath.Join(shelf.TasksDir(root), task.ID+".md") || fields[3] != "focus" {
		t.Fatalf("unexpected selected fields: %#v", fields)
	}
}

func TestCLINextTSVFieldsRejectUnknownField(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Task"}); err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	if _, err := executeCLI(t, "next", "--root", root, "--format", "tsv", "--fields", "id,body"); err == nil || !strings.Contains(err.Error(), "unknown --fields entry: body") {
		t.Fatalf("expected unknown field error, got: %v", err)
	}
}

func TestCLINextJSONLFormatUsesOneRecordPerLine(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Ready"}); err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	out, err := executeCLI(t, "next", "--root", root, "--format", "jsonl")
	if err != nil {
		t.Fatalf("next --format jsonl failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 jsonl line, got %d: %q", len(lines), out)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &payload); err != nil {
		t.Fatalf("parse jsonl failed: %v\n%s", err, lines[0])
	}
	if payload["title"] != "Ready" {
		t.Fatalf("unexpected jsonl payload: %#v", payload)
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
