package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestBuildParentSelectionOptionsHierarchyOrder(t *testing.T) {
	now := time.Now().UTC().Round(time.Second)
	tasks := []shelf.Task{
		{ID: "01A", Title: "A", Kind: "todo", Status: "open", CreatedAt: now, UpdatedAt: now},
		{ID: "01B", Title: "B", Kind: "todo", Status: "open", Parent: "01A", CreatedAt: now, UpdatedAt: now},
		{ID: "01C", Title: "C", Kind: "todo", Status: "open", Parent: "01B", CreatedAt: now, UpdatedAt: now},
		{ID: "01D", Title: "D", Kind: "todo", Status: "open", CreatedAt: now, UpdatedAt: now},
	}

	options := buildParentSelectionOptions(tasks, "")
	if len(options) != 5 {
		t.Fatalf("unexpected option length: %d", len(options))
	}

	wantLabels := []string{
		"(root)",
		"A",
		"└─ B",
		"   └─ C",
		"D",
	}
	for i, want := range wantLabels {
		if options[i].Label != want {
			t.Fatalf("label[%d] = %q, want %q", i, options[i].Label, want)
		}
	}
	if options[0].Value != "root" || options[1].Value != "01A" || options[2].Value != "01B" || options[3].Value != "01C" {
		t.Fatalf("unexpected option values: %+v", options)
	}
}

func TestBuildParentSelectionOptionsDuplicateTitles(t *testing.T) {
	now := time.Now().UTC().Round(time.Second)
	tasks := []shelf.Task{
		{ID: "01A", Title: "same", Kind: "todo", Status: "open", CreatedAt: now, UpdatedAt: now},
		{ID: "01B", Title: "same", Kind: "idea", Status: "blocked", CreatedAt: now, UpdatedAt: now},
	}

	options := buildParentSelectionOptions(tasks, "")
	if len(options) != 3 {
		t.Fatalf("unexpected option length: %d", len(options))
	}
	if options[1].Label != "same (todo/open)" {
		t.Fatalf("unexpected duplicate label[1]: %q", options[1].Label)
	}
	if options[2].Label != "same (idea/blocked)" {
		t.Fatalf("unexpected duplicate label[2]: %q", options[2].Label)
	}
}

func TestBuildParentSelectionOptionsOmitsIDsForUniqueTitles(t *testing.T) {
	now := time.Now().UTC().Round(time.Second)
	tasks := []shelf.Task{
		{ID: "01A", Title: "unique", Kind: "todo", Status: "open", CreatedAt: now, UpdatedAt: now},
	}

	options := buildParentSelectionOptions(tasks, "")
	if len(options) != 2 {
		t.Fatalf("unexpected option length: %d", len(options))
	}
	if strings.Contains(options[1].Label, "[01A]") || strings.Contains(options[1].Label, shelf.ShortID("01A")) {
		t.Fatalf("label should not include ID for unique title: %q", options[1].Label)
	}
}

func TestBuildParentSelectionOptionsExcludeID(t *testing.T) {
	now := time.Now().UTC().Round(time.Second)
	tasks := []shelf.Task{
		{ID: "01A", Title: "A", Kind: "todo", Status: "open", CreatedAt: now, UpdatedAt: now},
		{ID: "01B", Title: "B", Kind: "todo", Status: "open", Parent: "01A", CreatedAt: now, UpdatedAt: now},
	}

	options := buildParentSelectionOptions(tasks, "01A")
	if len(options) != 1 {
		t.Fatalf("expected only root when excluding current branch, got %+v", options)
	}
	if options[0].Value != "root" {
		t.Fatalf("unexpected root option: %+v", options[0])
	}
}
