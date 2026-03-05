package interactive

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

var ErrCanceled = errors.New("selection canceled")

type Option struct {
	Value      string
	Label      string
	SearchText string
}

func Select(prompt string, options []Option) (Option, error) {
	if len(options) == 0 {
		return Option{}, errors.New("no options to select")
	}
	if !IsTTY() {
		return Option{}, ErrNonTTY
	}

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return Option{}, fmt.Errorf("failed to switch terminal to raw mode: %w", err)
	}
	defer func() {
		_ = term.Restore(fd, oldState)
		fmt.Fprint(os.Stdout, "\n")
	}()

	reader := bufio.NewReader(os.Stdin)

	search := ""
	searchMode := false
	cursor := 0

	for {
		filtered := filterOptions(options, search)
		if len(filtered) == 0 {
			cursor = 0
		} else if cursor >= len(filtered) {
			cursor = len(filtered) - 1
		}

		render(prompt, filtered, cursor, search, searchMode)

		b, err := reader.ReadByte()
		if err != nil {
			return Option{}, fmt.Errorf("failed to read key input: %w", err)
		}

		switch b {
		case 3:
			return Option{}, ErrCanceled
		case 13, 10:
			if searchMode {
				searchMode = false
				continue
			}
			if len(filtered) == 0 {
				continue
			}
			return filtered[cursor], nil
		case 27:
			seq1, _ := reader.Peek(1)
			if len(seq1) == 1 && seq1[0] == '[' {
				_, _ = reader.ReadByte()
				seq2, _ := reader.ReadByte()
				switch seq2 {
				case 'A':
					cursor = moveUp(cursor, len(filtered))
				case 'B':
					cursor = moveDown(cursor, len(filtered))
				}
				continue
			}
			if searchMode {
				search = ""
				searchMode = false
				cursor = 0
				continue
			}
			return Option{}, ErrCanceled
		case '/':
			searchMode = true
			cursor = 0
		case 127:
			if searchMode && len(search) > 0 {
				_, size := utf8.DecodeLastRuneInString(search)
				search = search[:len(search)-size]
				cursor = 0
			}
		case 'j':
			if searchMode {
				search += string(b)
				cursor = 0
				continue
			}
			cursor = moveDown(cursor, len(filtered))
		case 'k':
			if searchMode {
				search += string(b)
				cursor = 0
				continue
			}
			cursor = moveUp(cursor, len(filtered))
		default:
			if searchMode && b >= 32 {
				search += string(b)
				cursor = 0
			}
		}
	}
}

func render(prompt string, options []Option, cursor int, search string, searchMode bool) {
	var b strings.Builder
	b.WriteString("\033[H\033[2J")
	b.WriteString(prompt)
	b.WriteString("\n")
	b.WriteString("j/k: 移動  Enter: 決定  /: 検索  Esc/Ctrl+C: キャンセル\n")

	if searchMode {
		b.WriteString(fmt.Sprintf("検索: %s_\n", search))
	} else if search != "" {
		b.WriteString(fmt.Sprintf("検索: %s\n", search))
	} else {
		b.WriteString("検索: (なし)\n")
	}

	max := min(len(options), 15)
	for i := 0; i < max; i++ {
		prefix := "  "
		if i == cursor {
			prefix = "> "
		}
		b.WriteString(prefix + options[i].Label + "\n")
	}
	if len(options) == 0 {
		b.WriteString("(候補なし)\n")
	}
	fmt.Fprint(os.Stdout, b.String())
}

func filterOptions(options []Option, query string) []Option {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return slices.Clone(options)
	}
	filtered := make([]Option, 0, len(options))
	for _, option := range options {
		target := option.SearchText
		if target == "" {
			target = option.Label
		}
		if strings.Contains(strings.ToLower(target), query) {
			filtered = append(filtered, option)
		}
	}
	return filtered
}

func moveUp(cursor, length int) int {
	if length == 0 {
		return 0
	}
	if cursor <= 0 {
		return length - 1
	}
	return cursor - 1
}

func moveDown(cursor, length int) int {
	if length == 0 {
		return 0
	}
	if cursor >= length-1 {
		return 0
	}
	return cursor + 1
}
