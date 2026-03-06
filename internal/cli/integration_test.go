package cli

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	if !strings.Contains(showOutput, "Context Tree:") || !strings.Contains(showOutput, "Child (todo/open)") {
		t.Fatalf("show should include context tree output: %s", showOutput)
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

func TestCLIAddSetAndShowDueOn(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	addOut, err := executeCLI(t, "add", "--root", root, "--title", "due task", "--due", "2026-04-01")
	if err != nil {
		t.Fatalf("add with due failed: %v", err)
	}
	id := extractIDFromAddOutput(addOut)
	if id == "" {
		t.Fatalf("failed to parse task id from add output: %s", addOut)
	}

	showOut, err := executeCLI(t, "show", "--root", root, id)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if !strings.Contains(showOut, `due_on = "2026-04-01"`) {
		t.Fatalf("show should include due_on: %s", showOut)
	}

	if _, err := executeCLI(t, "set", "--root", root, id, "--due", "2026-04-05"); err != nil {
		t.Fatalf("set due failed: %v", err)
	}
	showOut, err = executeCLI(t, "show", "--root", root, id)
	if err != nil {
		t.Fatalf("show after set due failed: %v", err)
	}
	if !strings.Contains(showOut, `due_on = "2026-04-05"`) {
		t.Fatalf("show should include updated due_on: %s", showOut)
	}

	if _, err := executeCLI(t, "set", "--root", root, id, "--clear-due"); err != nil {
		t.Fatalf("clear due failed: %v", err)
	}
	showOut, err = executeCLI(t, "show", "--root", root, id)
	if err != nil {
		t.Fatalf("show after clear due failed: %v", err)
	}
	if strings.Contains(showOut, "due_on =") {
		t.Fatalf("show should not include due_on after clear: %s", showOut)
	}

	if _, err := executeCLI(t, "set", "--root", root, id, "--due", "2026-04-10", "--clear-due"); err == nil || !strings.Contains(err.Error(), "同時に指定できません") {
		t.Fatalf("expected due/clear-due conflict error, got: %v", err)
	}
}

func TestCLIAddAndSetDueKeywords(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	addOut, err := executeCLI(t, "add", "--root", root, "--title", "keyword due", "--due", "tomorrow")
	if err != nil {
		t.Fatalf("add with tomorrow failed: %v", err)
	}
	id := extractIDFromAddOutput(addOut)
	wantTomorrow := time.Now().Local().AddDate(0, 0, 1).Format("2006-01-02")

	showOut, err := executeCLI(t, "show", "--root", root, id)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if !strings.Contains(showOut, `due_on = "`+wantTomorrow+`"`) {
		t.Fatalf("show should contain normalized tomorrow due_on: %s", showOut)
	}

	if _, err := executeCLI(t, "set", "--root", root, id, "--due", "today"); err != nil {
		t.Fatalf("set due today failed: %v", err)
	}
	wantToday := time.Now().Local().Format("2006-01-02")
	showOut, err = executeCLI(t, "show", "--root", root, id)
	if err != nil {
		t.Fatalf("show after set today failed: %v", err)
	}
	if !strings.Contains(showOut, `due_on = "`+wantToday+`"`) {
		t.Fatalf("show should contain normalized today due_on: %s", showOut)
	}

	if _, err := executeCLI(t, "set", "--root", root, id, "--due", "+2d"); err != nil {
		t.Fatalf("set due +2d failed: %v", err)
	}
	wantPlus2 := time.Now().Local().AddDate(0, 0, 2).Format("2006-01-02")
	showOut, err = executeCLI(t, "show", "--root", root, id)
	if err != nil {
		t.Fatalf("show after set +2d failed: %v", err)
	}
	if !strings.Contains(showOut, `due_on = "`+wantPlus2+`"`) {
		t.Fatalf("show should contain normalized +2d due_on: %s", showOut)
	}
}

func TestCLIAddAndSetRepeatEvery(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	addOut, err := executeCLI(t, "add", "--root", root, "--title", "repeat task", "--repeat-every", "1w")
	if err != nil {
		t.Fatalf("add with repeat failed: %v", err)
	}
	id := extractIDFromAddOutput(addOut)
	showOut, err := executeCLI(t, "show", "--root", root, id)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if !strings.Contains(showOut, `repeat_every = "1w"`) {
		t.Fatalf("show should include repeat_every: %s", showOut)
	}

	if _, err := executeCLI(t, "set", "--root", root, id, "--repeat-every", "2w"); err != nil {
		t.Fatalf("set repeat failed: %v", err)
	}
	showOut, err = executeCLI(t, "show", "--root", root, id)
	if err != nil {
		t.Fatalf("show after set repeat failed: %v", err)
	}
	if !strings.Contains(showOut, `repeat_every = "2w"`) {
		t.Fatalf("show should include updated repeat_every: %s", showOut)
	}

	if _, err := executeCLI(t, "set", "--root", root, id, "--clear-repeat"); err != nil {
		t.Fatalf("clear repeat failed: %v", err)
	}
	showOut, err = executeCLI(t, "show", "--root", root, id)
	if err != nil {
		t.Fatalf("show after clear repeat failed: %v", err)
	}
	if strings.Contains(showOut, "repeat_every =") {
		t.Fatalf("show should not include repeat_every after clear: %s", showOut)
	}

	if _, err := executeCLI(t, "set", "--root", root, id, "--repeat-every", "bad"); err == nil || !strings.Contains(err.Error(), "invalid repeat_every") {
		t.Fatalf("expected invalid repeat error, got: %v", err)
	}
	if _, err := executeCLI(t, "set", "--root", root, id, "--repeat-every", "1w", "--clear-repeat"); err == nil || !strings.Contains(err.Error(), "同時に指定できません") {
		t.Fatalf("expected repeat/clear conflict error, got: %v", err)
	}
}

func TestCLINextShowsOnlyReadyTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	a, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "A depends on B"})
	if err != nil {
		t.Fatalf("add A failed: %v", err)
	}
	b, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "B prerequisite"})
	if err != nil {
		t.Fatalf("add B failed: %v", err)
	}
	_, err = shelf.AddTask(root, shelf.AddTaskInput{Title: "C independent"})
	if err != nil {
		t.Fatalf("add C failed: %v", err)
	}
	if err := shelf.LinkTasks(root, a.ID, b.ID, "depends_on"); err != nil {
		t.Fatalf("link A->B failed: %v", err)
	}

	nextOut, err := executeCLI(t, "next", "--root", root)
	if err != nil {
		t.Fatalf("next failed: %v", err)
	}
	if strings.Contains(nextOut, "A depends on B") {
		t.Fatalf("A should not be listed while dependency is open: %s", nextOut)
	}
	if !strings.Contains(nextOut, "B prerequisite") || !strings.Contains(nextOut, "C independent") {
		t.Fatalf("expected B and C in next output: %s", nextOut)
	}

	done := shelf.Status("done")
	if _, err := shelf.SetTask(root, b.ID, shelf.SetTaskInput{Status: &done}); err != nil {
		t.Fatalf("set B done failed: %v", err)
	}
	nextOut, err = executeCLI(t, "next", "--root", root)
	if err != nil {
		t.Fatalf("next after done failed: %v", err)
	}
	if !strings.Contains(nextOut, "A depends on B") || !strings.Contains(nextOut, "C independent") {
		t.Fatalf("expected A and C after B done: %s", nextOut)
	}
	if strings.Contains(nextOut, "B prerequisite") {
		t.Fatalf("done task B should not be listed in next: %s", nextOut)
	}

	nextJSON, err := executeCLI(t, "next", "--root", root, "--json")
	if err != nil {
		t.Fatalf("next --json failed: %v", err)
	}
	var rows []map[string]any
	if err := json.Unmarshal([]byte(nextJSON), &rows); err != nil {
		t.Fatalf("failed to parse next json output: %v output=%s", err, nextJSON)
	}
	if len(rows) == 0 {
		t.Fatalf("expected ready tasks in next json output: %s", nextJSON)
	}
}

