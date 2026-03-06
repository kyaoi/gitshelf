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

func TestResolveSnoozeMode(t *testing.T) {
	mode, err := resolveSnoozeMode(true, false, false)
	if err != nil || mode != snoozeModeBy {
		t.Fatalf("expected by mode, got mode=%q err=%v", mode, err)
	}

	mode, err = resolveSnoozeMode(false, true, false)
	if err != nil || mode != snoozeModeTo {
		t.Fatalf("expected to mode, got mode=%q err=%v", mode, err)
	}

	if _, err := resolveSnoozeMode(true, true, false); err == nil {
		t.Fatal("expected conflict error")
	}
	if _, err := resolveSnoozeMode(false, false, false); err == nil {
		t.Fatal("expected missing value error on non-tty")
	}
	mode, err = resolveSnoozeMode(false, false, true)
	if err != nil || mode != "" {
		t.Fatalf("expected interactive fallback, got mode=%q err=%v", mode, err)
	}
}
