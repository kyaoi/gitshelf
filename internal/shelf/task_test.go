package shelf

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTaskMarkdownRoundTrip(t *testing.T) {
	now := time.Date(2026, 3, 5, 12, 34, 56, 0, time.FixedZone("JST", 9*60*60))
	orig := Task{
		ID:        "01JABCDEF0123456789XYZ",
		Title:     "月曜日にやること",
		Kind:      Kind("todo"),
		State:     State("open"),
		Parent:    "01JWEEKGOAL000000000000",
		CreatedAt: now,
		UpdatedAt: now,
		Body:      "メモ本文",
	}

	data, err := FormatTaskMarkdown(orig)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	parsed, err := ParseTaskMarkdown(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if parsed.ID != orig.ID || parsed.Title != orig.Title || parsed.Kind != orig.Kind || parsed.State != orig.State {
		t.Fatalf("parsed task mismatch: %+v", parsed)
	}
	if parsed.Parent != orig.Parent || parsed.Body != orig.Body {
		t.Fatalf("parsed optional fields mismatch: %+v", parsed)
	}
	if !parsed.CreatedAt.Equal(orig.CreatedAt) || !parsed.UpdatedAt.Equal(orig.UpdatedAt) {
		t.Fatalf("parsed timestamps mismatch: %+v", parsed)
	}
}

func TestTaskStoreCRUD(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(TasksDir(root), 0o755); err != nil {
		t.Fatal(err)
	}
	store := NewTaskStore(root)

	now := time.Now().UTC().Round(time.Second)
	task := Task{
		ID:        "01JABCDEF0123456789XYZ",
		Title:     "test task",
		Kind:      Kind("todo"),
		State:     State("open"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := store.Create(task); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	got, err := store.Get(task.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Title != "test task" {
		t.Fatalf("unexpected title: %q", got.Title)
	}

	got.Title = "updated"
	got.UpdatedAt = now.Add(time.Minute)
	if err := store.Update(got); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	updated, err := store.Get(task.ID)
	if err != nil {
		t.Fatalf("get after update failed: %v", err)
	}
	if updated.Title != "updated" {
		t.Fatalf("unexpected updated title: %q", updated.Title)
	}

	listed, err := store.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != task.ID {
		t.Fatalf("unexpected list result: %+v", listed)
	}

	path := filepath.Join(TasksDir(root), task.ID+".md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("task file missing: %v", err)
	}
}
