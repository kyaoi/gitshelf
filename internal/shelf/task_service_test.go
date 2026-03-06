package shelf

import (
	"os"
	"strings"
	"testing"
)

func TestAddTaskWithDefaults(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	task, err := AddTask(root, AddTaskInput{
		Title: "new task",
	})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}
	if task.Kind != "todo" || task.Status != "open" {
		t.Fatalf("unexpected defaults: kind=%s status=%s", task.Kind, task.Status)
	}

	if _, err := os.Stat(TasksDir(root) + "/" + task.ID + ".md"); err != nil {
		t.Fatalf("task file does not exist: %v", err)
	}
}

func TestAddTaskWithParent(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	parent, err := AddTask(root, AddTaskInput{Title: "parent"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	child, err := AddTask(root, AddTaskInput{
		Title:  "child",
		Parent: parent.ID,
	})
	if err != nil {
		t.Fatalf("add child failed: %v", err)
	}
	if child.Parent != parent.ID {
		t.Fatalf("unexpected parent: %s", child.Parent)
	}
}

func TestAddTaskWithDueOn(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	task, err := AddTask(root, AddTaskInput{
		Title: "with due",
		DueOn: "2026-03-31",
	})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}
	if task.DueOn != "2026-03-31" {
		t.Fatalf("unexpected due_on: %q", task.DueOn)
	}
}

func TestAddTaskRejectsInvalidDueOn(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	if _, err := AddTask(root, AddTaskInput{
		Title: "bad due",
		DueOn: "2026-99-31",
	}); err == nil || !strings.Contains(err.Error(), "invalid due_on") {
		t.Fatalf("expected invalid due_on error, got: %v", err)
	}
}
