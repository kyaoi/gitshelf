package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kyaoi/gitshelf/internal/interactive"
)

func applyBubbleTextInputKey(value string, cursor int, msg tea.KeyMsg) (string, int, bool) {
	switch msg.String() {
	case "backspace":
		nextValue, nextCursor := interactive.DeleteRuneBeforeCursor(value, cursor)
		return nextValue, nextCursor, true
	case "delete":
		nextValue, nextCursor := interactive.DeleteRuneAtCursor(value, cursor)
		return nextValue, nextCursor, true
	case "left":
		return value, interactive.MoveTextCursorLeft(value, cursor), true
	case "right":
		return value, interactive.MoveTextCursorRight(value, cursor), true
	case "home", "ctrl+a":
		return value, interactive.MoveTextCursorStart(value, cursor), true
	case "end", "ctrl+e":
		return value, interactive.MoveTextCursorEnd(value, cursor), true
	case " ", "space":
		nextValue, nextCursor := interactive.InsertRuneAtCursor(value, cursor, ' ')
		return nextValue, nextCursor, true
	default:
		if msg.Type == tea.KeyRunes {
			nextValue := value
			nextCursor := cursor
			for _, r := range msg.Runes {
				nextValue, nextCursor = interactive.InsertRuneAtCursor(nextValue, nextCursor, r)
			}
			return nextValue, nextCursor, true
		}
		return value, interactive.ClampTextCursor(value, cursor), false
	}
}
