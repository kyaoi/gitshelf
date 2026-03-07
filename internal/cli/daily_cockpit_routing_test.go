package cli

import (
	"testing"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestResolveReviewOutputMode(t *testing.T) {
	if got := resolveReviewOutputMode(true, false, false); got != dailyCockpitOutputTUI {
		t.Fatalf("expected tui mode, got %s", got)
	}
	if got := resolveReviewOutputMode(true, false, true); got != dailyCockpitOutputText {
		t.Fatalf("expected text mode for --plain, got %s", got)
	}
	if got := resolveReviewOutputMode(false, true, false); got != dailyCockpitOutputJSON {
		t.Fatalf("expected json mode, got %s", got)
	}
}

func TestResolveTodayOutputMode(t *testing.T) {
	if got := resolveTodayOutputMode(true, false, false, false); got != dailyCockpitOutputTUI {
		t.Fatalf("expected tui mode, got %s", got)
	}
	if got := resolveTodayOutputMode(true, false, false, true); got != dailyCockpitOutputText {
		t.Fatalf("expected text mode when carry-over is set, got %s", got)
	}
	if got := resolveTodayOutputMode(false, true, false, false); got != dailyCockpitOutputJSON {
		t.Fatalf("expected json mode, got %s", got)
	}
}

func TestResolveTreeOutputMode(t *testing.T) {
	if got := resolveTreeOutputMode(true, false, false); got != dailyCockpitOutputTUI {
		t.Fatalf("expected tui mode, got %s", got)
	}
	if got := resolveTreeOutputMode(true, false, true); got != dailyCockpitOutputText {
		t.Fatalf("expected text mode for --plain, got %s", got)
	}
	if got := resolveTreeOutputMode(false, true, false); got != dailyCockpitOutputJSON {
		t.Fatalf("expected json mode, got %s", got)
	}
}

func TestReviewCommandRoutesTTYToCockpit(t *testing.T) {
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
		if rootDir != root {
			t.Fatalf("unexpected root: %s", rootDir)
		}
		if opts.Mode != calendarModeReview {
			t.Fatalf("unexpected mode: %s", opts.Mode)
		}
		if opts.SectionLimit != 7 {
			t.Fatalf("unexpected section limit: %d", opts.SectionLimit)
		}
		if len(opts.Filter.Statuses) != 3 || opts.Filter.Statuses[0] != "open" {
			t.Fatalf("unexpected filter statuses: %+v", opts.Filter.Statuses)
		}
		return nil
	}
	cmd := newReviewCommand(&commandContext{rootDir: root})
	cmd.SetArgs([]string{"--limit", "7"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("review command failed: %v", err)
	}
	if !called {
		t.Fatal("expected calendar cockpit route")
	}
}

func TestTodayCommandRoutesTTYToCockpitWithFilter(t *testing.T) {
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
		if opts.Mode != calendarModeToday {
			t.Fatalf("unexpected mode: %s", opts.Mode)
		}
		if len(opts.Filter.Kinds) != 1 || opts.Filter.Kinds[0] != "todo" {
			t.Fatalf("unexpected kind filter: %+v", opts.Filter.Kinds)
		}
		if len(opts.Filter.NotStatuses) != 1 || opts.Filter.NotStatuses[0] != "blocked" {
			t.Fatalf("unexpected not-status filter: %+v", opts.Filter.NotStatuses)
		}
		if len(opts.Filter.Statuses) != 3 || opts.Filter.Statuses[1] != "in_progress" {
			t.Fatalf("unexpected statuses: %+v", opts.Filter.Statuses)
		}
		return nil
	}
	cmd := newTodayCommand(&commandContext{rootDir: root})
	cmd.SetArgs([]string{"--kind", "todo", "--not-status", "blocked"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("today command failed: %v", err)
	}
	if !called {
		t.Fatal("expected calendar cockpit route")
	}
}

func TestTodayCommandCarryOverStaysOnTextPath(t *testing.T) {
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
	dailyCockpitIsTTY = func() bool { return true }
	runCalendarModeTUIFn = func(rootDir string, startDate time.Time, daysCount int, statuses []shelf.Status, opts calendarTUIOptions) error {
		t.Fatal("carry-over path should not route to calendar cockpit")
		return nil
	}
	cmd := newTodayCommand(&commandContext{rootDir: root})
	cmd.SetArgs([]string{"--carry-over", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("today command failed: %v", err)
	}
}

func TestTreeCommandRoutesTTYToCockpit(t *testing.T) {
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
		if opts.Mode != calendarModeTree {
			t.Fatalf("unexpected mode: %s", opts.Mode)
		}
		if len(opts.Filter.Kinds) != 1 || opts.Filter.Kinds[0] != "todo" {
			t.Fatalf("unexpected kind filter: %+v", opts.Filter.Kinds)
		}
		return nil
	}
	cmd := newTreeCommand(&commandContext{rootDir: root})
	cmd.SetArgs([]string{"--kind", "todo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("tree command failed: %v", err)
	}
	if !called {
		t.Fatal("expected calendar cockpit route")
	}
}
