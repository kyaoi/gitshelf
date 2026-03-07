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

const eol = "\r\n"
const defaultMaxRows = 15
const defaultHelpText = "j/k: 移動  Enter: 決定  /: 検索  ?: ヘルプ  q/Esc/Ctrl+C: キャンセル"
const defaultSearchPlaceholder = "検索"

type Option struct {
	Value      string
	Label      string
	SearchText string
	Preview    string
}

func Select(prompt string, options []Option) (Option, error) {
	return SelectWithConfig(SelectConfig{
		Prompt:            prompt,
		Options:           options,
		ShowPreview:       true,
		MaxRows:           defaultMaxRows,
		HelpText:          defaultHelpText,
		SearchPlaceholder: defaultSearchPlaceholder,
	})
}

type SelectConfig struct {
	Prompt            string
	Options           []Option
	ShowPreview       bool
	MaxRows           int
	HelpText          string
	SearchPlaceholder string
	SubmitValue       string
	SubmitShortcuts   []string
}

func SelectWithConfig(cfg SelectConfig) (Option, error) {
	if len(cfg.Options) == 0 {
		return Option{}, errors.New("no options to select")
	}
	if !IsTTY() {
		return Option{}, ErrNonTTY
	}
	if cfg.MaxRows <= 0 {
		cfg.MaxRows = defaultMaxRows
	}
	if strings.TrimSpace(cfg.HelpText) == "" {
		cfg.HelpText = defaultHelpText
	}
	if strings.TrimSpace(cfg.SearchPlaceholder) == "" {
		cfg.SearchPlaceholder = defaultSearchPlaceholder
	}

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return Option{}, fmt.Errorf("failed to switch terminal to raw mode: %w", err)
	}
	defer func() {
		_ = term.Restore(fd, oldState)
		fmt.Fprint(os.Stdout, eol)
	}()

	reader := bufio.NewReader(os.Stdin)

	search := ""
	searchMode := false
	showHelp := false
	cursor := 0
	offset := 0

	for {
		filtered := filterOptions(cfg.Options, search)
		if len(filtered) == 0 {
			cursor = 0
			offset = 0
		} else if cursor >= len(filtered) {
			cursor = len(filtered) - 1
		}
		visibleRows, previewRows := resolveSelectorLayout(cfg, filtered, cursor, showHelp)
		offset = clampSelectorOffset(offset, cursor, len(filtered), visibleRows)

		render(cfg, filtered, cursor, offset, visibleRows, previewRows, search, searchMode, showHelp)

		key, err := readKeyEvent(reader)
		if err != nil {
			return Option{}, fmt.Errorf("failed to read key input: %w", err)
		}

		switch key.Kind {
		case keyKindCtrlC:
			return Option{}, ErrCanceled
		case keyKindCtrlS, keyKindCtrlEnter:
			if matched := matchSubmitShortcut(cfg, key); matched {
				option, ok := findOptionByValue(cfg.Options, cfg.SubmitValue)
				if ok {
					return option, nil
				}
			}
		case keyKindEnter:
			if searchMode {
				searchMode = false
				continue
			}
			if len(filtered) == 0 {
				continue
			}
			return filtered[cursor], nil
		case keyKindEsc:
			if searchMode {
				search = ""
				searchMode = false
				cursor = 0
				offset = 0
				continue
			}
			return Option{}, ErrCanceled
		case keyKindUp:
			cursor = moveUp(cursor, len(filtered))
		case keyKindDown:
			cursor = moveDown(cursor, len(filtered))
		case keyKindBackspace:
			if searchMode && len(search) > 0 {
				_, size := utf8.DecodeLastRuneInString(search)
				search = search[:len(search)-size]
				cursor = 0
				offset = 0
			}
		case keyKindRune:
			r := key.Rune
			if searchMode {
				if isPrintableRune(r) {
					search += string(r)
					cursor = 0
					offset = 0
				}
				continue
			}

			switch r {
			case '/':
				searchMode = true
				cursor = 0
			case '?':
				showHelp = !showHelp
			case 'q':
				return Option{}, ErrCanceled
			case 'j':
				cursor = moveDown(cursor, len(filtered))
			case 'k':
				cursor = moveUp(cursor, len(filtered))
			default:
				if searchMode && isPrintableRune(r) {
					search += string(r)
					cursor = 0
					offset = 0
				}
			}
		default:
			if key.Kind == keyKindRune && searchMode && isPrintableRune(key.Rune) {
				search += string(key.Rune)
				cursor = 0
				offset = 0
			}
		}
	}
}

func matchSubmitShortcut(cfg SelectConfig, key keyEvent) bool {
	if strings.TrimSpace(cfg.SubmitValue) == "" || len(cfg.SubmitShortcuts) == 0 {
		return false
	}
	for _, shortcut := range cfg.SubmitShortcuts {
		switch strings.ToLower(strings.TrimSpace(shortcut)) {
		case "ctrl+s":
			if key.Kind == keyKindCtrlS {
				return true
			}
		case "ctrl+enter":
			if key.Kind == keyKindCtrlEnter {
				return true
			}
		}
	}
	return false
}

