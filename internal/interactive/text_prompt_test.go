package interactive

import "testing"

func TestApplyTextPromptByteAppendsAndEnterDone(t *testing.T) {
	value := ""
	for _, b := range []byte("Task q1") {
		done, canceled, next := applyTextPromptByte(value, b)
		if done || canceled {
			t.Fatalf("unexpected terminal state while typing: done=%v canceled=%v", done, canceled)
		}
		value = next
	}
	if value != "Task q1" {
		t.Fatalf("unexpected value: %q", value)
	}

	done, canceled, next := applyTextPromptByte(value, '\n')
	if !done || canceled {
		t.Fatalf("enter should finish input: done=%v canceled=%v", done, canceled)
	}
	if next != value {
		t.Fatalf("enter should keep value: got %q want %q", next, value)
	}
}

func TestApplyTextPromptByteBackspace(t *testing.T) {
	done, canceled, next := applyTextPromptByte("ab", 127)
	if done || canceled {
		t.Fatalf("backspace should not finish: done=%v canceled=%v", done, canceled)
	}
	if next != "a" {
		t.Fatalf("unexpected value after backspace: %q", next)
	}
}

func TestApplyTextPromptByteCancelKeys(t *testing.T) {
	for _, key := range []byte{3, 27} {
		done, canceled, next := applyTextPromptByte("abc", key)
		if done || !canceled {
			t.Fatalf("expected canceled for key %d: done=%v canceled=%v", key, done, canceled)
		}
		if next != "abc" {
			t.Fatalf("cancel should not alter value: %q", next)
		}
	}
}
