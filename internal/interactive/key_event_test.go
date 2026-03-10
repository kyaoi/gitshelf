package interactive

import (
	"bufio"
	"strings"
	"testing"
)

func TestReadKeyEventUnicodeRune(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("あ"))
	got, err := readKeyEvent(reader)
	if err != nil {
		t.Fatalf("readKeyEvent failed: %v", err)
	}
	if got.Kind != keyKindRune || got.Rune != 'あ' {
		t.Fatalf("unexpected key event: %+v", got)
	}
}

func TestReadKeyEventArrowKeys(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\x1b[A\x1b[B\x1b[C\x1b[D"))

	up, err := readKeyEvent(reader)
	if err != nil {
		t.Fatalf("readKeyEvent up failed: %v", err)
	}
	if up.Kind != keyKindUp {
		t.Fatalf("expected up, got %+v", up)
	}

	down, err := readKeyEvent(reader)
	if err != nil {
		t.Fatalf("readKeyEvent down failed: %v", err)
	}
	if down.Kind != keyKindDown {
		t.Fatalf("expected down, got %+v", down)
	}

	right, err := readKeyEvent(reader)
	if err != nil {
		t.Fatalf("readKeyEvent right failed: %v", err)
	}
	if right.Kind != keyKindRight {
		t.Fatalf("expected right, got %+v", right)
	}

	left, err := readKeyEvent(reader)
	if err != nil {
		t.Fatalf("readKeyEvent left failed: %v", err)
	}
	if left.Kind != keyKindLeft {
		t.Fatalf("expected left, got %+v", left)
	}
}

func TestReadKeyEventEsc(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\x1b"))
	got, err := readKeyEvent(reader)
	if err != nil {
		t.Fatalf("readKeyEvent failed: %v", err)
	}
	if got.Kind != keyKindEsc {
		t.Fatalf("expected esc, got %+v", got)
	}
}

func TestReadKeyEventCtrlS(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(string([]byte{19})))
	got, err := readKeyEvent(reader)
	if err != nil {
		t.Fatalf("readKeyEvent failed: %v", err)
	}
	if got.Kind != keyKindCtrlS {
		t.Fatalf("expected ctrl+s, got %+v", got)
	}
}

func TestReadKeyEventCtrlEnterCSI(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\x1b[13;5u"))
	got, err := readKeyEvent(reader)
	if err != nil {
		t.Fatalf("readKeyEvent failed: %v", err)
	}
	if got.Kind != keyKindCtrlEnter {
		t.Fatalf("expected ctrl+enter, got %+v", got)
	}
}
