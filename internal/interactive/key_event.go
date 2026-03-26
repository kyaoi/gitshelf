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
	keyKindCtrlS
	keyKindCtrlEnter
	keyKindBackspace
	keyKindDelete
	keyKindUp
	keyKindDown
	keyKindLeft
	keyKindRight
	keyKindHome
	keyKindEnd
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
	case 19:
		return keyEvent{Kind: keyKindCtrlS}, nil
	case 13, 10:
		return keyEvent{Kind: keyKindEnter}, nil
	case 8, 127:
		return keyEvent{Kind: keyKindBackspace}, nil
	case 27:
		if matched, event, err := readCSIKeyEvent(reader); matched {
			return event, err
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

func readCSIKeyEvent(reader *bufio.Reader) (bool, keyEvent, error) {
	if reader.Buffered() < 2 {
		return false, keyEvent{}, nil
	}
	seq, err := reader.Peek(reader.Buffered())
	if err != nil || len(seq) < 2 || seq[0] != '[' {
		return false, keyEvent{}, err
	}
	switch {
	case len(seq) >= 2 && seq[1] == 'A':
		_, _ = reader.Discard(2)
		return true, keyEvent{Kind: keyKindUp}, nil
	case len(seq) >= 2 && seq[1] == 'B':
		_, _ = reader.Discard(2)
		return true, keyEvent{Kind: keyKindDown}, nil
	case len(seq) >= 2 && seq[1] == 'C':
		_, _ = reader.Discard(2)
		return true, keyEvent{Kind: keyKindRight}, nil
	case len(seq) >= 2 && seq[1] == 'D':
		_, _ = reader.Discard(2)
		return true, keyEvent{Kind: keyKindLeft}, nil
	case len(seq) >= 2 && seq[1] == 'H':
		_, _ = reader.Discard(2)
		return true, keyEvent{Kind: keyKindHome}, nil
	case len(seq) >= 2 && seq[1] == 'F':
		_, _ = reader.Discard(2)
		return true, keyEvent{Kind: keyKindEnd}, nil
	}

	for _, pattern := range []string{"[13;5u", "[13;5~", "[27;5;13~"} {
		if len(seq) >= len(pattern) && string(seq[:len(pattern)]) == pattern {
			_, _ = reader.Discard(len(pattern))
			return true, keyEvent{Kind: keyKindCtrlEnter}, nil
		}
	}
	for _, pattern := range []struct {
		seq  string
		kind keyKind
	}{
		{seq: "[1~", kind: keyKindHome},
		{seq: "[7~", kind: keyKindHome},
		{seq: "[4~", kind: keyKindEnd},
		{seq: "[8~", kind: keyKindEnd},
		{seq: "[3~", kind: keyKindDelete},
	} {
		if len(seq) >= len(pattern.seq) && string(seq[:len(pattern.seq)]) == pattern.seq {
			_, _ = reader.Discard(len(pattern.seq))
			return true, keyEvent{Kind: pattern.kind}, nil
		}
	}
	return false, keyEvent{}, nil
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
