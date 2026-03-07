package shelf

import "testing"

func TestListTasksTagFilters(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if _, err := AddTask(root, AddTaskInput{Title: "backend-open", Tags: []string{"backend"}}); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if _, err := AddTask(root, AddTaskInput{Title: "frontend-open", Tags: []string{"frontend", "wip"}}); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if _, err := AddTask(root, AddTaskInput{Title: "backend-done", Tags: []string{"backend", "done-tag"}}); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	tagged, err := ListTasks(root, TaskFilter{Tags: []string{"backend"}})
	if err != nil {
		t.Fatalf("list tagged failed: %v", err)
	}
	if len(tagged) != 2 {
		t.Fatalf("expected 2 backend tasks, got %d", len(tagged))
	}

	withoutWIP, err := ListTasks(root, TaskFilter{NotTags: []string{"wip"}})
	if err != nil {
		t.Fatalf("list not-tag failed: %v", err)
	}
	if len(withoutWIP) != 2 {
		t.Fatalf("expected 2 tasks without wip, got %d", len(withoutWIP))
	}
}

func TestListTasksUnknownTagFilter(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if _, err := AddTask(root, AddTaskInput{Title: "task", Tags: []string{"backend"}}); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if _, err := ListTasks(root, TaskFilter{Tags: []string{"unknown"}}); err == nil {
		t.Fatal("expected unknown tag error")
	}
}
