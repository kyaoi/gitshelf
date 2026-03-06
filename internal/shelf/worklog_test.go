package shelf

import (
	"testing"
	"time"
)

func TestParseWorkDurationMinutes(t *testing.T) {
	mins, err := ParseWorkDurationMinutes("2h30m")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if mins != 150 {
		t.Fatalf("unexpected minutes: %d", mins)
	}
}

func TestElapsedMinutesSince(t *testing.T) {
	now := time.Date(2026, 3, 7, 12, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	mins, err := ElapsedMinutesSince("2026-03-07T10:30:00+09:00", now)
	if err != nil {
		t.Fatalf("elapsed failed: %v", err)
	}
	if mins != 90 {
		t.Fatalf("unexpected elapsed minutes: %d", mins)
	}
}
