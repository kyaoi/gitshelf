package shelf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTaskMarkdownRoundTrip(t *testing.T) {
	now := time.Date(2026, 3, 5, 12, 34, 56, 0, time.FixedZone("JST", 9*60*60))
	orig := Task{
		ID:          "01JABCDEF0123456789XYZ",
		Title:       "月曜日にやること",
		Kind:        Kind("todo"),
		Status:      Status("open"),
		Tags:        []string{"backend", "urgent"},
		EstimateMin: 90,
		SpentMin:    30,
		TimerStart:  "2026-03-05T13:00:00+09:00",
		DueOn:       "2026-03-10",
		RepeatEvery: "1w",
		ArchivedAt:  "2026-03-05T12:34:56+09:00",
		Parent:      "01JWEEKGOAL000000000000",
		CreatedAt:   now,
		UpdatedAt:   now,
		Body:        "メモ本文",
	}

	data, err := FormatTaskMarkdown(orig)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}
	if !strings.Contains(string(data), "status = \"open\"") {
		t.Fatalf("formatted task should include status key: %s", string(data))
	}

	parsed, err := ParseTaskMarkdown(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if parsed.ID != orig.ID || parsed.Title != orig.Title || parsed.Kind != orig.Kind || parsed.Status != orig.Status {
		t.Fatalf("parsed task mismatch: %+v", parsed)
	}
	if len(parsed.Tags) != 2 || parsed.Tags[0] != "backend" || parsed.Tags[1] != "urgent" {
		t.Fatalf("parsed tags mismatch: %+v", parsed.Tags)
	}
	if parsed.EstimateMin != 90 || parsed.SpentMin != 30 || parsed.TimerStart != "2026-03-05T13:00:00+09:00" {
		t.Fatalf("parsed worklog mismatch: %+v", parsed)
	}
	if parsed.DueOn != orig.DueOn {
		t.Fatalf("parsed due_on mismatch: %+v", parsed)
	}
	if parsed.RepeatEvery != orig.RepeatEvery || parsed.ArchivedAt != orig.ArchivedAt {
		t.Fatalf("parsed repeat/archive mismatch: %+v", parsed)
	}
	if parsed.Parent != orig.Parent || parsed.Body != orig.Body {
		t.Fatalf("parsed optional fields mismatch: %+v", parsed)
	}
	if !parsed.CreatedAt.Equal(orig.CreatedAt) || !parsed.UpdatedAt.Equal(orig.UpdatedAt) {
		t.Fatalf("parsed timestamps mismatch: %+v", parsed)
	}
}

func TestParseTaskMarkdownInvalidDueOn(t *testing.T) {
	raw := `+++
id = "01JABCDEF0123456789XYZ"
title = "invalid due"
kind = "todo"
status = "open"
due_on = "2026-99-99"
created_at = "2026-03-05T12:34:56+09:00"
updated_at = "2026-03-05T12:34:56+09:00"
+++

body
`
	if _, err := ParseTaskMarkdown([]byte(raw)); err == nil || !strings.Contains(err.Error(), "invalid due_on") {
		t.Fatalf("expected invalid due_on error, got: %v", err)
	}
}

func TestParseTaskMarkdownInvalidRepeatEvery(t *testing.T) {
	raw := `+++
id = "01JABCDEF0123456789XYZ"
title = "invalid repeat"
kind = "todo"
status = "open"
repeat_every = "daily"
created_at = "2026-03-05T12:34:56+09:00"
updated_at = "2026-03-05T12:34:56+09:00"
+++

body
`
	if _, err := ParseTaskMarkdown([]byte(raw)); err == nil || !strings.Contains(err.Error(), "invalid repeat_every") {
		t.Fatalf("expected invalid repeat_every error, got: %v", err)
	}
}

func TestParseTaskMarkdownInvalidArchivedAt(t *testing.T) {
	raw := `+++
id = "01JABCDEF0123456789XYZ"
title = "invalid archived"
kind = "todo"
status = "open"
archived_at = "2026-03-05"
created_at = "2026-03-05T12:34:56+09:00"
updated_at = "2026-03-05T12:34:56+09:00"
+++

body
`
	if _, err := ParseTaskMarkdown([]byte(raw)); err == nil || !strings.Contains(err.Error(), "invalid archived_at") {
		t.Fatalf("expected invalid archived_at error, got: %v", err)
	}
}

func TestParseTaskMarkdownDueKeywordNormalized(t *testing.T) {
	raw := `+++
id = "01JABCDEF0123456789XYZ"
title = "keyword due"
kind = "todo"
status = "open"
due_on = "tomorrow"
created_at = "2026-03-05T12:34:56+09:00"
updated_at = "2026-03-05T12:34:56+09:00"
+++

body
`
	task, err := ParseTaskMarkdown([]byte(raw))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	want := time.Now().Local().AddDate(0, 0, 1).Format("2006-01-02")
	if task.DueOn != want {
		t.Fatalf("unexpected normalized due_on: got=%q want=%q", task.DueOn, want)
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
		Status:    Status("open"),
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

func TestParseTaskMarkdownLegacyStateField(t *testing.T) {
	raw := `+++
id = "01JABCDEF0123456789XYZ"
title = "legacy"
kind = "todo"
state = "open"
created_at = "2026-03-05T12:34:56+09:00"
updated_at = "2026-03-05T12:34:56+09:00"
+++

body
`
	task, err := ParseTaskMarkdown([]byte(raw))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if task.Status != "open" {
		t.Fatalf("unexpected status: %s", task.Status)
	}

	data, err := FormatTaskMarkdown(task)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}
	formatted := string(data)
	if !strings.Contains(formatted, "status = \"open\"") || strings.Contains(formatted, "state = ") {
		t.Fatalf("formatted task should use status key: %s", formatted)
	}
}