func TestCLIStatusShortcutCommands(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "status task"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	if _, err := executeCLI(t, "start", "--root", root, task.ID); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	showOut, err := executeCLI(t, "show", "--root", root, task.ID)
	if err != nil {
		t.Fatalf("show after start failed: %v", err)
	}
	if !strings.Contains(showOut, `status = "in_progress"`) {
		t.Fatalf("expected in_progress status: %s", showOut)
	}

	if _, err := executeCLI(t, "block", "--root", root, task.ID); err != nil {
		t.Fatalf("block failed: %v", err)
	}
	showOut, err = executeCLI(t, "show", "--root", root, task.ID)
	if err != nil {
		t.Fatalf("show after block failed: %v", err)
	}
	if !strings.Contains(showOut, `status = "blocked"`) {
		t.Fatalf("expected blocked status: %s", showOut)
	}

	if _, err := executeCLI(t, "cancel", "--root", root, task.ID); err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	showOut, err = executeCLI(t, "show", "--root", root, task.ID)
	if err != nil {
		t.Fatalf("show after cancel failed: %v", err)
	}
	if !strings.Contains(showOut, `status = "cancelled"`) {
		t.Fatalf("expected cancelled status: %s", showOut)
	}
}

func TestCLITreeFiltersAndJSON(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "ChildTodo", Kind: "todo", Status: "in_progress", Parent: parent.ID}); err != nil {
		t.Fatalf("add child todo failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "ChildMemoDone", Kind: "memo", Status: "done", Parent: parent.ID}); err != nil {
		t.Fatalf("add child memo done failed: %v", err)
	}

	out, err := executeCLI(t, "tree", "--root", root, "--kind", "todo", "--not-status", "done")
	if err != nil {
		t.Fatalf("tree filter failed: %v", err)
	}
	if !strings.Contains(out, "Parent") || !strings.Contains(out, "ChildTodo") {
		t.Fatalf("expected todo nodes in filtered tree: %s", out)
	}
	if strings.Contains(out, "ChildMemoDone") {
		t.Fatalf("done memo node should be filtered out: %s", out)
	}

	out, err = executeCLI(t, "tree", "--root", root, "--json")
	if err != nil {
		t.Fatalf("tree --json failed: %v", err)
	}
	var rows []map[string]any
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		t.Fatalf("failed to parse tree json output: %v output=%s", err, out)
	}
	if len(rows) == 0 {
		t.Fatalf("expected json tree nodes, got empty: %s", out)
	}

	if _, err := executeCLI(t, "tree", "--root", root, "--kind", "unknown"); err == nil || !strings.Contains(err.Error(), "unknown kind") {
		t.Fatalf("expected unknown kind error, got: %v", err)
	}
	if _, err := executeCLI(t, "tree", "--root", root, "--status", "unknown"); err == nil || !strings.Contains(err.Error(), "unknown status") {
		t.Fatalf("expected unknown status error, got: %v", err)
	}
}

