package shelf

import (
	"testing"
	"time"
)

func TestNormalizeDueOnKeywordsAndRelativeTokens(t *testing.T) {
	today := time.Now().Local().Format("2006-01-02")
	tomorrow := time.Now().Local().AddDate(0, 0, 1).Format("2006-01-02")
	nextWeek := time.Now().Local().AddDate(0, 0, 7).Format("2006-01-02")
	thisWeek := endOfWeekDate(time.Now().Local()).Format("2006-01-02")
	plus3 := time.Now().Local().AddDate(0, 0, 3).Format("2006-01-02")
	minus1 := time.Now().Local().AddDate(0, 0, -1).Format("2006-01-02")
	wantFri := nextWeekdayDate("fri", time.Now().Local()).Format("2006-01-02")
	wantNextMon := nextWeekdayDateStrict("mon", time.Now().Local()).Format("2006-01-02")
	wantIn4Days := time.Now().Local().AddDate(0, 0, 4).Format("2006-01-02")

	tests := []struct {
		input string
		want  string
	}{
		{input: "today", want: today},
		{input: "tomorrow", want: tomorrow},
		{input: "next-week", want: nextWeek},
		{input: "this-week", want: thisWeek},
		{input: "+3d", want: plus3},
		{input: "-1d", want: minus1},
		{input: "fri", want: wantFri},
		{input: "next-mon", want: wantNextMon},
		{input: "in 4 days", want: wantIn4Days},
	}
	for _, tc := range tests {
		got, err := NormalizeDueOn(tc.input)
		if err != nil {
			t.Fatalf("NormalizeDueOn(%q) failed: %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizeDueOn(%q) mismatch: got=%q want=%q", tc.input, got, tc.want)
		}
	}
}

func TestNormalizeDueOnRejectsInvalidToken(t *testing.T) {
	if _, err := NormalizeDueOn("next month"); err == nil {
		t.Fatal("expected invalid due_on error")
	}
}
