package shelf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDoctorDetectsParentCycleAndUnknownStatus(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	a, err := AddTask(root, AddTaskInput{Title: "A"})
	if err != nil {
		t.Fatalf("add A failed: %v", err)
	}
	b, err := AddTask(root, AddTaskInput{Title: "B", Parent: a.ID})
	if err != nil {
		t.Fatalf("add B failed: %v", err)
	}

	// Force a parent cycle and unknown status by direct file write for doctor test.
	taskAPath := filepath.Join(TasksDir(root), a.ID+".md")
	badTask := Task{
		ID:        a.ID,
		Title:     "A",
		Kind:      "todo",
		Status:    "invalid_state",
		Parent:    b.ID,
		CreatedAt: time.Now().UTC().Round(time.Second),
		UpdatedAt: time.Now().UTC().Round(time.Second),
	}
	data, err := FormatTaskMarkdown(badTask)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}
	if err := os.WriteFile(taskAPath, data, 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	report, err := RunDoctor(root)
	if err == nil {
		t.Fatal("expected doctor to report issues")
	}
	if len(report.Issues) == 0 {
		t.Fatal("expected issues")
	}
}

func TestDoctorDetectsBrokenEdge(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	task, err := AddTask(root, AddTaskInput{Title: "A"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	content := `[[edge]]
to = "MISSING"
type = "depends_on"

[[edge]]
to = "MISSING"
type = "depends_on"

[[edge]]
to = "MISSING2"
type = "unknown_type"
`
	edgePath := filepath.Join(EdgesDir(root), task.ID+".toml")
	if err := os.WriteFile(edgePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write edge file failed: %v", err)
	}

	report, err := RunDoctor(root)
	if err == nil {
		t.Fatal("expected doctor to report issues")
	}
	if len(report.Issues) < 2 {
		t.Fatalf("expected multiple issues, got %+v", report.Issues)
	}
}

func TestRunDoctorWithFixNormalizesTaskAndEdges(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	task, err := AddTask(root, AddTaskInput{Title: "A"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}
	dep, err := AddTask(root, AddTaskInput{Title: "B"})
	if err != nil {
		t.Fatalf("add dep failed: %v", err)
	}

	taskPath := filepath.Join(TasksDir(root), task.ID+".md")
	legacyTask := `+++
id = "` + task.ID + `"
title = "A"
kind = "todo"
state = "open"
created_at = "2026-03-05T12:34:56+09:00"
updated_at = "2026-03-05T12:34:56+09:00"
+++

body
`
	if err := os.WriteFile(taskPath, []byte(legacyTask), 0o644); err != nil {
		t.Fatalf("write legacy task failed: %v", err)
	}

	edgePath := filepath.Join(EdgesDir(root), task.ID+".toml")
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

	report, fixed, err := RunDoctorWithFix(root)
	if err != nil {
		t.Fatalf("doctor with fix should resolve fixable issues: report=%+v err=%v", report, err)
	}
	if fixed == 0 {
		t.Fatalf("expected fix count > 0")
	}

	normalizedTask, err := os.ReadFile(taskPath)
	if err != nil {
		t.Fatalf("read normalized task failed: %v", err)
	}
	content := string(normalizedTask)
	if !strings.Contains(content, `status = "open"`) || strings.Contains(content, "state = ") {
		t.Fatalf("task should be rewritten to status key: %s", content)
	}

	normalizedEdge, err := os.ReadFile(edgePath)
	if err != nil {
		t.Fatalf("read normalized edge failed: %v", err)
	}
	if strings.Count(string(normalizedEdge), "[[edge]]") != 1 {
		t.Fatalf("edge file should be deduplicated: %s", string(normalizedEdge))
	}
}

func TestRunDoctorWithFixRepairsSafeDataIssues(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	a, err := AddTask(root, AddTaskInput{Title: "A"})
	if err != nil {
		t.Fatalf("add A failed: %v", err)
	}
	b, err := AddTask(root, AddTaskInput{Title: "B"})
	if err != nil {
		t.Fatalf("add B failed: %v", err)
	}

	taskAPath := filepath.Join(TasksDir(root), a.ID+".md")
	a.Parent = "MISSING_PARENT"
	taskData, err := FormatTaskMarkdown(a)
	if err != nil {
		t.Fatalf("format task failed: %v", err)
	}
	if err := os.WriteFile(taskAPath, taskData, 0o644); err != nil {
		t.Fatalf("write task failed: %v", err)
	}

	missingSrcEdgePath := filepath.Join(EdgesDir(root), "MISSING_SRC.toml")
	if err := os.WriteFile(missingSrcEdgePath, []byte(`[[edge]]
to = "`+a.ID+`"
type = "depends_on"
`), 0o644); err != nil {
		t.Fatalf("write missing src edge failed: %v", err)
	}
	aEdgePath := filepath.Join(EdgesDir(root), a.ID+".toml")
	if err := os.WriteFile(aEdgePath, []byte(`[[edge]]
to = "MISSING_DST"
type = "depends_on"

[[edge]]
to = "`+b.ID+`"
type = "unknown_type"

[[edge]]
to = "`+b.ID+`"
type = "depends_on"
`), 0o644); err != nil {
		t.Fatalf("write A edge failed: %v", err)
	}

	report, fixed, err := RunDoctorWithFix(root)
	if err != nil {
		t.Fatalf("doctor with fix should pass: report=%+v err=%v", report, err)
	}
	if fixed == 0 {
		t.Fatalf("expected fixed > 0")
	}

	updatedA, err := NewTaskStore(root).Get(a.ID)
	if err != nil {
		t.Fatalf("load fixed A failed: %v", err)
	}
	if updatedA.Parent != "" {
		t.Fatalf("expected parent to be cleared, got: %q", updatedA.Parent)
	}
	if _, err := os.Stat(missingSrcEdgePath); !os.IsNotExist(err) {
		t.Fatalf("missing source edge file should be removed")
	}
	edges, err := NewEdgeStore(root).ListOutbound(a.ID)
	if err != nil {
		t.Fatalf("list outbound failed: %v", err)
	}
	if len(edges) != 1 || edges[0].To != b.ID || edges[0].Type != "depends_on" {
		t.Fatalf("expected only valid edge to remain, got: %+v", edges)
	}
}

func TestRunDoctorStrictWarnsTodoWithoutDue(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if _, err := AddTask(root, AddTaskInput{Title: "todo-no-due", Kind: "todo", Status: "open"}); err != nil {
		t.Fatalf("add todo failed: %v", err)
	}
	if _, err := AddTask(root, AddTaskInput{Title: "memo-no-due", Kind: "memo", Status: "open"}); err != nil {
		t.Fatalf("add memo failed: %v", err)
	}

	normalReport, err := RunDoctor(root)
	if err != nil {
		t.Fatalf("normal doctor should pass: %v", err)
	}
	if len(normalReport.Warnings) != 0 {
		t.Fatalf("normal doctor should not return strict warnings: %+v", normalReport.Warnings)
	}

	strictReport, err := RunDoctorWithOptions(root, DoctorOptions{Strict: true})
	if err != nil {
		t.Fatalf("strict doctor should not fail on warnings only: %v", err)
	}
	if len(strictReport.Warnings) != 1 {
		t.Fatalf("strict doctor should emit one warning, got: %+v", strictReport.Warnings)
	}
	if !strings.Contains(strictReport.Warnings[0].Message, "todo task has no due_on") {
		t.Fatalf("unexpected warning: %+v", strictReport.Warnings[0])
	}
}
