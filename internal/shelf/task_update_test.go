package shelf

import (
	"strings"
	"testing"
)

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

func TestSetTaskUpdatesAndClearsDueOn(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	task, err := AddTask(root, AddTaskInput{Title: "with due"})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	due := "2026-04-01"
	updated, err := SetTask(root, task.ID, SetTaskInput{DueOn: &due})
	if err != nil {
		t.Fatalf("set due failed: %v", err)
	}
	if updated.DueOn != due {
		t.Fatalf("unexpected due after set: %+v", updated)
	}

	empty := ""
	cleared, err := SetTask(root, task.ID, SetTaskInput{DueOn: &empty})
	if err != nil {
		t.Fatalf("clear due failed: %v", err)
	}
	if cleared.DueOn != "" {
		t.Fatalf("expected due_on to be cleared, got: %q", cleared.DueOn)
	}
}

func TestSetTaskRejectsInvalidDueOn(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	task, err := AddTask(root, AddTaskInput{Title: "bad due"})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	due := "2026-99-01"
	if _, err := SetTask(root, task.ID, SetTaskInput{DueOn: &due}); err == nil || !strings.Contains(err.Error(), "invalid due_on") {
		t.Fatalf("expected invalid due_on error, got: %v", err)
	}
}

func TestSetTaskUpdatesRepeatEvery(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	task, err := AddTask(root, AddTaskInput{Title: "repeat target"})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	repeatEvery := "2w"
	updated, err := SetTask(root, task.ID, SetTaskInput{RepeatEvery: &repeatEvery})
	if err != nil {
		t.Fatalf("set repeat failed: %v", err)
	}
	if updated.RepeatEvery != "2w" {
		t.Fatalf("unexpected repeat_every: %+v", updated)
	}
}

func TestSetTaskRejectsInvalidRepeatEvery(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	task, err := AddTask(root, AddTaskInput{Title: "bad repeat"})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	bad := "foo"
	if _, err := SetTask(root, task.ID, SetTaskInput{RepeatEvery: &bad}); err == nil || !strings.Contains(err.Error(), "invalid repeat_every") {
		t.Fatalf("expected invalid repeat_every error, got: %v", err)
	}
}

func TestSetTaskUpdatesAndClearsArchivedAt(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	task, err := AddTask(root, AddTaskInput{Title: "archive target"})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	archivedAt := "2026-03-06T10:00:00+09:00"
	updated, err := SetTask(root, task.ID, SetTaskInput{ArchivedAt: &archivedAt})
	if err != nil {
		t.Fatalf("set archived failed: %v", err)
	}
	if updated.ArchivedAt != archivedAt {
		t.Fatalf("unexpected archived_at: %+v", updated)
	}

	empty := ""
	cleared, err := SetTask(root, task.ID, SetTaskInput{ArchivedAt: &empty})
	if err != nil {
		t.Fatalf("clear archived failed: %v", err)
	}
	if cleared.ArchivedAt != "" {
		t.Fatalf("expected archived_at cleared, got: %q", cleared.ArchivedAt)
	}
}

func TestSetTaskUpdatesTagsAndRegistersConfig(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	task, err := AddTask(root, AddTaskInput{Title: "tag target", Tags: []string{"backend"}})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	updated, err := SetTask(root, task.ID, SetTaskInput{
		AddTags:    []string{"urgent"},
		RemoveTags: []string{"backend"},
	})
	if err != nil {
		t.Fatalf("set tags failed: %v", err)
	}
	if len(updated.Tags) != 1 || updated.Tags[0] != "urgent" {
		t.Fatalf("unexpected tags: %+v", updated.Tags)
	}

	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if len(cfg.Tags) != 2 || cfg.Tags[0] != "backend" || cfg.Tags[1] != "urgent" {
		t.Fatalf("config tags should keep catalog entries: %+v", cfg.Tags)
	}
}
