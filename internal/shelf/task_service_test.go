package shelf

import (
	"os"
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
	if task.Kind != "todo" || task.State != "open" {
		t.Fatalf("unexpected defaults: kind=%s state=%s", task.Kind, task.State)
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
