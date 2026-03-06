package cli

import "testing"

func TestApplyByDays(t *testing.T) {
	got, err := applyByDays("2026-03-10", "2d")
	if err != nil {
		t.Fatalf("applyByDays failed: %v", err)
	}
	if got != "2026-03-12" {
		t.Fatalf("unexpected due: %s", got)
	}

	got, err = applyByDays("", "-1d")
	if err != nil {
		t.Fatalf("applyByDays with empty due failed: %v", err)
	}
	if got == "" {
		t.Fatal("expected non-empty due")
	}
}

func TestApplyByDaysRejectsInvalid(t *testing.T) {
	if _, err := applyByDays("2026-03-10", "2"); err == nil {
		t.Fatal("expected invalid --by error")
	}
}
