package interactive

import "testing"

func TestTextCursorHelpersSupportMidStringEdit(t *testing.T) {
	value := "ab"
	cursor := 2

	cursor = MoveTextCursorLeft(value, cursor)
	if cursor != 1 {
		t.Fatalf("expected cursor at 1, got %d", cursor)
	}

	value, cursor = InsertRuneAtCursor(value, cursor, 'X')
	if value != "aXb" || cursor != 2 {
		t.Fatalf("unexpected insert result: value=%q cursor=%d", value, cursor)
	}

	value, cursor = DeleteRuneBeforeCursor(value, cursor)
	if value != "ab" || cursor != 1 {
		t.Fatalf("unexpected delete result: value=%q cursor=%d", value, cursor)
	}

	cursor = MoveTextCursorRight(value, cursor)
	if cursor != 2 {
		t.Fatalf("expected cursor at 2, got %d", cursor)
	}
}

func TestRenderTextCursorUsesClampedCursor(t *testing.T) {
	got := RenderTextCursor("abc", 99)
	if got != "abc_" {
		t.Fatalf("unexpected rendered cursor: %q", got)
	}
}
