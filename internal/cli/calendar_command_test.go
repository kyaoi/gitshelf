package cli

import (
	"testing"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestStartOfWeek(t *testing.T) {
	value := time.Date(2026, 3, 11, 10, 0, 0, 0, time.Local)
	got := startOfWeek(value)
	if got.Format("2006-01-02") != "2026-03-09" {
		t.Fatalf("unexpected start of week: %s", got.Format("2006-01-02"))
	}
}

func TestBuildCalendarDays(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	tasks := []shelf.Task{
		{ID: "01A", Title: "A", DueOn: "2026-03-09"},
		{ID: "01B", Title: "B", DueOn: "2026-03-10"},
		{ID: "01C", Title: "C", DueOn: "2026-03-10"},
	}
	days := buildCalendarDays(tasks, start, 3)
	if len(days) != 3 {
		t.Fatalf("unexpected day count: %d", len(days))
	}
	if len(days[0].Tasks) != 1 || len(days[1].Tasks) != 2 || len(days[2].Tasks) != 0 {
		t.Fatalf("unexpected grouped calendar: %+v", days)
	}
}

func TestBuildCalendarMonthView(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	tasks := []shelf.Task{
		{ID: "01A", Title: "A", DueOn: "2026-03-09"},
		{ID: "01B", Title: "B", DueOn: "2026-03-10"},
		{ID: "01C", Title: "C", DueOn: "2026-03-10"},
	}
	days := buildCalendarDays(tasks, start, 14)
	month := buildCalendarMonthView(days, time.Date(2026, 3, 10, 0, 0, 0, 0, time.Local))

	if month.Label != "March 2026" {
		t.Fatalf("unexpected month label: %s", month.Label)
	}
	if len(month.Weeks) == 0 || len(month.Weeks[0]) != 7 {
		t.Fatalf("unexpected month grid shape: %+v", month.Weeks)
	}

	first := month.Weeks[0][0]
	if first.Date.Format("2006-01-02") != "2026-02-23" {
		t.Fatalf("unexpected grid start: %s", first.Date.Format("2006-01-02"))
	}

	found := false
	for _, week := range month.Weeks {
		for _, cell := range week {
			if cell.Date.Format("2006-01-02") != "2026-03-10" {
				continue
			}
			found = true
			if !cell.InRange {
				t.Fatalf("expected focused cell to be in range: %+v", cell)
			}
			if cell.TaskCount != 2 {
				t.Fatalf("expected task count 2, got %+v", cell)
			}
		}
	}
	if !found {
		t.Fatal("expected to find 2026-03-10 cell")
	}
}
