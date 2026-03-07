package cli

import (
	"testing"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestParseCockpitMode(t *testing.T) {
	modes := map[string]calendarMode{
		"":         calendarModeCalendar,
		"calendar": calendarModeCalendar,
		"tree":     calendarModeTree,
		"board":    calendarModeBoard,
		"review":   calendarModeReview,
		"today":    calendarModeToday,
	}
	for input, expected := range modes {
		got, err := parseCockpitMode(input)
		if err != nil {
			t.Fatalf("parseCockpitMode(%q) failed: %v", input, err)
		}
		if got != expected {
			t.Fatalf("parseCockpitMode(%q) = %s, want %s", input, got, expected)
		}
	}
	if _, err := parseCockpitMode("unknown"); err == nil {
		t.Fatal("expected invalid mode error")
	}
}

func TestCockpitCommandRoutesToSelectedMode(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	cfg.Tags = append(cfg.Tags, "backend")
	if err := shelf.SaveConfig(root, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
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
		if len(opts.Filter.Statuses) != 1 || opts.Filter.Statuses[0] != "open" {
			t.Fatalf("unexpected filter statuses: %+v", opts.Filter.Statuses)
		}
		if len(opts.Filter.Tags) != 1 || opts.Filter.Tags[0] != "backend" {
			t.Fatalf("unexpected filter tags: %+v", opts.Filter.Tags)
		}
		return nil
	}
	cmd := newCockpitCommand(&commandContext{rootDir: root})
	cmd.SetArgs([]string{"--mode", "board", "--status", "open", "--tag", "backend"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("cockpit command failed: %v", err)
	}
	if !called {
		t.Fatal("expected cockpit TUI launcher to run")
	}
}
