package interactive

import (
	"bufio"
	"fmt"
	"unicode"
	"unicode/utf8"
)

type keyKind int

const (
	keyKindRune keyKind = iota
	keyKindEnter
	keyKindEsc
	keyKindCtrlC
	keyKindBackspace
	keyKindUp
	keyKindDown
)

type keyEvent struct {
	Kind keyKind
	Rune rune
}

func readKeyEvent(reader *bufio.Reader) (keyEvent, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return keyEvent{}, err
	}

	switch b {
	case 3:
		return keyEvent{Kind: keyKindCtrlC}, nil
	case 13, 10:
		return keyEvent{Kind: keyKindEnter}, nil
	case 8, 127:
		return keyEvent{Kind: keyKindBackspace}, nil
	case 27:
		// Handle CSI arrows when already buffered to avoid blocking on plain Esc.
		if reader.Buffered() >= 2 {
			seq, err := reader.Peek(2)
			if err == nil && seq[0] == '[' {
				switch seq[1] {
				case 'A':
					_, _ = reader.Discard(2)
					return keyEvent{Kind: keyKindUp}, nil
				case 'B':
					_, _ = reader.Discard(2)
					return keyEvent{Kind: keyKindDown}, nil
				}
			}
		}
		return keyEvent{Kind: keyKindEsc}, nil
	default:
		r, err := decodeRuneFromFirstByte(reader, b)
		if err != nil {
			return keyEvent{}, err
		}
		return keyEvent{Kind: keyKindRune, Rune: r}, nil
	}
}

func decodeRuneFromFirstByte(reader *bufio.Reader, first byte) (rune, error) {
	if first < utf8.RuneSelf {
		return rune(first), nil
	}

	buf := []byte{first}
	for len(buf) < utf8.UTFMax {
		if utf8.FullRune(buf) {
			r, _ := utf8.DecodeRune(buf)
			return r, nil
		}
		next, err := reader.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("failed to read utf-8 continuation byte: %w", err)
		}
		buf = append(buf, next)
	}

	r, _ := utf8.DecodeRune(buf)
	return r, nil
}

func isPrintableRune(r rune) bool {
	if r == utf8.RuneError {
		return false
	}
	return !unicode.IsControl(r)
}