func TestCLIShowBodyModesAndJSON(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "show target", Body: "line1\nline2"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	out, err := executeCLI(t, "show", "--root", root, task.ID, "--only-body")
	if err != nil {
		t.Fatalf("show --only-body failed: %v", err)
	}
	if strings.Contains(out, "+++") || !strings.Contains(out, "line1") {
		t.Fatalf("only-body output should contain only body text: %s", out)
	}

	out, err = executeCLI(t, "show", "--root", root, task.ID, "--no-body")
	if err != nil {
		t.Fatalf("show --no-body failed: %v", err)
	}
	if strings.Contains(out, "line1") || !strings.Contains(out, "Hierarchy:") {
		t.Fatalf("no-body output should hide body and keep metadata/hierarchy: %s", out)
	}

	out, err = executeCLI(t, "show", "--root", root, task.ID, "--json")
	if err != nil {
		t.Fatalf("show --json failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("failed to parse show json output: %v output=%s", err, out)
	}
	if _, ok := payload["task"]; !ok {
		t.Fatalf("show json should include task object: %s", out)
	}

	if _, err := executeCLI(t, "show", "--root", root, task.ID, "--no-body", "--only-body"); err == nil || !strings.Contains(err.Error(), "同時に指定できません") {
		t.Fatalf("expected no-body/only-body conflict error, got: %v", err)
	}
}

func TestCLILinksTransitiveAndJSON(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	a, _ := shelf.AddTask(root, shelf.AddTaskInput{Title: "A"})
	b, _ := shelf.AddTask(root, shelf.AddTaskInput{Title: "B"})
	c, _ := shelf.AddTask(root, shelf.AddTaskInput{Title: "C"})
	if err := shelf.LinkTasks(root, a.ID, b.ID, "depends_on"); err != nil {
		t.Fatalf("link A->B failed: %v", err)
	}
	if err := shelf.LinkTasks(root, b.ID, c.ID, "depends_on"); err != nil {
		t.Fatalf("link B->C failed: %v", err)
	}

	out, err := executeCLI(t, "links", "--root", root, a.ID, "--transitive")
	if err != nil {
		t.Fatalf("links --transitive failed: %v", err)
	}
	if !strings.Contains(out, "Transitive depends_on:") {
		t.Fatalf("expected transitive section in links output: %s", out)
	}

	out, err = executeCLI(t, "links", "--root", root, a.ID, "--transitive", "--json")
	if err != nil {
		t.Fatalf("links --json failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("failed to parse links json output: %v output=%s", err, out)
	}
	if _, ok := payload["transitive_depends_on"]; !ok {
		t.Fatalf("expected transitive_depends_on in json output: %s", out)
	}
}

func TestCLIDoctorFixAndJSON(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "A"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}
	dep, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "B"})
	if err != nil {
		t.Fatalf("add dep failed: %v", err)
	}

	edgePath := filepath.Join(root, ".shelf", "edges", task.ID+".toml")
	edgeData := `[[edge]]
to = "` + dep.ID + `"
type = "depends_on"

[[edge]]
to = "` + dep.ID + `"
type = "depends_on"
`
	if err := os.WriteFile(edgePath, []byte(edgeData), 0o644); err != nil {
		t.Fatalf("write edge file failed: %v", err)
	}

	out, err := executeCLI(t, "doctor", "--root", root, "--fix", "--json")
	if err != nil {
		t.Fatalf("doctor --fix --json should pass after fix, got: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("failed to parse doctor json output: %v output=%s", err, out)
	}
	if payload["ok"] != true {
		t.Fatalf("doctor json should report ok=true: %s", out)
	}
}

func TestCLIDoctorShowsAdvice(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "bad status"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}
	taskPath := filepath.Join(root, ".shelf", "tasks", task.ID+".md")
	data, err := os.ReadFile(taskPath)
	if err != nil {
		t.Fatalf("read task failed: %v", err)
	}
	corrupt := strings.Replace(string(data), `status = "open"`, `status = "oops"`, 1)
	if err := os.WriteFile(taskPath, []byte(corrupt), 0o644); err != nil {
		t.Fatalf("write corrupt task failed: %v", err)
	}

	out, err := executeCLI(t, "doctor", "--root", root)
	if err == nil {
		t.Fatalf("doctor should report error for bad status")
	}
	if !strings.Contains(out, "hint:") || !strings.Contains(out, "statuses") {
		t.Fatalf("doctor output should include advice hint: %s", out)
	}

	out, err = executeCLI(t, "doctor", "--root", root, "--json")
	if err == nil {
		t.Fatalf("doctor --json should report error for bad status")
	}
	if idx := strings.Index(out, "\nError:"); idx >= 0 {
		out = out[:idx]
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("failed to parse doctor json: %v output=%s", err, out)
	}
	issues, ok := payload["issues"].([]any)
	if !ok || len(issues) == 0 {
		t.Fatalf("doctor json should include issues: %s", out)
	}
	first, ok := issues[0].(map[string]any)
	if !ok {
		t.Fatalf("doctor json issue format mismatch: %s", out)
	}
	if _, ok := first["advice"]; !ok {
		t.Fatalf("doctor json issue should include advice: %s", out)
	}
}

func TestCLILsReadinessDueFiltersAndJSON(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	today := time.Now().Local().Format("2006-01-02")
	yesterday := time.Now().Local().AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := time.Now().Local().AddDate(0, 0, 1).Format("2006-01-02")

	dep, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "dep", DueOn: yesterday})
	if err != nil {
		t.Fatalf("add dep failed: %v", err)
	}
	blocked, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "blocked", DueOn: tomorrow})
	if err != nil {
		t.Fatalf("add blocked failed: %v", err)
	}
	if err := shelf.LinkTasks(root, blocked.ID, dep.ID, "depends_on"); err != nil {
		t.Fatalf("link blocked->dep failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "independent", DueOn: tomorrow}); err != nil {
		t.Fatalf("add independent failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "nodue"}); err != nil {
		t.Fatalf("add nodue failed: %v", err)
	}

	out, err := executeCLI(t, "ls", "--root", root, "--ready")
	if err != nil {
		t.Fatalf("ls --ready failed: %v", err)
	}
	if strings.Contains(out, "blocked") {
		t.Fatalf("blocked task should not be ready: %s", out)
	}

	out, err = executeCLI(t, "ls", "--root", root, "--blocked-by-deps")
	if err != nil {
		t.Fatalf("ls --blocked-by-deps failed: %v", err)
	}
	if !strings.Contains(out, "blocked") || strings.Contains(out, "independent") {
		t.Fatalf("unexpected blocked-by-deps result: %s", out)
	}

	out, err = executeCLI(t, "ls", "--root", root, "--overdue")
	if err != nil {
		t.Fatalf("ls --overdue failed: %v", err)
	}
	if !strings.Contains(out, "dep") || strings.Contains(out, "independent") {
		t.Fatalf("unexpected overdue result: %s", out)
	}

	out, err = executeCLI(t, "ls", "--root", root, "--due-before", today)
	if err != nil {
		t.Fatalf("ls --due-before failed: %v", err)
	}
	if !strings.Contains(out, "dep") || strings.Contains(out, "independent") {
		t.Fatalf("unexpected due-before result: %s", out)
	}

	out, err = executeCLI(t, "ls", "--root", root, "--due-after", today)
	if err != nil {
		t.Fatalf("ls --due-after failed: %v", err)
	}
	if !strings.Contains(out, "independent") || !strings.Contains(out, "blocked") || strings.Contains(out, "dep  (") {
		t.Fatalf("unexpected due-after result: %s", out)
	}

	out, err = executeCLI(t, "ls", "--root", root, "--no-due")
	if err != nil {
		t.Fatalf("ls --no-due failed: %v", err)
	}
	if !strings.Contains(out, "nodue") || strings.Contains(out, "independent") {
		t.Fatalf("unexpected no-due result: %s", out)
	}

	out, err = executeCLI(t, "ls", "--root", root, "--json", "--ready")
	if err != nil {
		t.Fatalf("ls --json failed: %v", err)
	}
	var rows []map[string]any
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		t.Fatalf("failed to parse ls json output: %v output=%s", err, out)
	}
	if len(rows) == 0 {
		t.Fatalf("expected json rows, got empty: %s", out)
	}

	if _, err := executeCLI(t, "ls", "--root", root, "--ready", "--blocked-by-deps"); err == nil || !strings.Contains(err.Error(), "同時に指定できません") {
		t.Fatalf("expected ready/deps conflict error, got: %v", err)
	}
	if _, err := executeCLI(t, "ls", "--root", root, "--due-before", "2026-99-01"); err == nil || !strings.Contains(err.Error(), "invalid due_on") {
		t.Fatalf("expected invalid due_on error, got: %v", err)
	}
}

