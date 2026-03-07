package cli

import "testing"

func TestParseTagFlagValues(t *testing.T) {
	got := parseTagFlagValues([]string{"backend", "urgent,review", " backend "})
	if len(got) != 3 {
		t.Fatalf("unexpected parsed tags: %+v", got)
	}
	if got[0] != "backend" || got[1] != "urgent" || got[2] != "review" {
		t.Fatalf("unexpected parsed tags: %+v", got)
	}
}

func TestFormatTagSummary(t *testing.T) {
	if got := formatTagSummary(nil); got != "(none)" {
		t.Fatalf("unexpected empty summary: %q", got)
	}
	if got := formatTagSummary([]string{"a", "b"}); got != "a, b" {
		t.Fatalf("unexpected summary: %q", got)
	}
}
