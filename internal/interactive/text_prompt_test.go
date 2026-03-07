package interactive

import "testing"

func TestApplyTextPromptKeyAppendsAndEnterDone(t *testing.T) {
	value := ""
	for _, r := range []rune("Task q1") {
		done, canceled, next := applyTextPromptKey(value, keyEvent{Kind: keyKindRune, Rune: r})
		if done || canceled {
			t.Fatalf("unexpected terminal state while typing: done=%v canceled=%v", done, canceled)
		}
		value = next
	}
	if value != "Task q1" {
		t.Fatalf("unexpected value: %q", value)
	}

	done, canceled, next := applyTextPromptKey(value, keyEvent{Kind: keyKindEnter})
	if !done || canceled {
		t.Fatalf("enter should finish input: done=%v canceled=%v", done, canceled)
	}
	if next != value {
		t.Fatalf("enter should keep value: got %q want %q", next, value)
	}
}

func TestApplyTextPromptKeyBackspace(t *testing.T) {
	done, canceled, next := applyTextPromptKey("ab", keyEvent{Kind: keyKindBackspace})
	if done || canceled {
		t.Fatalf("backspace should not finish: done=%v canceled=%v", done, canceled)
	}
	if next != "a" {
		t.Fatalf("unexpected value after backspace: %q", next)
	}
}

func TestApplyTextPromptKeyCancelKeys(t *testing.T) {
	for _, key := range []keyKind{keyKindCtrlC, keyKindEsc} {
		done, canceled, next := applyTextPromptKey("abc", keyEvent{Kind: key})
		if done || !canceled {
			t.Fatalf("expected canceled for key %v: done=%v canceled=%v", key, done, canceled)
		}
		if next != "abc" {
			t.Fatalf("cancel should not alter value: %q", next)
		}
	}
}

func TestApplyTextPromptKeyUnicodeInputAndBackspace(t *testing.T) {
	value := ""
	for _, r := range []rune("日本語") {
		done, canceled, next := applyTextPromptKey(value, keyEvent{Kind: keyKindRune, Rune: r})
		if done || canceled {
			t.Fatalf("unexpected terminal state while typing unicode: done=%v canceled=%v", done, canceled)
		}
		value = next
	}
	if value != "日本語" {
		t.Fatalf("unexpected unicode value: %q", value)
	}

	_, _, value = applyTextPromptKey(value, keyEvent{Kind: keyKindBackspace})
	if value != "日本" {
		t.Fatalf("unexpected unicode value after backspace: %q", value)
	}
}
