package interactive

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

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

	for {
		renderTextPrompt(prompt, value)

		key, err := readKeyEvent(reader)
		if err != nil {
			return "", fmt.Errorf("failed to read key input: %w", err)
		}

		done, canceled, next := applyTextPromptKey(value, key)
		if canceled {
			return "", ErrCanceled
		}
		if done {
			return strings.TrimSpace(next), nil
		}
		value = next
	}
}

func applyTextPromptKey(value string, key keyEvent) (done bool, canceled bool, next string) {
	switch key.Kind {
	case keyKindCtrlC, keyKindEsc:
		return false, true, value
	case keyKindEnter:
		return true, false, value
	case keyKindBackspace:
		if len(value) == 0 {
			return false, false, value
		}
		_, size := utf8.DecodeLastRuneInString(value)
		return false, false, value[:len(value)-size]
	default:
		if key.Kind == keyKindRune && isPrintableRune(key.Rune) {
			return false, false, value + string(key.Rune)
		}
		return false, false, value
	}
}

func renderTextPrompt(prompt string, value string) {
	var b strings.Builder
	b.WriteString("\r\033[H\033[2J")
	b.WriteString(uiPrompt(prompt))
	b.WriteString(eol)
	b.WriteString(uiHelp("入力して Enter で確定  Esc/Ctrl+C でキャンセル"))
	b.WriteString(eol)
	b.WriteString(eol)
	b.WriteString(uiSelected("入力: " + value + "_"))
	b.WriteString(eol)
	fmt.Fprint(os.Stdout, b.String())
}
