package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
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
		{ID: "01A", Title: "A", DueOn: "2026-03-09", Status: "open"},
		{ID: "01B", Title: "B", DueOn: "2026-03-10"},
		{ID: "01C", Title: "C", DueOn: "2026-03-10", Status: "blocked"},
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
			if cell.DominantStatus != "blocked" {
				t.Fatalf("expected blocked dominant status, got %+v", cell)
			}
		}
	}
	if !found {
		t.Fatal("expected to find 2026-03-10 cell")
	}
}

func TestMoveCalendarIndexByMonth(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	days := buildCalendarDays(nil, start, 40)

	got := moveCalendarIndexByMonth(days, 0, 1)
	if got != 31 {
		t.Fatalf("unexpected next month index: %d", got)
	}

	got = moveCalendarIndexByMonth(days, 31, -1)
	if got != 0 {
		t.Fatalf("unexpected previous month index: %d", got)
	}
}

func TestDominantCalendarStatus(t *testing.T) {
	tasks := []shelf.Task{
		{Status: "open"},
		{Status: "done"},
		{Status: "blocked"},
		{Status: "in_progress"},
	}
	if got := dominantCalendarStatus(tasks); got != "blocked" {
		t.Fatalf("unexpected dominant status: %s", got)
	}
}

func TestResolveCalendarRangeDays(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	gotStart, gotDays, err := resolveCalendarRange(start, 14, 0, 0, shelf.DefaultConfig(), true, false, false)
	if err != nil {
		t.Fatalf("resolveCalendarRange failed: %v", err)
	}
	if gotStart.Format("2006-01-02") != "2026-03-09" || gotDays != 14 {
		t.Fatalf("unexpected range: %s %d", gotStart.Format("2006-01-02"), gotDays)
	}
}

func TestResolveCalendarRangeMonths(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	gotStart, gotDays, err := resolveCalendarRange(start, 0, 2, 0, shelf.DefaultConfig(), false, true, false)
	if err != nil {
		t.Fatalf("resolveCalendarRange failed: %v", err)
	}
	if gotStart.Format("2006-01-02") != "2026-03-01" {
		t.Fatalf("unexpected month start: %s", gotStart.Format("2006-01-02"))
	}
	if gotDays != 61 {
		t.Fatalf("unexpected day count: %d", gotDays)
	}
}

func TestResolveCalendarRangeYears(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	gotStart, gotDays, err := resolveCalendarRange(start, 0, 0, 2, shelf.DefaultConfig(), false, false, true)
	if err != nil {
		t.Fatalf("resolveCalendarRange failed: %v", err)
	}
	if gotStart.Format("2006-01-02") != "2026-01-01" || gotDays != 730 {
		t.Fatalf("unexpected range: %s %d", gotStart.Format("2006-01-02"), gotDays)
	}
}

func TestResolveCalendarRangeRejectsMixedFlags(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	if _, _, err := resolveCalendarRange(start, 7, 1, 0, shelf.DefaultConfig(), true, true, false); err == nil {
		t.Fatal("expected mixed flag error")
	}
}

func TestResolveCalendarRangeUsesConfigDefaultDays(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	cfg := shelf.DefaultConfig()
	cfg.CalendarDefaultUse = "days"
	cfg.CalendarDefaultDays = 21
	gotStart, gotDays, err := resolveCalendarRange(start, 0, 0, 0, cfg, false, false, false)
	if err != nil {
		t.Fatalf("resolveCalendarRange failed: %v", err)
	}
	if gotStart.Format("2006-01-02") != "2026-03-09" || gotDays != 21 {
		t.Fatalf("unexpected range: %s %d", gotStart.Format("2006-01-02"), gotDays)
	}
}

