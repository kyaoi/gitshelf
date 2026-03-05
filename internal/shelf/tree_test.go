package shelf

import (
	"os"
	"testing"
	"time"
)

func TestBuildTreeFromRoot(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if err := os.MkdirAll(TasksDir(root), 0o755); err != nil {
		t.Fatal(err)
	}

	store := NewTaskStore(root)
	now := time.Now().UTC().Round(time.Second)
	mustUpsertTask(t, store, Task{ID: "01A", Title: "A", Kind: "todo", State: "open", CreatedAt: now, UpdatedAt: now})
	mustUpsertTask(t, store, Task{ID: "01B", Title: "B", Kind: "todo", State: "open", Parent: "01A", CreatedAt: now, UpdatedAt: now})
	mustUpsertTask(t, store, Task{ID: "01C", Title: "C", Kind: "todo", State: "open", Parent: "01A", CreatedAt: now, UpdatedAt: now})
	mustUpsertTask(t, store, Task{ID: "01D", Title: "D", Kind: "todo", State: "done", Parent: "01B", CreatedAt: now, UpdatedAt: now})

	tree, err := BuildTree(root, TreeOptions{})
	if err != nil {
		t.Fatalf("build tree failed: %v", err)
	}
	if len(tree) != 1 {
		t.Fatalf("expected 1 root node, got %d", len(tree))
	}
	if tree[0].Task.ID != "01A" {
		t.Fatalf("unexpected root: %s", tree[0].Task.ID)
	}
	if len(tree[0].Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(tree[0].Children))
	}
	if tree[0].Children[0].Task.ID != "01B" || tree[0].Children[1].Task.ID != "01C" {
		t.Fatalf("unexpected child order: %+v", tree[0].Children)
	}
}

func TestBuildTreeFromNodeAndStateFilter(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	store := NewTaskStore(root)
	now := time.Now().UTC().Round(time.Second)
	mustUpsertTask(t, store, Task{ID: "01A", Title: "A", Kind: "todo", State: "open", CreatedAt: now, UpdatedAt: now})
	mustUpsertTask(t, store, Task{ID: "01B", Title: "B", Kind: "todo", State: "done", Parent: "01A", CreatedAt: now, UpdatedAt: now})
	mustUpsertTask(t, store, Task{ID: "01C", Title: "C", Kind: "todo", State: "open", Parent: "01B", CreatedAt: now, UpdatedAt: now})

	tree, err := BuildTree(root, TreeOptions{FromID: "01B", State: "done"})
	if err != nil {
		t.Fatalf("build tree failed: %v", err)
	}
	if len(tree) != 1 {
		t.Fatalf("expected 1 node, got %d", len(tree))
	}
	if tree[0].Task.ID != "01B" {
		t.Fatalf("unexpected root node: %s", tree[0].Task.ID)
	}
	if len(tree[0].Children) != 0 {
		t.Fatalf("expected children to be filtered out by state, got %d", len(tree[0].Children))
	}
}

func mustUpsertTask(t *testing.T, store *TaskStore, task Task) {
	t.Helper()
	if err := store.Upsert(task); err != nil {
		t.Fatalf("upsert failed: %v", err)
	}
}
