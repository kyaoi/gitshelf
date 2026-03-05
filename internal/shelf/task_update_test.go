package shelf

import "testing"

func TestSetTaskParentCycleRejected(t *testing.T) {
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

	parent := b.ID
	if _, err := SetTask(root, a.ID, SetTaskInput{Parent: &parent}); err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestSetTaskUpdatesStatusAndTitle(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	task, err := AddTask(root, AddTaskInput{Title: "Before"})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	title := "After"
	status := Status("done")
	updated, err := SetTask(root, task.ID, SetTaskInput{
		Title:  &title,
		Status: &status,
	})
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if updated.Title != "After" || updated.Status != "done" {
		t.Fatalf("unexpected updated task: %+v", updated)
	}
}