func TestResolveCalendarRangeUsesConfigDefaultMonths(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	cfg := shelf.DefaultConfig()
	cfg.CalendarDefaultUse = "months"
	cfg.CalendarDefaultMonths = 2
	gotStart, gotDays, err := resolveCalendarRange(start, 0, 0, 0, cfg, false, false, false)
	if err != nil {
		t.Fatalf("resolveCalendarRange failed: %v", err)
	}
	if gotStart.Format("2006-01-02") != "2026-03-01" || gotDays != 61 {
		t.Fatalf("unexpected range: %s %d", gotStart.Format("2006-01-02"), gotDays)
	}
}

func TestRenderCalendarCellKeepsFixedWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	cell := calendarMonthCell{
		Date:           time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local),
		InCurrentMonth: true,
		InRange:        true,
		TaskCount:      2,
		DominantStatus: "blocked",
	}
	rendered := renderCalendarCell(cell, "2026-03-09", 14)
	if got := lipgloss.Width(rendered); got != 14 {
		t.Fatalf("unexpected rendered width: %d", got)
	}
}

func TestCalendarApplyStatusChange(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Task",
		Kind:   "todo",
		Status: "open",
		DueOn:  "2026-03-09",
	})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked", "done", "cancelled"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}

	updatedModel, _ := model.applyStatusChange("done")
	calendarModel := updatedModel.(calendarTUIModel)
	updated, err := shelf.EnsureTaskExists(root, task.ID)
	if err != nil {
		t.Fatalf("EnsureTaskExists failed: %v", err)
	}
	if updated.Status != "done" {
		t.Fatalf("unexpected status: %s", updated.Status)
	}
	if calendarModel.message == "" {
		t.Fatal("expected status change message")
	}
}

func TestCalendarApplyStatusChangeKeepsTaskVisibleWhenFilteredOut(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Task",
		Kind:   "todo",
		Status: "open",
		DueOn:  "2026-03-09",
	})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}

	updatedModel, _ := model.applyStatusChange("done")
	calendarModel := updatedModel.(calendarTUIModel)
	updated, err := shelf.EnsureTaskExists(root, task.ID)
	if err != nil {
		t.Fatalf("EnsureTaskExists failed: %v", err)
	}
	if updated.Status != "done" {
		t.Fatalf("unexpected status: %s", updated.Status)
	}
	if len(calendarModel.days) == 0 || len(calendarModel.days[0].Tasks) != 1 {
		t.Fatalf("expected task to stay visible in current view: %+v", calendarModel.days)
	}
	if calendarModel.days[0].Tasks[0].Status != "done" {
		t.Fatalf("expected visible task status to update: %+v", calendarModel.days[0].Tasks[0])
	}
	if !strings.Contains(calendarModel.message, "visible until reload") {
		t.Fatalf("unexpected message: %s", calendarModel.message)
	}
}

func TestCalendarCreateTaskOnFocusedDay(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}

	if err := model.createTaskOnFocusedDay("Created from calendar"); err != nil {
		t.Fatalf("createTaskOnFocusedDay failed: %v", err)
	}

	if len(model.days) == 0 || len(model.days[0].Tasks) != 1 {
		t.Fatalf("expected created task in focused day: %+v", model.days)
	}
	created := model.days[0].Tasks[0]
	if created.Title != "Created from calendar" || created.DueOn != "2026-03-09" {
		t.Fatalf("unexpected created task: %+v", created)
	}
	if model.taskIndex != 0 {
		t.Fatalf("expected selected created task, got taskIndex=%d", model.taskIndex)
	}
}

func TestCalendarCreateTaskOnFocusedDayKeepsFilteredTaskVisible(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	cfg.DefaultStatus = "done"
	if err := shelf.SaveConfig(root, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}

	if err := model.createTaskOnFocusedDay("Filtered task"); err != nil {
		t.Fatalf("createTaskOnFocusedDay failed: %v", err)
	}

	if len(model.days[0].Tasks) != 1 || model.days[0].Tasks[0].Status != "done" {
		t.Fatalf("expected visible filtered task: %+v", model.days[0].Tasks)
	}
	if !strings.Contains(model.message, "visible until reload") {
		t.Fatalf("unexpected message: %s", model.message)
	}
}
