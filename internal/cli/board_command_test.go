package cli

import (
	"testing"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestBuildBoardColumns(t *testing.T) {
	statuses := []shelf.Status{"open", "done"}
	tasks := []shelf.Task{
		{ID: "01A", Title: "A", Status: "open"},
		{ID: "01B", Title: "B", Status: "done"},
		{ID: "01C", Title: "C", Status: "open"},
	}
	columns := buildBoardColumns(statuses, tasks)
	if len(columns) != 2 {
		t.Fatalf("unexpected column count: %d", len(columns))
	}
	if len(columns[0].Tasks) != 2 || len(columns[1].Tasks) != 1 {
		t.Fatalf("unexpected grouped columns: %+v", columns)
	}
}

func TestBoardCommandRoutesTTYToCockpit(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	oldTTY := dailyCockpitIsTTY
	oldRun := runCalendarModeTUIFn
	defer func() {
		dailyCockpitIsTTY = oldTTY
		runCalendarModeTUIFn = oldRun
	}()
	called := false
	dailyCockpitIsTTY = func() bool { return true }
	runCalendarModeTUIFn = func(rootDir string, startDate time.Time, daysCount int, statuses []shelf.Status, opts calendarTUIOptions) error {
		called = true
		if opts.Mode != calendarModeBoard {
			t.Fatalf("unexpected mode: %s", opts.Mode)
		}
		if len(statuses) == 0 {
			t.Fatal("expected configured statuses")
		}
		if len(opts.Filter.Statuses) != len(statuses) {
			t.Fatalf("unexpected filter statuses: %+v", opts.Filter.Statuses)
		}
		return nil
	}
	cmd := newBoardCommand(&commandContext{rootDir: root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("board command failed: %v", err)
	}
	if !called {
		t.Fatal("expected cockpit launcher to run")
	}
}