func TestCLIViewPresetsForLsTreeNext(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "ChildDone", Kind: "todo", Status: "done", Parent: parent.ID}); err != nil {
		t.Fatalf("add done child failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "ChildOpen", Kind: "todo", Status: "open", Parent: parent.ID}); err != nil {
		t.Fatalf("add open child failed: %v", err)
	}

	out, err := executeCLI(t, "ls", "--root", root, "--view", "active")
	if err != nil {
		t.Fatalf("ls --view active failed: %v", err)
	}
	if strings.Contains(out, "ChildDone") || !strings.Contains(out, "ChildOpen") {
		t.Fatalf("unexpected ls active output: %s", out)
	}

	out, err = executeCLI(t, "tree", "--root", root, "--view", "active")
	if err != nil {
		t.Fatalf("tree --view active failed: %v", err)
	}
	if strings.Contains(out, "ChildDone") || !strings.Contains(out, "ChildOpen") {
		t.Fatalf("unexpected tree active output: %s", out)
	}

	if _, err := executeCLI(t, "tree", "--root", root, "--view", "ready"); err == nil || !strings.Contains(err.Error(), "not supported for tree") {
		t.Fatalf("expected unsupported tree view error, got: %v", err)
	}

	out, err = executeCLI(t, "next", "--root", root, "--view", "active")
	if err != nil {
		t.Fatalf("next --view active failed: %v", err)
	}
	if strings.Contains(out, "ChildDone") {
		t.Fatalf("next active should not include done tasks: %s", out)
	}

	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	cfg.Views["only_done"] = shelf.TaskView{
		Statuses: []shelf.Status{"done"},
	}
	if err := shelf.SaveConfig(root, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	out, err = executeCLI(t, "ls", "--root", root, "--view", "only_done")
	if err != nil {
		t.Fatalf("ls --view only_done failed: %v", err)
	}
	if !strings.Contains(out, "ChildDone") || strings.Contains(out, "ChildOpen") {
		t.Fatalf("unexpected custom view output: %s", out)
	}
}

func TestCLIViewManageCommands(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	listOut, err := executeCLI(t, "view", "--root", root, "list")
	if err != nil {
		t.Fatalf("view list failed: %v", err)
	}
	if !strings.Contains(listOut, "active") || !strings.Contains(listOut, "ready") {
		t.Fatalf("view list should contain built-ins: %s", listOut)
	}

	if _, err := executeCLI(t, "view", "--root", root, "set", "todo_open", "--kind", "todo", "--status", "open", "--limit", "20"); err != nil {
		t.Fatalf("view set failed: %v", err)
	}
	if _, err := executeCLI(t, "view", "--root", root, "copy", "todo_open", "todo_open_copy"); err != nil {
		t.Fatalf("view copy failed: %v", err)
	}
	if _, err := executeCLI(t, "view", "--root", root, "rename", "todo_open_copy", "todo_open_renamed"); err != nil {
		t.Fatalf("view rename failed: %v", err)
	}
	if _, err := executeCLI(t, "view", "--root", root, "set", "progress", "--status", "in_progress"); err != nil {
		t.Fatalf("view set progress failed: %v", err)
	}
	if _, err := executeCLI(t, "view", "--root", root, "merge", "todo_or_progress", "--from", "todo_open", "--from", "progress", "--strategy", "union"); err != nil {
		t.Fatalf("view merge failed: %v", err)
	}

	showOut, err := executeCLI(t, "view", "--root", root, "show", "todo_open")
	if err != nil {
		t.Fatalf("view show failed: %v", err)
	}
	if !strings.Contains(showOut, "name: todo_open") || !strings.Contains(showOut, "limit: 20") {
		t.Fatalf("unexpected view show output: %s", showOut)
	}

	if _, err := executeCLI(t, "view", "--root", root, "set", "active", "--status", "open"); err == nil || !strings.Contains(err.Error(), "cannot overwrite built-in view") {
		t.Fatalf("expected built-in overwrite error, got: %v", err)
	}

	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "todo-open", Kind: "todo", Status: "open"}); err != nil {
		t.Fatalf("add todo-open failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "memo-open", Kind: "memo", Status: "open"}); err != nil {
		t.Fatalf("add memo-open failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "todo-progress", Kind: "todo", Status: "in_progress"}); err != nil {
		t.Fatalf("add todo-progress failed: %v", err)
	}

	lsOut, err := executeCLI(t, "ls", "--root", root, "--view", "todo_open")
	if err != nil {
		t.Fatalf("ls --view todo_open failed: %v", err)
	}
	if !strings.Contains(lsOut, "todo-open") || strings.Contains(lsOut, "memo-open") {
		t.Fatalf("unexpected filtered output: %s", lsOut)
	}
	lsOut, err = executeCLI(t, "ls", "--root", root, "--view", "todo_or_progress")
	if err != nil {
		t.Fatalf("ls --view todo_or_progress failed: %v", err)
	}
	if !strings.Contains(lsOut, "todo-open") || !strings.Contains(lsOut, "todo-progress") || strings.Contains(lsOut, "memo-open") {
		t.Fatalf("unexpected merged view output: %s", lsOut)
	}

	if _, err := executeCLI(t, "view", "--root", root, "delete", "todo_open"); err != nil {
		t.Fatalf("view delete failed: %v", err)
	}
	if _, err := executeCLI(t, "ls", "--root", root, "--view", "todo_open"); err == nil || !strings.Contains(err.Error(), "unknown view") {
		t.Fatalf("expected unknown view error after delete, got: %v", err)
	}
}

func TestCLIAgendaAndSnooze(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	today := time.Now().Local().Format("2006-01-02")
	yesterday := time.Now().Local().AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := time.Now().Local().AddDate(0, 0, 1).Format("2006-01-02")
	later := time.Now().Local().AddDate(0, 0, 20).Format("2006-01-02")

	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "overdue", DueOn: yesterday}); err != nil {
		t.Fatalf("add overdue failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "today", DueOn: today}); err != nil {
		t.Fatalf("add today failed: %v", err)
	}
	target, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "tomorrow", DueOn: tomorrow})
	if err != nil {
		t.Fatalf("add tomorrow failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "later", DueOn: later}); err != nil {
		t.Fatalf("add later failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "no-due"}); err != nil {
		t.Fatalf("add no-due failed: %v", err)
	}
	done := shelf.Status("done")
	doneTask, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "done-task", DueOn: today})
	if err != nil {
		t.Fatalf("add done task failed: %v", err)
	}
	if _, err := shelf.SetTask(root, doneTask.ID, shelf.SetTaskInput{Status: &done}); err != nil {
		t.Fatalf("set done failed: %v", err)
	}

	out, err := executeCLI(t, "agenda", "--root", root)
	if err != nil {
		t.Fatalf("agenda failed: %v", err)
	}
	if !strings.Contains(out, "Overdue:") || !strings.Contains(out, "Today:") || !strings.Contains(out, "Tomorrow:") {
		t.Fatalf("agenda sections missing: %s", out)
	}
	if strings.Contains(out, "done-task") {
		t.Fatalf("agenda should exclude done by default: %s", out)
	}

	out, err = executeCLI(t, "agenda", "--root", root, "--json")
	if err != nil {
		t.Fatalf("agenda --json failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("agenda json parse failed: %v output=%s", err, out)
	}
	if _, ok := payload["overdue"]; !ok {
		t.Fatalf("agenda json should contain overdue: %s", out)
	}

	if _, err := executeCLI(t, "snooze", "--root", root, target.ID, "--by", "2d"); err != nil {
		t.Fatalf("snooze --by failed: %v", err)
	}
	showOut, err := executeCLI(t, "show", "--root", root, target.ID)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	wantBy := time.Now().Local().AddDate(0, 0, 3).Format("2006-01-02")
	if !strings.Contains(showOut, `due_on = "`+wantBy+`"`) {
		t.Fatalf("snooze --by did not update due: %s", showOut)
	}

	if _, err := executeCLI(t, "snooze", "--root", root, target.ID, "--to", "today"); err != nil {
		t.Fatalf("snooze --to failed: %v", err)
	}
	showOut, err = executeCLI(t, "show", "--root", root, target.ID)
	if err != nil {
		t.Fatalf("show after --to failed: %v", err)
	}
	if !strings.Contains(showOut, `due_on = "`+today+`"`) {
		t.Fatalf("snooze --to did not set due: %s", showOut)
	}

	if _, err := executeCLI(t, "snooze", "--root", root, target.ID, "--by", "1d", "--to", "today"); err == nil || !strings.Contains(err.Error(), "どちらか一方") {
		t.Fatalf("expected by/to conflict error, got: %v", err)
	}
	if _, err := executeCLI(t, "snooze", "--root", root, target.ID, "--by", "x"); err == nil || !strings.Contains(err.Error(), "invalid --by") {
		t.Fatalf("expected invalid by error, got: %v", err)
	}
}

