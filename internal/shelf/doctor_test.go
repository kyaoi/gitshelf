package shelf

import (
	"os"
	"path/filepath"
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
