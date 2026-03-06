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