func TestCLIArchiveAndArchivedFilters(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	active, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "active"})
	if err != nil {
		t.Fatalf("add active failed: %v", err)
	}
	archived, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "archived"})
	if err != nil {
		t.Fatalf("add archived failed: %v", err)
	}

	if _, err := executeCLI(t, "archive", "--root", root, archived.ID); err != nil {
		t.Fatalf("archive failed: %v", err)
	}

	out, err := executeCLI(t, "ls", "--root", root)
	if err != nil {
		t.Fatalf("ls failed: %v", err)
	}
	if strings.Contains(out, "archived") || !strings.Contains(out, "active") {
		t.Fatalf("default ls should hide archived: %s", out)
	}

	out, err = executeCLI(t, "ls", "--root", root, "--include-archived")
	if err != nil {
		t.Fatalf("ls --include-archived failed: %v", err)
	}
	if !strings.Contains(out, "archived") || !strings.Contains(out, "active") {
		t.Fatalf("include-archived should show both: %s", out)
	}

	out, err = executeCLI(t, "ls", "--root", root, "--only-archived")
	if err != nil {
		t.Fatalf("ls --only-archived failed: %v", err)
	}
	if !strings.Contains(out, "archived") || strings.Contains(out, "active") {
		t.Fatalf("only-archived should show only archived tasks: %s", out)
	}

	if _, err := executeCLI(t, "ls", "--root", root, "--include-archived", "--only-archived"); err == nil || !strings.Contains(err.Error(), "同時に指定") {
		t.Fatalf("expected include/only conflict error, got: %v", err)
	}

	if _, err := executeCLI(t, "unarchive", "--root", root, archived.ID); err != nil {
		t.Fatalf("unarchive failed: %v", err)
	}
	showOut, err := executeCLI(t, "show", "--root", root, archived.ID)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if strings.Contains(showOut, "archived_at =") {
		t.Fatalf("expected archived_at cleared after unarchive: %s", showOut)
	}

	_ = active
}

