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
	reader := bufio.NewReader(strings.NewReader("\x1b[A\x1b[B"))

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
