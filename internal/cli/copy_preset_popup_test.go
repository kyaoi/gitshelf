package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestRenderCopyPresetPayloadSupportsSubtreeAndAbsolutePath(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	parent := shelf.Task{ID: "01A", Title: "Parent"}
	child := shelf.Task{ID: "01B", Title: "Child", Parent: "01A"}
	model := calendarTUIModel{
		rootDir:           root,
		mode:              calendarModeTree,
		copySeparator:     "\n",
		taskByID:          map[string]shelf.Task{"01A": parent, "01B": child},
		allTasks:          []shelf.Task{parent, child},
		treeRows:          []cockpitTreeRow{{Task: parent}, {Task: child}},
		treeRowIndex:      0,
		markedTaskIDs:     map[string]struct{}{},
		rangeBaseIDs:      map[string]struct{}{},
		sectionRows:       map[calendarSectionID]int{},
		boardRowIndex:     map[int]int{},
		collapsedTree:     map[string]struct{}{},
		outboundCount:     map[string]int{},
		inboundCount:      map[string]int{},
		readiness:         map[string]shelf.TaskReadiness{},
		titleByID:         map[string]string{},
		effectiveDue:      map[string]string{},
		linkCollapsedTree: map[string]struct{}{},
	}

	text, count, err := model.renderCopyPresetPayload(shelf.CopyPreset{
		Name:     "subtree-path",
		Scope:    shelf.CopyPresetScopeSubtree,
		Template: "{{path}}\n{{subtree}}",
		JoinWith: "\n\n",
	})
	if err != nil {
		t.Fatalf("render copy preset failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("unexpected rendered item count: %d", count)
	}
	expectedPath := filepath.Join(shelf.TasksDir(root), "01A.md")
	expected := expectedPath + "\nParent\n  Child"
	if text != expected {
		t.Fatalf("unexpected rendered copy payload:\n%s\nwant:\n%s", text, expected)
	}
}

func TestRenderCopyPresetPayloadDeduplicatesNestedMarkedSubtrees(t *testing.T) {
	parent := shelf.Task{ID: "01A", Title: "Parent"}
	child := shelf.Task{ID: "01B", Title: "Child", Parent: "01A"}
	model := calendarTUIModel{
		rootDir:       t.TempDir(),
		mode:          calendarModeTree,
		copySeparator: "\n",
		taskByID:      map[string]shelf.Task{"01A": parent, "01B": child},
		allTasks:      []shelf.Task{parent, child},
		treeRows:      []cockpitTreeRow{{Task: parent}, {Task: child}},
		treeRowIndex:  0,
		markedTaskIDs: map[string]struct{}{"01A": {}, "01B": {}},
		rangeBaseIDs:  map[string]struct{}{},
		sectionRows:   map[calendarSectionID]int{},
		boardRowIndex: map[int]int{},
	}

	text, count, err := model.renderCopyPresetPayload(shelf.CopyPreset{
		Name:     "subtree-only",
		Scope:    shelf.CopyPresetScopeSubtree,
		Template: "{{subtree}}",
		JoinWith: "\n\n",
	})
	if err != nil {
		t.Fatalf("render copy preset failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one deduplicated subtree, got %d", count)
	}
	if text != "Parent\n  Child" {
		t.Fatalf("unexpected subtree text: %q", text)
	}
}

func TestCopyPresetSaveCommandUsesShellEscapes(t *testing.T) {
	model := calendarTUIModel{}
	command := model.copyPresetSaveCommand(shelf.CopyPreset{
		Name:     "subtree-path",
		Scope:    shelf.CopyPresetScopeSubtree,
		Template: "{{path}}\n{{subtree}}",
		JoinWith: "\n\n",
	})
	if !strings.Contains(command, "config copy-preset set") {
		t.Fatalf("unexpected save command: %s", command)
	}
	if !strings.Contains(command, "--template $'{{path}}\\n{{subtree}}'") {
		t.Fatalf("template should use escaped shell string: %s", command)
	}
	if !strings.Contains(command, "--join-with $'\\n\\n'") {
		t.Fatalf("join-with should use escaped shell string: %s", command)
	}
}
