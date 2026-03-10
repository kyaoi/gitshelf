package interactive

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func PromptText(prompt string) (string, error) {
	if !IsTTY() {
		return "", ErrNonTTY
	}

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", fmt.Errorf("failed to switch terminal to raw mode: %w", err)
	}
	defer func() {
		_ = term.Restore(fd, oldState)
		fmt.Fprint(os.Stdout, eol)
	}()

	reader := bufio.NewReader(os.Stdin)
	value := ""
	cursor := 0

	for {
		renderTextPrompt(prompt, value, cursor)

		key, err := readKeyEvent(reader)
		if err != nil {
			return "", fmt.Errorf("failed to read key input: %w", err)
		}

		done, canceled, next, nextCursor := applyTextPromptKey(value, cursor, key)
		if canceled {
			return "", ErrCanceled
		}
		if done {
			return strings.TrimSpace(next), nil
		}
		value = next
		cursor = nextCursor
	}
}

func applyTextPromptKey(value string, cursor int, key keyEvent) (done bool, canceled bool, next string, nextCursor int) {
	switch key.Kind {
	case keyKindCtrlC, keyKindEsc:
		return false, true, value, ClampTextCursor(value, cursor)
	case keyKindEnter:
		return true, false, value, ClampTextCursor(value, cursor)
	case keyKindBackspace:
		nextValue, nextCursor := DeleteRuneBeforeCursor(value, cursor)
		return false, false, nextValue, nextCursor
	case keyKindLeft:
		return false, false, value, MoveTextCursorLeft(value, cursor)
	case keyKindRight:
		return false, false, value, MoveTextCursorRight(value, cursor)
	default:
		if key.Kind == keyKindRune && isPrintableRune(key.Rune) {
			nextValue, nextCursor := InsertRuneAtCursor(value, cursor, key.Rune)
			return false, false, nextValue, nextCursor
		}
		return false, false, value, ClampTextCursor(value, cursor)
	}
}

func renderTextPrompt(prompt string, value string, cursor int) {
	var b strings.Builder
	b.WriteString("\r\033[H\033[2J")
	b.WriteString(uiPrompt(prompt))
	b.WriteString(eol)
	b.WriteString(uiHelp("入力して Enter で確定  Esc/Ctrl+C でキャンセル"))
	b.WriteString(eol)
	b.WriteString(eol)
	b.WriteString(uiSelected("入力: " + RenderTextCursor(value, cursor)))
	b.WriteString(eol)
	fmt.Fprint(os.Stdout, b.String())
}