func TestCLIDoneRecurringAndReopen(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	today := time.Now().Local().Format("2006-01-02")
	rec, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:       "recurring create",
		DueOn:       today,
		RepeatEvery: "1w",
	})
	if err != nil {
		t.Fatalf("add recurring task failed: %v", err)
	}

	if _, err := executeCLI(t, "done", "--root", root, rec.ID); err == nil || !strings.Contains(err.Error(), "--recurring-action") {
		t.Fatalf("expected recurring-action required error, got: %v", err)
	}

	if _, err := executeCLI(t, "done", "--root", root, rec.ID, "--recurring-action", "create"); err != nil {
		t.Fatalf("done recurring create failed: %v", err)
	}
	updated, err := shelf.EnsureTaskExists(root, rec.ID)
	if err != nil {
		t.Fatalf("get updated task failed: %v", err)
	}
	if updated.Status != "done" {
		t.Fatalf("expected original recurring task done, got: %s", updated.Status)
	}
	all, err := shelf.NewTaskStore(root).List()
	if err != nil {
		t.Fatalf("list tasks failed: %v", err)
	}
	foundNext := false
	for _, task := range all {
		if task.ID == rec.ID {
			continue
		}
		if task.Title == rec.Title && task.RepeatEvery == "1w" && task.Status == "open" {
			foundNext = true
		}
	}
	if !foundNext {
		t.Fatalf("expected next recurring task to be created: %+v", all)
	}

	rec2, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:       "recurring reopen",
		DueOn:       today,
		RepeatEvery: "1d",
	})
	if err != nil {
		t.Fatalf("add recurring reopen task failed: %v", err)
	}
	if _, err := executeCLI(t, "done", "--root", root, rec2.ID, "--recurring-action", "reopen"); err != nil {
		t.Fatalf("done recurring reopen failed: %v", err)
	}
	updated2, err := shelf.EnsureTaskExists(root, rec2.ID)
	if err != nil {
		t.Fatalf("get updated2 failed: %v", err)
	}
	if updated2.Status != "open" {
		t.Fatalf("expected reopened status open, got: %s", updated2.Status)
	}
	if updated2.DueOn == today {
		t.Fatalf("expected due_on to advance on reopen, got: %s", updated2.DueOn)
	}

	if _, err := executeCLI(t, "reopen", "--root", root, rec.ID); err != nil {
		t.Fatalf("reopen command failed: %v", err)
	}
	reopened, err := shelf.EnsureTaskExists(root, rec.ID)
	if err != nil {
		t.Fatalf("get reopened failed: %v", err)
	}
	if reopened.Status != "open" {
		t.Fatalf("expected reopen command to set open, got: %s", reopened.Status)
	}
}

func TestCLITodayAndFormatFlags(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	today := time.Now().Local().Format("2006-01-02")
	yesterday := time.Now().Local().AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := time.Now().Local().AddDate(0, 0, 1).Format("2006-01-02")
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "overdue", DueOn: yesterday}); err != nil {
		t.Fatalf("add overdue failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "today", DueOn: today}); err != nil {
		t.Fatalf("add today failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "tomorrow", DueOn: tomorrow}); err != nil {
		t.Fatalf("add tomorrow failed: %v", err)
	}

	out, err := executeCLI(t, "today", "--root", root)
	if err != nil {
		t.Fatalf("today failed: %v", err)
	}
	if !strings.Contains(out, "Overdue:") || !strings.Contains(out, "Today:") {
		t.Fatalf("today output should include sections: %s", out)
	}
	if strings.Contains(out, "tomorrow") {
		t.Fatalf("today output should not include tomorrow tasks: %s", out)
	}

	if _, err := executeCLI(t, "ls", "--root", root, "--format", "bad"); err == nil || !strings.Contains(err.Error(), "invalid --format") {
		t.Fatalf("expected invalid ls format error, got: %v", err)
	}
	if _, err := executeCLI(t, "tree", "--root", root, "--format", "bad"); err == nil || !strings.Contains(err.Error(), "invalid --format") {
		t.Fatalf("expected invalid tree format error, got: %v", err)
	}
	if _, err := executeCLI(t, "agenda", "--root", root, "--format", "bad"); err == nil || !strings.Contains(err.Error(), "invalid --format") {
		t.Fatalf("expected invalid agenda format error, got: %v", err)
	}
}

func TestCLITodayCarryOver(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	today := time.Now().Local().Format("2006-01-02")
	yesterday := time.Now().Local().AddDate(0, 0, -1).Format("2006-01-02")
	openTask, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "carry open", DueOn: yesterday, Status: "open"})
	if err != nil {
		t.Fatalf("add open task failed: %v", err)
	}
	doneTask, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "keep done", DueOn: yesterday, Status: "done"})
	if err != nil {
		t.Fatalf("add done task failed: %v", err)
	}

	if _, err := executeCLI(t, "today", "--root", root, "--carry-over"); err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("expected --yes required error in non-TTY, got: %v", err)
	}
	if _, err := executeCLI(t, "today", "--root", root, "--carry-over", "--yes"); err != nil {
		t.Fatalf("today --carry-over --yes failed: %v", err)
	}

	openAfter, err := shelf.EnsureTaskExists(root, openTask.ID)
	if err != nil {
		t.Fatalf("load open task failed: %v", err)
	}
	if openAfter.DueOn != today {
		t.Fatalf("open task due should move to today: got=%s want=%s", openAfter.DueOn, today)
	}
	doneAfter, err := shelf.EnsureTaskExists(root, doneTask.ID)
	if err != nil {
		t.Fatalf("load done task failed: %v", err)
	}
	if doneAfter.DueOn != yesterday {
		t.Fatalf("done task due should stay unchanged: got=%s want=%s", doneAfter.DueOn, yesterday)
	}
}

