package interactive

import "testing"

func TestFilterOptions(t *testing.T) {
	options := []Option{
		{Label: "[01AAAAAA] task one  (todo/open)", SearchText: "task one 01AAAAAA"},
		{Label: "[01BBBBBB] note two  (memo/done)", SearchText: "note two 01BBBBBB"},
	}

	got := filterOptions(options, "note")
	if len(got) != 1 || got[0].Label != options[1].Label {
		t.Fatalf("unexpected filter result: %+v", got)
	}
}

func TestCursorMoveWraps(t *testing.T) {
	if got := moveUp(0, 3); got != 2 {
		t.Fatalf("moveUp wrap failed: %d", got)
	}
	if got := moveDown(2, 3); got != 0 {
		t.Fatalf("moveDown wrap failed: %d", got)
	}
}

func TestClampSelectorOffsetKeepsCursorVisible(t *testing.T) {
	if got := clampSelectorOffset(0, 0, 20, 5); got != 0 {
		t.Fatalf("expected offset 0, got %d", got)
	}
	if got := clampSelectorOffset(0, 6, 20, 5); got != 2 {
		t.Fatalf("expected offset 2, got %d", got)
	}
	if got := clampSelectorOffset(10, 3, 20, 5); got != 3 {
		t.Fatalf("expected offset 3, got %d", got)
	}
	if got := clampSelectorOffset(99, 19, 20, 5); got != 15 {
		t.Fatalf("expected final offset 15, got %d", got)
	}
}

func TestSelectorSearchLineShowsViewportRange(t *testing.T) {
	line := selectorSearchLine("Search", "(none)", 5, 10, 42)
	if line != "Search: (none)  [6-15/42]" {
		t.Fatalf("unexpected search line: %q", line)
	}
}

func TestMatchSubmitShortcut(t *testing.T) {
	cfg := SelectConfig{
		SubmitValue:     "save",
		SubmitShortcuts: []string{"Ctrl+S", "Ctrl+Enter"},
	}
	if !matchSubmitShortcut(cfg, keyEvent{Kind: keyKindCtrlS}) {
		t.Fatal("expected ctrl+s to match")
	}
	if !matchSubmitShortcut(cfg, keyEvent{Kind: keyKindCtrlEnter}) {
		t.Fatal("expected ctrl+enter to match")
	}
	if matchSubmitShortcut(cfg, keyEvent{Kind: keyKindEnter}) {
		t.Fatal("plain enter should not match submit shortcut")
	}
}
