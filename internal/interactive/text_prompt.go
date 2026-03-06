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

		b, err := reader.ReadByte()
		if err != nil {
			return "", fmt.Errorf("failed to read key input: %w", err)
		}

		done, canceled, next := applyTextPromptByte(value, b)
		if canceled {
			return "", ErrCanceled
		}
		if done {
			return strings.TrimSpace(next), nil
		}
		value = next
	}
}

func applyTextPromptByte(value string, b byte) (done bool, canceled bool, next string) {
	switch b {
	case 3:
		return false, true, value
	case 13, 10:
		return true, false, value
	case 27:
		return false, true, value
	case 8, 127:
		if len(value) == 0 {
			return false, false, value
		}
		_, size := utf8.DecodeLastRuneInString(value)
		return false, false, value[:len(value)-size]
	default:
		if b >= 32 {
			return false, false, value + string(b)
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
