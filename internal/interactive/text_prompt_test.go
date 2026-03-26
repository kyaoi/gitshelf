package interactive

import "testing"

func TestApplyTextPromptKeyAppendsAndEnterDone(t *testing.T) {
	value := ""
	cursor := 0
	for _, r := range []rune("Task q1") {
		done, canceled, next, nextCursor := applyTextPromptKey(value, cursor, keyEvent{Kind: keyKindRune, Rune: r})
		if done || canceled {
			t.Fatalf("unexpected terminal state while typing: done=%v canceled=%v", done, canceled)
		}
		value = next
		cursor = nextCursor
	}
	if value != "Task q1" {
		t.Fatalf("unexpected value: %q", value)
	}

	done, canceled, next, nextCursor := applyTextPromptKey(value, cursor, keyEvent{Kind: keyKindEnter})
	if !done || canceled {
		t.Fatalf("enter should finish input: done=%v canceled=%v", done, canceled)
	}
	if next != value {
		t.Fatalf("enter should keep value: got %q want %q", next, value)
	}
	if nextCursor != cursor {
		t.Fatalf("enter should keep cursor: got %d want %d", nextCursor, cursor)
	}
}

func TestApplyTextPromptKeyBackspace(t *testing.T) {
	done, canceled, next, nextCursor := applyTextPromptKey("ab", 2, keyEvent{Kind: keyKindBackspace})
	if done || canceled {
		t.Fatalf("backspace should not finish: done=%v canceled=%v", done, canceled)
	}
	if next != "a" {
		t.Fatalf("unexpected value after backspace: %q", next)
	}
	if nextCursor != 1 {
		t.Fatalf("unexpected cursor after backspace: %d", nextCursor)
	}
}

func TestApplyTextPromptKeyCancelKeys(t *testing.T) {
	for _, key := range []keyKind{keyKindCtrlC, keyKindEsc} {
		done, canceled, next, nextCursor := applyTextPromptKey("abc", 3, keyEvent{Kind: key})
		if done || !canceled {
			t.Fatalf("expected canceled for key %v: done=%v canceled=%v", key, done, canceled)
		}
		if next != "abc" {
			t.Fatalf("cancel should not alter value: %q", next)
		}
		if nextCursor != 3 {
			t.Fatalf("cancel should keep cursor: %d", nextCursor)
		}
	}
}

func TestApplyTextPromptKeyUnicodeInputAndBackspace(t *testing.T) {
	value := ""
	cursor := 0
	for _, r := range []rune("日本語") {
		done, canceled, next, nextCursor := applyTextPromptKey(value, cursor, keyEvent{Kind: keyKindRune, Rune: r})
		if done || canceled {
			t.Fatalf("unexpected terminal state while typing unicode: done=%v canceled=%v", done, canceled)
		}
		value = next
		cursor = nextCursor
	}
	if value != "日本語" {
		t.Fatalf("unexpected unicode value: %q", value)
	}

	_, _, value, cursor = applyTextPromptKey(value, cursor, keyEvent{Kind: keyKindBackspace})
	if value != "日本" {
		t.Fatalf("unexpected unicode value after backspace: %q", value)
	}
	if cursor != 2 {
		t.Fatalf("unexpected unicode cursor after backspace: %d", cursor)
	}
}

func TestApplyTextPromptKeySupportsCursorMovementAndMidStringInsert(t *testing.T) {
	value := "ab"
	cursor := 2

	_, _, value, cursor = applyTextPromptKey(value, cursor, keyEvent{Kind: keyKindLeft})
	if cursor != 1 {
		t.Fatalf("expected cursor to move left, got %d", cursor)
	}

	_, _, value, cursor = applyTextPromptKey(value, cursor, keyEvent{Kind: keyKindRune, Rune: 'X'})
	if value != "aXb" || cursor != 2 {
		t.Fatalf("unexpected mid-string insert: value=%q cursor=%d", value, cursor)
	}
}

func TestApplyTextPromptKeySupportsHomeEndAndDelete(t *testing.T) {
	value := "abcd"
	cursor := 2

	_, _, value, cursor = applyTextPromptKey(value, cursor, keyEvent{Kind: keyKindHome})
	if cursor != 0 {
		t.Fatalf("expected cursor at start, got %d", cursor)
	}

	_, _, value, cursor = applyTextPromptKey(value, cursor, keyEvent{Kind: keyKindEnd})
	if cursor != 4 {
		t.Fatalf("expected cursor at end, got %d", cursor)
	}

	_, _, value, cursor = applyTextPromptKey(value, 1, keyEvent{Kind: keyKindDelete})
	if value != "acd" || cursor != 1 {
		t.Fatalf("unexpected delete result: value=%q cursor=%d", value, cursor)
	}
}