func TestCLIShowIncludesReadinessDetails(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	a, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "A"})
	if err != nil {
		t.Fatalf("add A failed: %v", err)
	}
	b, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "B"})
	if err != nil {
		t.Fatalf("add B failed: %v", err)
	}
	if err := shelf.LinkTasks(root, a.ID, b.ID, "depends_on"); err != nil {
		t.Fatalf("link A->B failed: %v", err)
	}

	out, err := executeCLI(t, "show", "--root", root, a.ID)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if !strings.Contains(out, "Readiness:") || !strings.Contains(out, "blocked_by_dependencies") || !strings.Contains(out, "B") {
		t.Fatalf("show should include readiness detail: %s", out)
	}
}

func TestCLIUndoRevertsLastMutatingAction(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if _, err := executeCLI(t, "undo", "--root", root); err == nil || !strings.Contains(err.Error(), "undo history is empty") {
		t.Fatalf("expected empty undo history error, got: %v", err)
	}

	addOut, err := executeCLI(t, "add", "--root", root, "--title", "undo add")
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}
	addedID := extractIDFromAddOutput(addOut)
	if addedID == "" {
		t.Fatalf("failed to parse add id: %s", addOut)
	}
	if _, err := shelf.EnsureTaskExists(root, addedID); err != nil {
		t.Fatalf("added task should exist: %v", err)
	}

	if _, err := executeCLI(t, "undo", "--root", root); err != nil {
		t.Fatalf("undo add failed: %v", err)
	}
	if _, err := shelf.EnsureTaskExists(root, addedID); err == nil {
		t.Fatalf("task should be removed after undo add")
	}

	a, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "undo set/link a"})
	if err != nil {
		t.Fatalf("add a failed: %v", err)
	}
	b, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "undo set/link b"})
	if err != nil {
		t.Fatalf("add b failed: %v", err)
	}

	if _, err := executeCLI(t, "set", "--root", root, a.ID, "--status", "done"); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	updated, err := shelf.EnsureTaskExists(root, a.ID)
	if err != nil {
		t.Fatalf("load updated failed: %v", err)
	}
	if updated.Status != "done" {
		t.Fatalf("expected done status after set, got: %s", updated.Status)
	}

	if _, err := executeCLI(t, "undo", "--root", root); err != nil {
		t.Fatalf("undo set failed: %v", err)
	}
	updated, err = shelf.EnsureTaskExists(root, a.ID)
	if err != nil {
		t.Fatalf("load reverted failed: %v", err)
	}
	if updated.Status != "open" {
		t.Fatalf("expected status open after undo, got: %s", updated.Status)
	}

	if _, err := executeCLI(t, "link", "--root", root, "--from", a.ID, "--to", b.ID, "--type", "related"); err != nil {
		t.Fatalf("link failed: %v", err)
	}
	outbound, inbound, err := shelf.ListLinks(root, a.ID)
	if err != nil {
		t.Fatalf("list links failed: %v", err)
	}
	if len(outbound) != 1 || len(inbound) != 0 {
		t.Fatalf("unexpected links after link: outbound=%d inbound=%d", len(outbound), len(inbound))
	}

	if _, err := executeCLI(t, "undo", "--root", root); err != nil {
		t.Fatalf("undo link failed: %v", err)
	}
	outbound, inbound, err = shelf.ListLinks(root, a.ID)
	if err != nil {
		t.Fatalf("list links after undo failed: %v", err)
	}
	if len(outbound) != 0 || len(inbound) != 0 {
		t.Fatalf("expected links to be removed after undo, outbound=%d inbound=%d", len(outbound), len(inbound))
	}
}

func TestCLIUndoRedoWithSteps(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "step target"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	if _, err := executeCLI(t, "set", "--root", root, task.ID, "--status", "in_progress"); err != nil {
		t.Fatalf("set in_progress failed: %v", err)
	}
	if _, err := executeCLI(t, "set", "--root", root, task.ID, "--status", "blocked"); err != nil {
		t.Fatalf("set blocked failed: %v", err)
	}

	cur, err := shelf.EnsureTaskExists(root, task.ID)
	if err != nil {
		t.Fatalf("load task failed: %v", err)
	}
	if cur.Status != "blocked" {
		t.Fatalf("expected blocked status, got: %s", cur.Status)
	}

	if _, err := executeCLI(t, "undo", "--root", root, "--steps", "2"); err != nil {
		t.Fatalf("undo --steps 2 failed: %v", err)
	}
	cur, err = shelf.EnsureTaskExists(root, task.ID)
	if err != nil {
		t.Fatalf("load task after undo failed: %v", err)
	}
	if cur.Status != "open" {
		t.Fatalf("expected open status after undo --steps 2, got: %s", cur.Status)
	}

	if _, err := executeCLI(t, "redo", "--root", root, "--steps", "2"); err != nil {
		t.Fatalf("redo --steps 2 failed: %v", err)
	}
	cur, err = shelf.EnsureTaskExists(root, task.ID)
	if err != nil {
		t.Fatalf("load task after redo failed: %v", err)
	}
	if cur.Status != "blocked" {
		t.Fatalf("expected blocked status after redo --steps 2, got: %s", cur.Status)
	}

	if _, err := executeCLI(t, "redo", "--root", root); err == nil || !strings.Contains(err.Error(), "redo history is empty") {
		t.Fatalf("expected redo empty error, got: %v", err)
	}
}

