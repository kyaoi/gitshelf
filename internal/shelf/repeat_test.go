package shelf

import (
	"testing"
	"time"
)

func TestAdvanceDueByRepeat(t *testing.T) {
	now := time.Date(2026, 3, 6, 10, 0, 0, 0, time.Local)

	cases := []struct {
		current string
		repeat  string
		want    string
	}{
		{"2026-03-10", "1d", "2026-03-11"},
		{"2026-03-10", "2w", "2026-03-24"},
		{"2026-03-10", "1m", "2026-04-10"},
		{"2026-03-10", "1y", "2027-03-10"},
		{"", "1d", "2026-03-07"},
	}
	for _, tc := range cases {
		got, err := AdvanceDueByRepeat(tc.current, tc.repeat, now)
		if err != nil {
			t.Fatalf("advance failed: current=%q repeat=%q err=%v", tc.current, tc.repeat, err)
		}
		if got != tc.want {
			t.Fatalf("unexpected next due: current=%q repeat=%q got=%q want=%q", tc.current, tc.repeat, got, tc.want)
		}
	}
}

func TestAdvanceDueByRepeatRejectsInvalid(t *testing.T) {
	if _, err := AdvanceDueByRepeat("2026-03-10", "bad", time.Now()); err == nil {
		t.Fatal("expected invalid repeat error")
	}
}
