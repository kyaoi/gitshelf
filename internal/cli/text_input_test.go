package cli

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestApplyBubbleTextInputKeySupportsSpaceHomeEndAndDelete(t *testing.T) {
	value, cursor, handled := applyBubbleTextInputKey("ab", 1, tea.KeyMsg{Type: tea.KeySpace})
	if !handled || value != "a b" || cursor != 2 {
		t.Fatalf("unexpected space insert: handled=%v value=%q cursor=%d", handled, value, cursor)
	}

	value, cursor, handled = applyBubbleTextInputKey(value, cursor, tea.KeyMsg{Type: tea.KeyHome})
	if !handled || cursor != 0 {
		t.Fatalf("unexpected home handling: handled=%v cursor=%d", handled, cursor)
	}

	value, cursor, handled = applyBubbleTextInputKey(value, cursor, tea.KeyMsg{Type: tea.KeyEnd})
	if !handled || cursor != len([]rune(value)) {
		t.Fatalf("unexpected end handling: handled=%v cursor=%d", handled, cursor)
	}

	value, cursor, handled = applyBubbleTextInputKey("abcd", 1, tea.KeyMsg{Type: tea.KeyDelete})
	if !handled || value != "acd" || cursor != 1 {
		t.Fatalf("unexpected delete handling: handled=%v value=%q cursor=%d", handled, value, cursor)
	}
}