func TestCLIExplainCommand(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	dep, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "dep"})
	if err != nil {
		t.Fatalf("add dep failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "target"})
	if err != nil {
		t.Fatalf("add target failed: %v", err)
	}
	if err := shelf.LinkTasks(root, task.ID, dep.ID, "depends_on"); err != nil {
		t.Fatalf("link target->dep failed: %v", err)
	}

	out, err := executeCLI(t, "explain", "--root", root, task.ID)
	if err != nil {
		t.Fatalf("explain failed: %v", err)
	}
	if !strings.Contains(out, "Built-in Views:") || !strings.Contains(out, "ready") || !strings.Contains(out, "task is not ready") {
		t.Fatalf("unexpected explain output: %s", out)
	}

	out, err = executeCLI(t, "explain", "--root", root, task.ID, "--view", "ready", "--json")
	if err != nil {
		t.Fatalf("explain --view --json failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("failed to parse explain json output: %v output=%s", err, out)
	}
	requested, ok := payload["requested_view"].(map[string]any)
	if !ok {
		t.Fatalf("requested_view missing in explain payload: %s", out)
	}
	result, ok := requested["result"].(map[string]any)
	if !ok {
		t.Fatalf("requested_view.result missing: %s", out)
	}
	if match, ok := result["match"].(bool); !ok || match {
		t.Fatalf("expected requested ready view match=false, got: %v", result["match"])
	}
}

func TestCLIExportImportRoundTrip(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	a, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "export-a", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add a failed: %v", err)
	}
	b, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "export-b", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add b failed: %v", err)
	}
	if err := shelf.LinkTasks(root, a.ID, b.ID, "depends_on"); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	exportPath := filepath.Join(root, "backup.json")
	if _, err := executeCLI(t, "export", "--root", root, "--out", exportPath); err != nil {
		t.Fatalf("export failed: %v", err)
	}
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("failed to read export file: %v", err)
	}
	var exported map[string]any
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("invalid exported json: %v", err)
	}
	if version, ok := exported["version"].(float64); !ok || int(version) != 1 {
		t.Fatalf("unexpected export version: %v", exported["version"])
	}

	if _, err := executeCLI(t, "set", "--root", root, a.ID, "--status", "done"); err != nil {
		t.Fatalf("set status failed: %v", err)
	}
	if _, err := executeCLI(t, "unlink", "--root", root, "--from", a.ID, "--to", b.ID, "--type", "depends_on"); err != nil {
		t.Fatalf("unlink failed: %v", err)
	}

	if _, err := executeCLI(t, "import", "--root", root, "--in", exportPath); err != nil {
		t.Fatalf("import failed: %v", err)
	}
	a2, err := shelf.EnsureTaskExists(root, a.ID)
	if err != nil {
		t.Fatalf("load restored task failed: %v", err)
	}
	if a2.Status != "open" {
		t.Fatalf("expected restored status open, got: %s", a2.Status)
	}
	outbound, _, err := shelf.ListLinks(root, a.ID)
	if err != nil {
		t.Fatalf("list links failed: %v", err)
	}
	if len(outbound) != 1 || outbound[0].To != b.ID || outbound[0].Type != "depends_on" {
		t.Fatalf("expected depends_on edge restored, got: %+v", outbound)
	}
}

func TestCLIImportValidateDryRunAndMerge(t *testing.T) {
	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	current, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "current", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add current failed: %v", err)
	}
	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	payload := shelfExport{
		Version:    1,
		ExportedAt: time.Now().Local().Format(time.RFC3339),
		Config:     cfg,
		Tasks: []shelf.Task{
			{
				ID:        current.ID,
				Title:     "current",
				Kind:      "todo",
				Status:    "done",
				CreatedAt: current.CreatedAt,
				UpdatedAt: current.UpdatedAt,
			},
			{
				ID:        "01KK0000000000000000000000",
				Title:     "incoming-new",
				Kind:      "todo",
				Status:    "open",
				CreatedAt: time.Now().Local().Round(time.Second),
				UpdatedAt: time.Now().Local().Round(time.Second),
			},
		},
		Edges: map[string][]shelf.Edge{},
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}
	importPath := filepath.Join(root, "merge.json")
	if err := os.WriteFile(importPath, data, 0o644); err != nil {
		t.Fatalf("write payload failed: %v", err)
	}

	if _, err := executeCLI(t, "import", "--root", root, "--in", importPath, "--validate-only", "--merge"); err != nil {
		t.Fatalf("import validate-only failed: %v", err)
	}
	taskAfterValidate, err := shelf.EnsureTaskExists(root, current.ID)
	if err != nil {
		t.Fatalf("load task after validate-only failed: %v", err)
	}
	if taskAfterValidate.Status != "open" {
		t.Fatalf("validate-only should not mutate task status, got: %s", taskAfterValidate.Status)
	}

	dryOut, err := executeCLI(t, "import", "--root", root, "--in", importPath, "--dry-run", "--merge")
	if err != nil {
		t.Fatalf("import dry-run failed: %v", err)
	}
	var summary map[string]any
	if err := json.Unmarshal([]byte(dryOut), &summary); err != nil {
		t.Fatalf("parse dry-run summary failed: %v output=%s", err, dryOut)
	}
	if summary["mode"] != "merge" {
		t.Fatalf("expected mode=merge summary, got: %v", summary["mode"])
	}

	if _, err := executeCLI(t, "import", "--root", root, "--in", importPath, "--merge"); err != nil {
		t.Fatalf("import merge failed: %v", err)
	}
	taskAfterMerge, err := shelf.EnsureTaskExists(root, current.ID)
	if err != nil {
		t.Fatalf("load task after merge failed: %v", err)
	}
	if taskAfterMerge.Status != "done" {
		t.Fatalf("merge should apply incoming task status, got: %s", taskAfterMerge.Status)
	}
	if _, err := shelf.EnsureTaskExists(root, "01KK0000000000000000000000"); err != nil {
		t.Fatalf("merge should add incoming new task: %v", err)
	}

	if _, err := executeCLI(t, "import", "--root", root, "--in", importPath, "--merge", "--replace"); err == nil || !strings.Contains(err.Error(), "同時に指定") {
		t.Fatalf("expected merge/replace conflict error, got: %v", err)
	}
}

func TestCLICompletionCommands(t *testing.T) {
	out, err := executeCLI(t, "completion", "bash")
	if err != nil {
		t.Fatalf("completion bash failed: %v", err)
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

func extractIDFromAddOutput(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID: ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "ID: "))
		}
	}
	return ""
}
