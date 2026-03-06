package cli

import (
	"strings"
	"testing"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestEnumOptionsFromKinds(t *testing.T) {
	options := enumOptionsFromKinds([]shelf.Kind{"todo", "memo"})
	if len(options) != 2 {
		t.Fatalf("unexpected option length: %d", len(options))
	}
	if options[0].Value != "todo" || options[1].Value != "memo" {
		t.Fatalf("unexpected values: %+v", options)
	}
}

func TestEnumOptionsFromStatuses(t *testing.T) {
	options := enumOptionsFromStatuses([]shelf.Status{"open", "done"})
	if len(options) != 2 {
		t.Fatalf("unexpected option length: %d", len(options))
	}
	if options[0].Value != "open" || options[1].Value != "done" {
		t.Fatalf("unexpected values: %+v", options)
	}
}

func TestBuildAddSummary(t *testing.T) {
	summary := buildAddSummary("A", "todo", "open", "today", "(none)", "root")
	for _, expected := range []string{
		"Title: A",
		"Kind: todo",
		"Status: open",
		"Due: today",
		"Repeat: (none)",
		"Parent: root",
	} {
		if !strings.Contains(summary, expected) {
			t.Fatalf("summary missing %q: %s", expected, summary)
		}
	}
}
