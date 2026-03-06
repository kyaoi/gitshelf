package cli

import (
	"strings"
	"testing"
)

func TestResolveEditorCommandPriority(t *testing.T) {
	lookup := func(key string) (string, bool) {
		switch key {
		case "VISUAL":
			return "nvim", true
		case "EDITOR":
			return "vim", true
		default:
			return "", false
		}
	}
	got, err := resolveEditorCommand(lookup)
	if err != nil {
		t.Fatalf("resolve editor failed: %v", err)
	}
	if got != "nvim" {
		t.Fatalf("expected VISUAL to win, got %q", got)
	}
}

func TestResolveEditorCommandFallback(t *testing.T) {
	lookup := func(_ string) (string, bool) {
		return "", false
	}
	got, err := resolveEditorCommand(lookup)
	if err != nil {
		t.Fatalf("resolve editor failed: %v", err)
	}
	if got != "vi" {
		t.Fatalf("expected default vi, got %q", got)
	}
}

func TestResolveEditTaskIDWithArgument(t *testing.T) {
	id, err := resolveEditTaskID(nil, []string{"01TESTID"}, false)
	if err != nil {
		t.Fatalf("resolve id failed: %v", err)
	}
	if id != "01TESTID" {
		t.Fatalf("unexpected id: %s", id)
	}
}

func TestResolveEditTaskIDWithoutArgumentOnNonTTY(t *testing.T) {
	_, err := resolveEditTaskID(nil, nil, false)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "<id> を指定してください") {
		t.Fatalf("unexpected error: %v", err)
	}
}
