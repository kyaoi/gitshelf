package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

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
		if len(opts.Filter.Statuses) != 0 {
			t.Fatalf("board launcher should not force status filters: %+v", opts.Filter.Statuses)
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
		if opts.Mode != calendarModeReview {
			t.Fatalf("unexpected mode: %s", opts.Mode)
		}
		if opts.SectionLimit != 7 {
			t.Fatalf("unexpected section limit: %d", opts.SectionLimit)
		}
		if len(statuses) != 3 || statuses[0] != "open" || statuses[1] != "in_progress" || statuses[2] != "blocked" {
			t.Fatalf("unexpected active statuses: %+v", statuses)
		}
		if len(opts.Filter.Statuses) != 0 {
			t.Fatalf("review launcher should not inject a status filter: %+v", opts.Filter.Statuses)
		}
		return nil
	}
	cmd := newReviewCommand(&commandContext{rootDir: root})
	cmd.SetArgs([]string{"--limit", "7"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("review command failed: %v", err)
	}
	if !called {
		t.Fatal("expected cockpit launcher to run")
	}
}

func TestNowCommandRoutesTTYToCockpitWithFilter(t *testing.T) {
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
		if opts.Mode != calendarModeNow {
			t.Fatalf("unexpected mode: %s", opts.Mode)
		}
		if len(opts.Filter.Kinds) != 1 || opts.Filter.Kinds[0] != "todo" {
			t.Fatalf("unexpected kind filter: %+v", opts.Filter.Kinds)
		}
		if len(opts.Filter.NotStatuses) != 1 || opts.Filter.NotStatuses[0] != "blocked" {
			t.Fatalf("unexpected not-status filter: %+v", opts.Filter.NotStatuses)
		}
		if len(statuses) != 3 || statuses[1] != "in_progress" {
			t.Fatalf("unexpected active statuses: %+v", statuses)
		}
		return nil
	}
	cmd := newNowCommand(&commandContext{rootDir: root})
	cmd.SetArgs([]string{"--kind", "todo", "--not-status", "blocked"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("now command failed: %v", err)
	}
	if !called {
		t.Fatal("expected cockpit launcher to run")
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
		t.Fatal("expected cockpit launcher to run")
	}
}

func TestLaunchersRequireTTY(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	oldTTY := dailyCockpitIsTTY
	defer func() { dailyCockpitIsTTY = oldTTY }()
	dailyCockpitIsTTY = func() bool { return false }

	cases := []struct {
		args []string
		want string
	}{
		{[]string{"cockpit", "--root", root}, "cockpit はTTYが必要です"},
		{[]string{"calendar", "--root", root}, "calendar はTTYが必要です"},
		{[]string{"tree", "--root", root}, "tree はTTYが必要です"},
		{[]string{"board", "--root", root}, "board はTTYが必要です"},
		{[]string{"review", "--root", root}, "review はTTYが必要です"},
		{[]string{"now", "--root", root}, "now はTTYが必要です"},
	}
	for _, tc := range cases {
		cmd := NewRootCommand("test")
		cmd.SetArgs(tc.args)
		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%v: expected %q, got %v", tc.args, tc.want, err)
		}
	}
}