func findOptionByValue(options []Option, value string) (Option, bool) {
	for _, option := range options {
		if option.Value == value {
			return option, true
		}
	}
	return Option{}, false
}

func render(cfg SelectConfig, options []Option, cursor int, offset int, visibleRows int, previewRows int, search string, searchMode bool, showHelp bool) {
	var b strings.Builder
	b.WriteString("\r\033[H\033[2J")
	b.WriteString(uiPrompt(cfg.Prompt))
	b.WriteString(eol)
	b.WriteString(uiHelp(cfg.HelpText))
	b.WriteString(eol)
	if showHelp {
		b.WriteString(uiHelp("↑/↓ でも移動できます。検索中は通常文字が検索語に追加されます。"))
		b.WriteString(eol)
	}

	if searchMode {
		b.WriteString(uiSearch(selectorSearchLine(cfg.SearchPlaceholder, search+"_", offset, visibleRows, len(options))))
		b.WriteString(eol)
	} else if search != "" {
		b.WriteString(uiSearch(selectorSearchLine(cfg.SearchPlaceholder, search, offset, visibleRows, len(options))))
		b.WriteString(eol)
	} else {
		b.WriteString(uiHelp(selectorSearchLine(cfg.SearchPlaceholder, "(なし)", offset, visibleRows, len(options))))
		b.WriteString(eol)
	}

	end := min(len(options), offset+visibleRows)
	for i := offset; i < end; i++ {
		prefix := "  "
		label := options[i].Label
		if i == cursor {
			prefix = uiColor("> ", "1;38;5;45")
			label = uiSelected(label)
		}
		b.WriteString(prefix + label + eol)
	}
	if len(options) == 0 {
		b.WriteString(uiHelp("(候補なし)"))
		b.WriteString(eol)
	} else if cfg.ShowPreview && cursor >= 0 && cursor < len(options) {
		preview := strings.TrimSpace(options[cursor].Preview)
		if preview != "" && previewRows > 0 {
			b.WriteString(eol)
			b.WriteString(uiPreviewHeader("----- preview -----"))
			b.WriteString(eol)
			lines := strings.Split(preview, "\n")
			maxPreviewLines := min(len(lines), previewRows)
			for i := 0; i < maxPreviewLines; i++ {
				b.WriteString(lines[i])
				b.WriteString(eol)
			}
			if len(lines) > maxPreviewLines {
				b.WriteString("...")
				b.WriteString(eol)
			}
		}
	}
	fmt.Fprint(os.Stdout, b.String())
}

func selectorSearchLine(placeholder string, value string, offset int, visibleRows int, total int) string {
	line := fmt.Sprintf("%s: %s", placeholder, value)
	if total <= 0 || visibleRows <= 0 {
		return line
	}
	start := offset + 1
	end := min(total, offset+visibleRows)
	return fmt.Sprintf("%s  [%d-%d/%d]", line, start, end, total)
}

func resolveSelectorLayout(cfg SelectConfig, options []Option, cursor int, showHelp bool) (int, int) {
	visibleRows := min(len(options), cfg.MaxRows)
	if visibleRows <= 0 {
		visibleRows = 1
	}

	hasPreview := cfg.ShowPreview && cursor >= 0 && cursor < len(options) && strings.TrimSpace(options[cursor].Preview) != ""
	if !hasPreview {
		return visibleRows, 0
	}

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	_ = width
	if err != nil || height <= 0 {
		return visibleRows, 8
	}

	reservedLines := 3 // prompt, help, search
	if showHelp {
		reservedLines++
	}
	available := height - reservedLines
	if available <= 1 {
		return 1, 0
	}

	const (
		minOptionRows      = 4
		defaultPreviewRows = 8
		previewOverhead    = 2 // blank line + preview header
	)

	maxOptionRows := available
	previewRows := 0
	if available > minOptionRows+previewOverhead {
		previewBudget := min(defaultPreviewRows, available-minOptionRows-previewOverhead)
		if previewBudget > 0 {
			previewRows = previewBudget
			maxOptionRows = available - previewOverhead - previewRows
		}
	}
	if maxOptionRows <= 0 {
		maxOptionRows = 1
	}
	return min(visibleRows, maxOptionRows), previewRows
}

func clampSelectorOffset(offset int, cursor int, total int, visibleRows int) int {
	if total <= 0 || visibleRows <= 0 {
		return 0
	}
	if visibleRows >= total {
		return 0
	}
	maxOffset := total - visibleRows
	if offset > maxOffset {
		offset = maxOffset
	}
	if cursor < offset {
		offset = cursor
	}
	if cursor >= offset+visibleRows {
		offset = cursor - visibleRows + 1
	}
	if offset < 0 {
		return 0
	}
	if offset > maxOffset {
		return maxOffset
	}
	return offset
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
