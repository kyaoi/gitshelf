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
