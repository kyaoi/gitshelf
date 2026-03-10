package cli

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestStartOfWeek(t *testing.T) {
	value := time.Date(2026, 3, 11, 10, 0, 0, 0, time.Local)
	got := startOfWeek(value)
	if got.Format("2006-01-02") != "2026-03-08" {
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
	if first.Date.Format("2006-01-02") != "2026-03-01" {
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

func TestRenderCalendarMonthShowsFocusedDateAndUsesFullWidth(t *testing.T) {
	month := calendarMonthView{
		Label: "March 2026",
		Weeks: [][]calendarMonthCell{{
			{Date: time.Date(2026, 3, 2, 0, 0, 0, 0, time.Local), InCurrentMonth: true, InRange: true, TaskCount: 1, DominantStatus: "open"},
			{Date: time.Date(2026, 3, 3, 0, 0, 0, 0, time.Local), InCurrentMonth: true, InRange: true},
			{Date: time.Date(2026, 3, 4, 0, 0, 0, 0, time.Local), InCurrentMonth: true, InRange: true},
			{Date: time.Date(2026, 3, 5, 0, 0, 0, 0, time.Local), InCurrentMonth: true, InRange: true},
			{Date: time.Date(2026, 3, 6, 0, 0, 0, 0, time.Local), InCurrentMonth: true, InRange: true},
			{Date: time.Date(2026, 3, 7, 0, 0, 0, 0, time.Local), InCurrentMonth: true, InRange: true},
			{Date: time.Date(2026, 3, 8, 0, 0, 0, 0, time.Local), InCurrentMonth: true, InRange: true},
		}},
	}

	compact := renderCalendarMonth(month, "2026-03-07", 120, true, 2)
	full := renderCalendarMonth(month, "2026-03-07", 120, false, 4)

	if !strings.Contains(full, "March 2026 - 2026/03/07") {
		t.Fatalf("month title should include focused date: %q", full)
	}

	compactLines := strings.Split(compact, "\n")
	fullLines := strings.Split(full, "\n")
	if len(compactLines) < 2 || len(fullLines) < 2 {
		t.Fatalf("unexpected rendered month output")
	}
	if lipgloss.Width(fullLines[1]) <= lipgloss.Width(compactLines[1]) {
		t.Fatalf("full calendar should use more horizontal space than compact mode: compact=%d full=%d", lipgloss.Width(compactLines[1]), lipgloss.Width(fullLines[1]))
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

func TestPlanCalendarWindowWithinRange(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	target := time.Date(2026, 3, 12, 12, 0, 0, 0, time.Local)
	gotStart, gotIndex := planCalendarWindow(start, 7, target)
	if gotStart.Format("2006-01-02") != "2026-03-09" || gotIndex != 3 {
		t.Fatalf("unexpected window plan: %s %d", gotStart.Format("2006-01-02"), gotIndex)
	}
}

func TestPlanCalendarWindowMovesBackward(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	target := time.Date(2026, 3, 7, 0, 0, 0, 0, time.Local)
	gotStart, gotIndex := planCalendarWindow(start, 7, target)
	if gotStart.Format("2006-01-02") != "2026-03-07" || gotIndex != 0 {
		t.Fatalf("unexpected backward window plan: %s %d", gotStart.Format("2006-01-02"), gotIndex)
	}
}

func TestPlanCalendarWindowMovesForward(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	target := time.Date(2026, 3, 20, 0, 0, 0, 0, time.Local)
	gotStart, gotIndex := planCalendarWindow(start, 7, target)
	if gotStart.Format("2006-01-02") != "2026-03-14" || gotIndex != 6 {
		t.Fatalf("unexpected forward window plan: %s %d", gotStart.Format("2006-01-02"), gotIndex)
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
	gotStart, gotDays, err := resolveCalendarRange(start, 14, 0, 0, shelf.DefaultConfig().Commands.Calendar, true, false, false)
	if err != nil {
		t.Fatalf("resolveCalendarRange failed: %v", err)
	}
	if gotStart.Format("2006-01-02") != "2026-03-09" || gotDays != 14 {
		t.Fatalf("unexpected range: %s %d", gotStart.Format("2006-01-02"), gotDays)
	}
}

func TestResolveCalendarRangeMonths(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	gotStart, gotDays, err := resolveCalendarRange(start, 0, 2, 0, shelf.DefaultConfig().Commands.Calendar, false, true, false)
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
	gotStart, gotDays, err := resolveCalendarRange(start, 0, 0, 2, shelf.DefaultConfig().Commands.Calendar, false, false, true)
	if err != nil {
		t.Fatalf("resolveCalendarRange failed: %v", err)
	}
	if gotStart.Format("2006-01-02") != "2026-01-01" || gotDays != 730 {
		t.Fatalf("unexpected range: %s %d", gotStart.Format("2006-01-02"), gotDays)
	}
}

func TestResolveCalendarRangeRejectsMixedFlags(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	if _, _, err := resolveCalendarRange(start, 7, 1, 0, shelf.DefaultConfig().Commands.Calendar, true, true, false); err == nil {
		t.Fatal("expected mixed flag error")
	}
}

func TestResolveCalendarRangeUsesConfigDefaultDays(t *testing.T) {
	start := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	cfg := shelf.DefaultConfig().Commands.Calendar
	cfg.DefaultRangeUnit = "days"
	cfg.DefaultDays = 21
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
	cfg := shelf.DefaultConfig().Commands.Calendar
	cfg.DefaultRangeUnit = "months"
	cfg.DefaultMonths = 2
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
	rendered := renderCalendarCell(cell, "2026-03-09", 14, true, 2)
	if got := lipgloss.Width(rendered); got != 14 {
		t.Fatalf("unexpected rendered width: %d", got)
	}
	if strings.Contains(rendered, "2026-03-09") {
		t.Fatalf("calendar cell should not render full date anymore: %q", rendered)
	}
}

func TestCalendarLayoutUsesNarrowerTallerMainGrid(t *testing.T) {
	model := calendarTUIModel{mode: calendarModeCalendar, width: 140}
	mainWidth, gapWidth, inspectorWidth := model.layoutColumns()
	if mainWidth != 89 || gapWidth != 1 || inspectorWidth != 48 {
		t.Fatalf("calendar layout should use 65:1:34 ratio at width=140, got main=%d gap=%d inspector=%d", mainWidth, gapWidth, inspectorWidth)
	}
	rendered := renderCalendarCell(calendarMonthCell{
		Date:           time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local),
		InCurrentMonth: true,
		InRange:        true,
		TaskCount:      2,
		DominantStatus: "open",
	}, "2026-03-09", 12, false, 4)
	if got := lipgloss.Height(rendered); got != 4 {
		t.Fatalf("calendar mode cells should be taller, got height=%d", got)
	}
}

func TestCalendarMainCellHeightScalesWithViewport(t *testing.T) {
	if got := calendarMainCellHeight(18); got != 4 {
		t.Fatalf("expected minimum cell height 4, got %d", got)
	}
	if got := calendarMainCellHeight(46); got <= 4 {
		t.Fatalf("expected larger viewport to increase cell height, got %d", got)
	}
}

func TestNowTriptychKeepsColumnOrderWithEmptySection(t *testing.T) {
	rendered := renderCalendarTriptychSections([]calendarSection{
		{ID: calendarSectionFocusedDay, Title: "Selected Day", Items: []calendarSectionItem{{Task: shelf.Task{ID: "01A", Title: "Focus"}}}},
		{ID: calendarSectionOverdue, Title: "Overdue"},
		{ID: calendarSectionToday, Title: "Today", Items: []calendarSectionItem{{Task: shelf.Task{ID: "01B", Title: "Today"}}}},
	}, 0, map[calendarSectionID]int{}, false, 96, 12)
	if !strings.Contains(rendered, "Selected Day 1") || !strings.Contains(rendered, "Overdue 0") || !strings.Contains(rendered, "Today 1") {
		t.Fatalf("triptych should keep all column headers visible: %q", rendered)
	}
	if !strings.Contains(rendered, "│") {
		t.Fatalf("triptych should render fixed separators between columns: %q", rendered)
	}
}

func TestBoardPaneKeepsEmptyColumnsFixed(t *testing.T) {
	rendered := renderCockpitBoardPane([]boardColumn{
		{Status: "open", Tasks: []shelf.Task{{ID: "01A", Title: "Open", Kind: "todo", Status: "open"}}},
		{Status: "blocked", Tasks: nil},
		{Status: "done", Tasks: []shelf.Task{{ID: "01B", Title: "Done", Kind: "todo", Status: "done"}}},
	}, 0, map[int]int{0: 0, 1: 0, 2: 0}, map[string]struct{}{}, false, 96, 12)
	for _, want := range []string{"open 1", "blocked 0", "done 1", "(none)"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("board pane should keep fixed columns, missing %q in %q", want, rendered)
		}
	}
	if !strings.Contains(rendered, "│") {
		t.Fatalf("board pane should render fixed separators between columns: %q", rendered)
	}
}

func TestRenderCalendarViewportKeepsHeaderVisible(t *testing.T) {
	rendered := renderCalendarViewport(
		[]string{"Header", "Tabs"},
		strings.Join([]string{"line1", "line2", "line3", "line4", "line5", "line6"}, "\n"),
		nil,
		6,
		0,
	)
	if !strings.Contains(rendered, "Header") || !strings.Contains(rendered, "Tabs") {
		t.Fatalf("viewport should keep header blocks visible: %q", rendered)
	}
	if strings.Contains(rendered, "line6") {
		t.Fatalf("viewport should clip overflowing body lines: %q", rendered)
	}
}

func TestRenderCalendarViewportScrollsBody(t *testing.T) {
	body := strings.Join([]string{"line1", "line2", "line3", "line4", "line5", "line6"}, "\n")
	rendered := renderCalendarViewport([]string{"Header", "Tabs"}, body, nil, 6, 3)
	if !strings.Contains(rendered, "line6") {
		t.Fatalf("viewport should reveal lower lines after scrolling: %q", rendered)
	}
	if strings.Contains(rendered, "line1") {
		t.Fatalf("viewport should hide top body lines after scrolling: %q", rendered)
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
	if model.selectedTaskID != created.ID {
		t.Fatalf("expected selected created task, got selectedTaskID=%s", model.selectedTaskID)
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

func TestCalendarKindPickerUpdatesAddKind(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}

	model.beginAddMode(true)
	model.beginKindMode(calendarKindTargetAdd)
	for i, kind := range model.kindChoices {
		if kind == "idea" {
			model.kindIndex = i
			break
		}
	}
	updatedModel, _ := model.updateKindMode(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(calendarTUIModel)
	if model.addKind != "idea" {
		t.Fatalf("expected add kind idea, got %s", model.addKind)
	}
	if err := model.createTaskFromAddMode("Idea from add"); err != nil {
		t.Fatalf("createTaskFromAddMode failed: %v", err)
	}
	created := model.taskByID[model.selectedTaskID]
	if created.Kind != "idea" {
		t.Fatalf("expected created kind idea, got %+v", created)
	}
}

func TestCalendarApplySelectedTaskKind(t *testing.T) {
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

	if err := model.applySelectedTaskKind("idea"); err != nil {
		t.Fatalf("applySelectedTaskKind failed: %v", err)
	}
	updated, err := shelf.EnsureTaskExists(root, task.ID)
	if err != nil {
		t.Fatalf("EnsureTaskExists failed: %v", err)
	}
	if updated.Kind != "idea" {
		t.Fatalf("expected updated kind idea, got %+v", updated)
	}
}

func TestCalendarApplySelectedTaskKindUsesMarkedTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "First",
		Kind:   "todo",
		Status: "open",
		DueOn:  "2026-03-09",
	})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Second",
		Kind:   "todo",
		Status: "open",
		DueOn:  "2026-03-09",
	})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.markedTaskIDs = map[string]struct{}{first.ID: {}, second.ID: {}}
	model.selectTaskByID(second.ID)

	if err := model.applySelectedTaskKind("idea"); err != nil {
		t.Fatalf("applySelectedTaskKind failed: %v", err)
	}

	for _, taskID := range []string{first.ID, second.ID} {
		updated, err := shelf.EnsureTaskExists(root, taskID)
		if err != nil {
			t.Fatalf("EnsureTaskExists failed for %s: %v", taskID, err)
		}
		if updated.Kind != "idea" {
			t.Fatalf("expected %s kind idea, got %+v", taskID, updated)
		}
	}
	if model.markedCount() != 0 {
		t.Fatalf("expected marks cleared after bulk kind update, got %d", model.markedCount())
	}
	if model.selectedTaskID != first.ID {
		t.Fatalf("expected first updated task selected, got %s", model.selectedTaskID)
	}
	if model.message != "Updated kind to idea for 2 tasks" {
		t.Fatalf("unexpected message: %s", model.message)
	}
}

func TestCalendarApplySelectedTaskTags(t *testing.T) {
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

	if err := model.applySelectedTaskTags([]string{"backend", "urgent"}); err != nil {
		t.Fatalf("applySelectedTaskTags failed: %v", err)
	}
	updated, err := shelf.EnsureTaskExists(root, task.ID)
	if err != nil {
		t.Fatalf("EnsureTaskExists failed: %v", err)
	}
	if !strings.Contains(strings.Join(updated.Tags, ","), "backend") || !strings.Contains(strings.Join(updated.Tags, ","), "urgent") {
		t.Fatalf("expected updated tags, got %+v", updated.Tags)
	}
	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if !containsTag(cfg.Tags, "backend") || !containsTag(cfg.Tags, "urgent") {
		t.Fatalf("expected config tags updated, got %+v", cfg.Tags)
	}
}

func TestCalendarApplySnoozeOptionUsesMarkedTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "First",
		Kind:   "todo",
		Status: "open",
		DueOn:  "2026-03-09",
	})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Second",
		Kind:   "todo",
		Status: "open",
		DueOn:  "2026-03-10",
	})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.markedTaskIDs = map[string]struct{}{first.ID: {}, second.ID: {}}
	model.selectTaskByID(second.ID)

	if err := model.applySnoozeOption(snoozePreset{Label: "Today", Mode: snoozeModeTo, Value: "2026-03-20"}); err != nil {
		t.Fatalf("applySnoozeOption failed: %v", err)
	}

	for _, taskID := range []string{first.ID, second.ID} {
		updated, err := shelf.EnsureTaskExists(root, taskID)
		if err != nil {
			t.Fatalf("EnsureTaskExists failed for %s: %v", taskID, err)
		}
		if updated.DueOn != "2026-03-20" {
			t.Fatalf("expected %s due_on 2026-03-20, got %+v", taskID, updated)
		}
	}
	if model.markedCount() != 0 {
		t.Fatalf("expected marks cleared after bulk snooze, got %d", model.markedCount())
	}
	if model.selectedTaskID != first.ID {
		t.Fatalf("expected first updated task selected, got %s", model.selectedTaskID)
	}
	if model.message != "Snoozed 2 tasks to 2026-03-20" {
		t.Fatalf("unexpected message: %s", model.message)
	}
}

func TestBeginDuePromptUsesSharedMarkedValue(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(first.ID)
	model.markedTaskIDs = map[string]struct{}{first.ID: {}, second.ID: {}}

	model.beginDuePrompt()

	if !model.textPromptMode || model.textPromptPurpose != calendarTextPromptDueOn {
		t.Fatalf("expected due prompt mode, got mode=%v purpose=%v", model.textPromptMode, model.textPromptPurpose)
	}
	if model.textPromptValue != "2026-03-09" {
		t.Fatalf("expected shared due prefilled, got %q", model.textPromptValue)
	}
}

func TestBeginRepeatPromptLeavesMixedMarkedValueBlank(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open", RepeatEvery: "1w"})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open", RepeatEvery: "2w"})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(first.ID)
	model.markedTaskIDs = map[string]struct{}{first.ID: {}, second.ID: {}}

	model.beginRepeatPrompt()

	if !model.textPromptMode || model.textPromptPurpose != calendarTextPromptRepeatEvery {
		t.Fatalf("expected repeat prompt mode, got mode=%v purpose=%v", model.textPromptMode, model.textPromptPurpose)
	}
	if model.textPromptValue != "" {
		t.Fatalf("expected blank repeat prompt for mixed values, got %q", model.textPromptValue)
	}
}

func TestApplySelectedTaskDueOnUsesMarkedTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open", DueOn: "2026-03-10"})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(second.ID)
	model.markedTaskIDs = map[string]struct{}{first.ID: {}, second.ID: {}}

	if err := model.applySelectedTaskDueOn("2026-03-20"); err != nil {
		t.Fatalf("applySelectedTaskDueOn failed: %v", err)
	}

	for _, taskID := range []string{first.ID, second.ID} {
		updated, err := shelf.EnsureTaskExists(root, taskID)
		if err != nil {
			t.Fatalf("EnsureTaskExists failed for %s: %v", taskID, err)
		}
		if updated.DueOn != "2026-03-20" {
			t.Fatalf("expected %s due_on updated, got %+v", taskID, updated)
		}
	}
	if model.markedCount() != 0 {
		t.Fatalf("expected marks cleared after due update, got %d", model.markedCount())
	}
	if model.selectedTaskID != first.ID {
		t.Fatalf("expected first updated task selected, got %s", model.selectedTaskID)
	}
	if model.message != "Updated due to 2026-03-20 for 2 tasks" {
		t.Fatalf("unexpected message: %s", model.message)
	}
}

func TestApplySelectedTaskRepeatEveryClearsMarkedTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open", RepeatEvery: "1w"})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open", RepeatEvery: "2w"})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(second.ID)
	model.markedTaskIDs = map[string]struct{}{first.ID: {}, second.ID: {}}

	if err := model.applySelectedTaskRepeatEvery(""); err != nil {
		t.Fatalf("applySelectedTaskRepeatEvery failed: %v", err)
	}

	for _, taskID := range []string{first.ID, second.ID} {
		updated, err := shelf.EnsureTaskExists(root, taskID)
		if err != nil {
			t.Fatalf("EnsureTaskExists failed for %s: %v", taskID, err)
		}
		if updated.RepeatEvery != "" {
			t.Fatalf("expected %s repeat_every cleared, got %+v", taskID, updated)
		}
	}
	if model.markedCount() != 0 {
		t.Fatalf("expected marks cleared after repeat update, got %d", model.markedCount())
	}
	if model.selectedTaskID != first.ID {
		t.Fatalf("expected first updated task selected, got %s", model.selectedTaskID)
	}
	if model.message != "Cleared repeat for 2 tasks" {
		t.Fatalf("unexpected message: %s", model.message)
	}
}

func TestBeginAppendBodyPromptTargetsMarkedTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(first.ID)
	model.markedTaskIDs = map[string]struct{}{first.ID: {}, second.ID: {}}

	model.beginAppendBodyPrompt()

	if !model.textPromptMode || model.textPromptPurpose != calendarTextPromptAppendBody {
		t.Fatalf("expected append body prompt mode, got mode=%v purpose=%v", model.textPromptMode, model.textPromptPurpose)
	}
	if !strings.Contains(model.textPromptHelp, "2 marked tasks") {
		t.Fatalf("expected marked target in help, got %q", model.textPromptHelp)
	}
}

func TestUpdateTextPromptModeSupportsCtrlJForAppendBody(t *testing.T) {
	model := calendarTUIModel{
		textPromptMode:    true,
		textPromptPurpose: calendarTextPromptAppendBody,
		textPromptValue:   "ab",
		textPromptCursor:  1,
	}

	updated, _ := model.updateTextPromptMode(tea.KeyMsg{Type: tea.KeyCtrlJ})
	model = updated.(calendarTUIModel)
	if model.textPromptValue != "a\nb" || model.textPromptCursor != 2 {
		t.Fatalf("unexpected append-body edit: value=%q cursor=%d", model.textPromptValue, model.textPromptCursor)
	}
}

func TestApplySelectedTaskAppendBodyUsesMarkedTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "First",
		Kind:   "todo",
		Status: "open",
		Body:   "first body",
	})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Second",
		Kind:   "todo",
		Status: "open",
	})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(second.ID)
	model.markedTaskIDs = map[string]struct{}{first.ID: {}, second.ID: {}}

	if err := model.applySelectedTaskAppendBody("new line 1\nnew line 2"); err != nil {
		t.Fatalf("applySelectedTaskAppendBody failed: %v", err)
	}

	reloadedFirst, err := shelf.EnsureTaskExists(root, first.ID)
	if err != nil {
		t.Fatalf("EnsureTaskExists failed for first: %v", err)
	}
	if reloadedFirst.Body != "first body\nnew line 1\nnew line 2" {
		t.Fatalf("unexpected first body: %q", reloadedFirst.Body)
	}
	reloadedSecond, err := shelf.EnsureTaskExists(root, second.ID)
	if err != nil {
		t.Fatalf("EnsureTaskExists failed for second: %v", err)
	}
	if reloadedSecond.Body != "new line 1\nnew line 2" {
		t.Fatalf("unexpected second body: %q", reloadedSecond.Body)
	}
	if model.markedCount() != 0 {
		t.Fatalf("expected marks cleared after append body, got %d", model.markedCount())
	}
	if model.selectedTaskID != first.ID {
		t.Fatalf("expected first updated task selected, got %s", model.selectedTaskID)
	}
	if model.message != "Appended note to 2 tasks" {
		t.Fatalf("unexpected message: %s", model.message)
	}
}

func TestCalendarBulkActionPopupLabelUsesMarkedCount(t *testing.T) {
	model := calendarTUIModel{
		showID: true,
		mode:   calendarModeReview,
		sections: []calendarSection{{
			ID: "today",
			Items: []calendarSectionItem{{
				Task: shelf.Task{ID: "01ABCDEFG", Title: "Selected"},
			}},
		}},
		sectionRows: map[calendarSectionID]int{"today": 0},
		markedTaskIDs: map[string]struct{}{
			"01ABCDEFG": {},
			"01HIJKLMN": {},
		},
	}
	if got := model.bulkActionPopupLabel(); got != "2 marked tasks" {
		t.Fatalf("expected marked task label, got %q", got)
	}

	model.markedTaskIDs = nil
	if got := model.bulkActionPopupLabel(); got != "[01ABCDEF] Selected" {
		t.Fatalf("expected selected task label, got %q", got)
	}
}

func TestCalendarShowsDescendantsOfDueParent(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	child, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Child", Kind: "todo", Status: "open", Parent: parent.ID})
	if err != nil {
		t.Fatalf("add child failed: %v", err)
	}
	grandchild, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Grandchild", Kind: "todo", Status: "open", Parent: child.ID})
	if err != nil {
		t.Fatalf("add grandchild failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	if len(model.days) == 0 || len(model.days[0].Tasks) != 3 {
		t.Fatalf("expected parent and descendants on focused day, got %+v", model.days)
	}
	byID := map[string]shelf.Task{}
	for _, task := range model.days[0].Tasks {
		byID[task.ID] = task
	}
	for _, id := range []string{parent.ID, child.ID, grandchild.ID} {
		if byID[id].DueOn != "2026-03-09" {
			t.Fatalf("expected effective due for %s, got %+v", id, byID[id])
		}
	}
}

func TestReviewModeIncludesDescendantsOfDueParentInToday(t *testing.T) {
	root := t.TempDir()
	today := time.Now().Local().Format("2006-01-02")
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent", Kind: "todo", Status: "open", DueOn: today})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	child, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Child", Kind: "todo", Status: "open", Parent: parent.ID})
	if err != nil {
		t.Fatalf("add child failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Grandchild", Kind: "todo", Status: "open", Parent: child.ID}); err != nil {
		t.Fatalf("add grandchild failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 7, []shelf.Status{"open", "in_progress", "blocked"}, calendarTUIOptions{
		Mode:   calendarModeReview,
		ShowID: false,
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	found := 0
	for _, section := range model.sections {
		if section.ID != calendarSectionToday {
			continue
		}
		found = len(section.Items)
	}
	if found != 3 {
		t.Fatalf("expected today section to include descendants, got %d", found)
	}
}

func TestSelectedTitleCopyTextUsesSelectedTask(t *testing.T) {
	model := calendarTUIModel{
		copySeparator: "\n",
		taskByID: map[string]shelf.Task{
			"01A": {ID: "01A", Title: "First"},
		},
		sections: []calendarSection{{
			ID:    calendarSectionFocusedDay,
			Title: "Focused",
			Items: []calendarSectionItem{{Task: shelf.Task{ID: "01A", Title: "First"}}},
		}},
		sectionRows: map[calendarSectionID]int{calendarSectionFocusedDay: 0},
	}
	text, count, err := model.selectedTitleCopyText()
	if err != nil {
		t.Fatalf("selectedTitleCopyText failed: %v", err)
	}
	if count != 1 || text != "First" {
		t.Fatalf("unexpected copy payload: count=%d text=%q", count, text)
	}
}

func TestSelectedTitleCopyTextUsesMarkedOrderAndSeparator(t *testing.T) {
	model := calendarTUIModel{
		mode:          calendarModeTree,
		copySeparator: ", ",
		taskByID: map[string]shelf.Task{
			"01A": {ID: "01A", Title: "First"},
			"01B": {ID: "01B", Title: "Second"},
		},
		treeRows: []cockpitTreeRow{
			{Task: shelf.Task{ID: "01A", Title: "First"}},
			{Task: shelf.Task{ID: "01B", Title: "Second"}},
		},
		markedTaskIDs: map[string]struct{}{
			"01A": {},
			"01B": {},
		},
	}
	text, count, err := model.selectedTitleCopyText()
	if err != nil {
		t.Fatalf("selectedTitleCopyText failed: %v", err)
	}
	if count != 2 || text != "First, Second" {
		t.Fatalf("unexpected copy payload: count=%d text=%q", count, text)
	}
}

func TestSelectedPathCopyTextUsesMarkedAbsolutePaths(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	model := calendarTUIModel{
		rootDir:       root,
		copySeparator: "\n",
		taskByID: map[string]shelf.Task{
			"01A": {ID: "01A", Title: "First"},
		},
		markedTaskIDs: map[string]struct{}{"01A": {}},
	}
	text, count, err := model.selectedPathCopyText()
	if err != nil {
		t.Fatalf("selectedPathCopyText failed: %v", err)
	}
	want := filepath.Join(shelf.TasksDir(root), "01A.md")
	if count != 1 || text != want {
		t.Fatalf("unexpected path payload: count=%d text=%q want=%q", count, text, want)
	}
}

func TestSelectedBodyCopyTextUsesMarkedOrderAndSeparator(t *testing.T) {
	model := calendarTUIModel{
		mode:          calendarModeTree,
		copySeparator: "\n---\n",
		taskByID: map[string]shelf.Task{
			"01A": {ID: "01A", Title: "First", Body: "alpha"},
			"01B": {ID: "01B", Title: "Second", Body: "beta"},
		},
		treeRows: []cockpitTreeRow{
			{Task: shelf.Task{ID: "01A", Title: "First"}},
			{Task: shelf.Task{ID: "01B", Title: "Second"}},
		},
		markedTaskIDs: map[string]struct{}{
			"01A": {},
			"01B": {},
		},
	}
	text, count, err := model.selectedBodyCopyText()
	if err != nil {
		t.Fatalf("selectedBodyCopyText failed: %v", err)
	}
	if count != 2 || text != "alpha\n---\nbeta" {
		t.Fatalf("unexpected body payload: count=%d text=%q", count, text)
	}
}

func TestSelectedSubtreeCopyTextUsesIndentedTreeAndDedupesMarkedDescendants(t *testing.T) {
	parent := shelf.Task{ID: "01A", Title: "Parent"}
	child := shelf.Task{ID: "01B", Title: "Child", Parent: "01A"}
	grandchild := shelf.Task{ID: "01C", Title: "Grandchild", Parent: "01B"}
	model := calendarTUIModel{
		taskByID: map[string]shelf.Task{
			parent.ID:     parent,
			child.ID:      child,
			grandchild.ID: grandchild,
		},
		allTasks: []shelf.Task{parent, child, grandchild},
		markedTaskIDs: map[string]struct{}{
			parent.ID: {},
			child.ID:  {},
		},
	}
	text, count, err := model.selectedSubtreeCopyText()
	if err != nil {
		t.Fatalf("selectedSubtreeCopyText failed: %v", err)
	}
	want := "Parent\n  Child\n    Grandchild"
	if count != 3 || text != want {
		t.Fatalf("unexpected subtree payload: count=%d text=%q want=%q", count, text, want)
	}
}

func TestCalendarLinkModeAddsLink(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	from, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "From", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add from failed: %v", err)
	}
	to, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "To", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add to failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.selectTaskByID(from.ID)
	model.beginLinkMode(calendarLinkActionAdd)
	candidates := model.currentLinkCandidates()
	for i, candidate := range candidates {
		if candidate.TaskID == to.ID {
			model.linkIndex = i
			break
		}
	}
	updatedModel, _ := model.updateLinkMode(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(calendarTUIModel)
	outbound, _, err := shelf.ListLinks(root, from.ID)
	if err != nil {
		t.Fatalf("ListLinks failed: %v", err)
	}
	if len(outbound) != 1 || outbound[0].To != to.ID {
		t.Fatalf("expected outbound link to %s, got %+v", to.ID, outbound)
	}
}

func TestCalendarLinkModeRemovesOutboundLink(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	from, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "From", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add from failed: %v", err)
	}
	to, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "To", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add to failed: %v", err)
	}
	if err := shelf.LinkTasks(root, from.ID, to.ID, "depends_on"); err != nil {
		t.Fatalf("LinkTasks failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.selectTaskByID(from.ID)
	model.beginLinkMode(calendarLinkActionRemove)
	updatedModel, _ := model.updateLinkMode(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(calendarTUIModel)
	outbound, _, err := shelf.ListLinks(root, from.ID)
	if err != nil {
		t.Fatalf("ListLinks failed: %v", err)
	}
	if len(outbound) != 0 {
		t.Fatalf("expected outbound links removed, got %+v", outbound)
	}
}

func TestCalendarLinkQueryModeTreatsKeysAsInput(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	from, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "From", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add from failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Target", Kind: "todo", Status: "open"}); err != nil {
		t.Fatalf("add target failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.selectTaskByID(from.ID)
	model.beginLinkMode(calendarLinkActionAdd)

	updated, _ := model.updateLinkMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(calendarTUIModel)
	if !model.linkQueryMode {
		t.Fatal("expected / to enter link query mode")
	}

	updated, _ = model.updateLinkMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = updated.(calendarTUIModel)
	if model.linkQuery != "q" {
		t.Fatalf("expected q to be appended to query, got %q", model.linkQuery)
	}

	updated, _ = model.updateLinkMode(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(calendarTUIModel)
	if model.linkQueryMode {
		t.Fatal("expected esc to exit query mode only")
	}
	if model.linkQuery != "q" {
		t.Fatalf("expected esc to keep current query text, got %q", model.linkQuery)
	}
}

func TestCalendarLinkQueryModeSupportsMidStringEditing(t *testing.T) {
	model := calendarTUIModel{
		linkMode:        true,
		linkQueryMode:   true,
		linkQuery:       "ab",
		linkQueryCursor: 2,
	}

	updated, _ := model.updateLinkMode(tea.KeyMsg{Type: tea.KeyLeft})
	model = updated.(calendarTUIModel)
	if model.linkQueryCursor != 1 {
		t.Fatalf("expected cursor to move left, got %d", model.linkQueryCursor)
	}

	updated, _ = model.updateLinkMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
	model = updated.(calendarTUIModel)
	if model.linkQuery != "aXb" || model.linkQueryCursor != 2 {
		t.Fatalf("unexpected query edit: value=%q cursor=%d", model.linkQuery, model.linkQueryCursor)
	}
}

func TestCalendarLinkModeUsesTabForTypeAndSupportsCollapse(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	source, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Source", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add source failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Child", Kind: "todo", Status: "open", Parent: parent.ID}); err != nil {
		t.Fatalf("add child failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(source.ID)
	model.beginLinkMode(calendarLinkActionAdd)

	initialType := model.linkTypeIndex
	updated, _ := model.updateLinkMode(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(calendarTUIModel)
	if len(model.linkTypes()) > 1 && model.linkTypeIndex == initialType {
		t.Fatal("expected tab to advance link type")
	}
	updated, _ = model.updateLinkMode(tea.KeyMsg{Type: tea.KeyShiftTab})
	model = updated.(calendarTUIModel)
	if model.linkTypeIndex != initialType {
		t.Fatalf("expected shift+tab to cycle back, got %d want %d", model.linkTypeIndex, initialType)
	}

	candidates := model.currentLinkCandidates()
	parentIndex := -1
	for i, candidate := range candidates {
		if candidate.TaskID == parent.ID {
			parentIndex = i
			break
		}
	}
	if parentIndex < 0 {
		t.Fatal("expected parent candidate in link picker")
	}
	model.linkIndex = parentIndex
	before := len(model.currentLinkCandidates())
	updated, _ = model.updateLinkMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	model = updated.(calendarTUIModel)
	after := len(model.currentLinkCandidates())
	if after >= before {
		t.Fatalf("expected h to collapse candidate subtree, before=%d after=%d", before, after)
	}
	updated, _ = model.updateLinkMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	model = updated.(calendarTUIModel)
	if len(model.currentLinkCandidates()) != before {
		t.Fatalf("expected l to expand candidate subtree, got %d want %d", len(model.currentLinkCandidates()), before)
	}
}

func TestBuildCalendarSectionsReviewMode(t *testing.T) {
	today := time.Now().Local().Format("2006-01-02")
	overdue := time.Now().Local().AddDate(0, 0, -1).Format("2006-01-02")
	rootTasks := []shelf.Task{
		{ID: "01INBOX", Title: "Inbox", Kind: "inbox", Status: "open"},
		{ID: "01OVER", Title: "Overdue", Kind: "todo", Status: "open", DueOn: overdue},
		{ID: "01TODAY", Title: "Today", Kind: "todo", Status: "in_progress", DueOn: today},
		{ID: "01BLOCK", Title: "Blocked", Kind: "todo", Status: "blocked"},
		{ID: "01READY", Title: "Ready", Kind: "todo", Status: "open"},
	}
	focused := &calendarDay{Date: today, Tasks: []shelf.Task{rootTasks[2]}}
	readiness := map[string]shelf.TaskReadiness{
		"01INBOX": {Ready: false},
		"01OVER":  {Ready: true},
		"01TODAY": {Ready: true},
		"01BLOCK": {Ready: false, BlockedByDeps: true, UnresolvedDependsOn: []string{"01OVER"}},
		"01READY": {Ready: true},
	}
	titles := map[string]string{"01OVER": "Overdue"}

	sections := buildCalendarSections(calendarModeReview, focused, rootTasks, readiness, titles, "depends_on", 0)
	if len(sections) != 6 {
		t.Fatalf("unexpected section count: %d", len(sections))
	}
	if sections[0].ID != calendarSectionFocusedDay || len(sections[0].Items) != 1 {
		t.Fatalf("unexpected focused day section: %+v", sections[0])
	}
	if sections[1].ID != calendarSectionInbox || len(sections[1].Items) != 1 || sections[1].Items[0].Task.ID != "01INBOX" {
		t.Fatalf("unexpected inbox section: %+v", sections[1])
	}
	if sections[2].ID != calendarSectionOverdue || len(sections[2].Items) != 1 || sections[2].Items[0].Task.ID != "01OVER" {
		t.Fatalf("unexpected overdue section: %+v", sections[2])
	}
	if sections[3].ID != calendarSectionToday || len(sections[3].Items) != 1 || sections[3].Items[0].Task.ID != "01TODAY" {
		t.Fatalf("unexpected today section: %+v", sections[3])
	}
	if sections[4].ID != calendarSectionBlocked || len(sections[4].Items) != 1 || !strings.Contains(sections[4].Items[0].Reason, "depends_on") {
		t.Fatalf("unexpected blocked section: %+v", sections[4])
	}
	if sections[5].ID != calendarSectionReady || len(sections[5].Items) != 3 {
		t.Fatalf("unexpected ready section: %+v", sections[5])
	}
}

func TestBuildCalendarSectionsTodayModeHonorsLimit(t *testing.T) {
	today := time.Now().Local().Format("2006-01-02")
	overdue := time.Now().Local().AddDate(0, 0, -1).Format("2006-01-02")
	tasks := []shelf.Task{
		{ID: "01A", Title: "A", Kind: "todo", Status: "open", DueOn: overdue},
		{ID: "01B", Title: "B", Kind: "todo", Status: "open", DueOn: overdue},
		{ID: "01C", Title: "C", Kind: "todo", Status: "open", DueOn: today},
		{ID: "01D", Title: "D", Kind: "todo", Status: "open", DueOn: today},
	}
	sections := buildCalendarSections(calendarModeNow, &calendarDay{Date: today, Tasks: []shelf.Task{tasks[2], tasks[3]}}, tasks, map[string]shelf.TaskReadiness{}, map[string]string{}, "depends_on", 1)
	if len(sections) != 3 {
		t.Fatalf("unexpected section count: %d", len(sections))
	}
	if len(sections[0].Items) != 2 {
		t.Fatalf("focused day section should not be limited: %+v", sections[0])
	}
	if len(sections[1].Items) != 1 || len(sections[2].Items) != 1 {
		t.Fatalf("non-focused sections should be limited: %+v", sections)
	}
}

func TestCalendarRebuildSectionsPreservesSelectedTask(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.selectTaskByID(second.ID)
	if model.selectedTaskID != second.ID {
		t.Fatalf("expected second selected before rebuild, got=%s", model.selectedTaskID)
	}
	model.rebuildSections()
	if model.selectedTaskID != second.ID {
		t.Fatalf("expected second selected after rebuild, got=%s", model.selectedTaskID)
	}
	selected, ok := model.selectedTask()
	if !ok || selected.ID != second.ID {
		t.Fatalf("unexpected selected task after rebuild: %+v ok=%t", selected, ok)
	}
	if first.ID == second.ID {
		t.Fatal("expected distinct test tasks")
	}
}

func TestFlattenCockpitTreeRows(t *testing.T) {
	nodes := []shelf.TreeNode{
		{
			Task: shelf.Task{ID: "01A", Title: "Parent", Kind: "todo", Status: "open"},
			Children: []shelf.TreeNode{
				{Task: shelf.Task{ID: "01B", Title: "Child", Kind: "memo", Status: "blocked"}},
			},
		},
	}
	rows := flattenCockpitTreeRows(nodes, "", true, false, map[string]struct{}{}, map[string]string{
		"01A": "2026-03-09",
		"01B": "2026-03-09",
	})
	if len(rows) != 2 {
		t.Fatalf("unexpected row count: %d", len(rows))
	}
	if !strings.Contains(rows[0].Label, "Parent") || rows[0].Meta != "todo/open" {
		t.Fatalf("unexpected parent row: %+v", rows[0])
	}
	if !strings.Contains(rows[1].Label, "Child") || !strings.Contains(rows[1].Label, "└─") || rows[1].Meta != "memo/blocked" {
		t.Fatalf("unexpected child row: %+v", rows[1])
	}
	if rows[1].DueOn != "2026-03-09" || !rows[1].DueInherited {
		t.Fatalf("expected inherited due on child row, got %+v", rows[1])
	}
}

func TestFlattenCockpitTreeRowsSkipsCollapsedChildren(t *testing.T) {
	nodes := []shelf.TreeNode{
		{
			Task: shelf.Task{ID: "01A", Title: "Parent", Kind: "todo", Status: "open"},
			Children: []shelf.TreeNode{
				{Task: shelf.Task{ID: "01B", Title: "Child", Kind: "memo", Status: "blocked"}},
			},
		},
	}
	rows := flattenCockpitTreeRows(nodes, "", true, false, map[string]struct{}{"01A": {}}, map[string]string{})
	if len(rows) != 1 {
		t.Fatalf("expected collapsed tree to hide children, got %d rows", len(rows))
	}
	if !rows[0].Collapsed || !rows[0].HasChildren {
		t.Fatalf("expected parent row marked collapsed with children, got %+v", rows[0])
	}
	if !strings.Contains(rows[0].Label, "[+]") {
		t.Fatalf("expected collapsed marker in label, got %q", rows[0].Label)
	}
}

func TestCreateTaskFromAddModeUsesSelectedTaskAsParent(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(parent.ID)
	model.beginAddMode(false)
	if err := model.createTaskFromAddMode("Child task"); err != nil {
		t.Fatalf("createTaskFromAddMode failed: %v", err)
	}
	created := model.taskByID[model.selectedTaskID]
	if created.Parent != parent.ID {
		t.Fatalf("expected child task parent %s, got %+v", parent.ID, created)
	}
}

func TestCreateTaskFromAddModeAtRootClearsParent(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(parent.ID)
	model.beginAddMode(true)
	if err := model.createTaskFromAddMode("Root task"); err != nil {
		t.Fatalf("createTaskFromAddMode failed: %v", err)
	}
	created := model.taskByID[model.selectedTaskID]
	if created.Parent != "" {
		t.Fatalf("expected root task without parent, got %+v", created)
	}
}

func TestUpdateAddModeUsesTabForFieldSwitchAndEnterForCreate(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(parent.ID)
	model.beginAddMode(false)
	model.addTitle = "Created"
	updated, _ := model.updateAddMode(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(calendarTUIModel)
	if model.addField != calendarAddFieldKind {
		t.Fatalf("expected tab to move to kind field, got %v", model.addField)
	}
	updated, _ = model.updateAddMode(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(calendarTUIModel)
	if model.addMode {
		t.Fatal("expected enter to confirm add")
	}
	created := model.taskByID[model.selectedTaskID]
	if created.Title != "Created" {
		t.Fatalf("expected created task selected, got %+v", created)
	}
}

func TestUpdateAddModeAllowsRuneInputAndShiftTabCycle(t *testing.T) {
	model := calendarTUIModel{
		addMode:        true,
		addField:       calendarAddFieldTitle,
		defaultKind:    "todo",
		addKind:        "todo",
		addTitleCursor: 0,
	}
	updated, _ := model.updateAddMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = updated.(calendarTUIModel)
	if model.addTitle != "q" {
		t.Fatalf("expected q to be appended to title, got %q", model.addTitle)
	}

	updated, _ = model.updateAddMode(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(calendarTUIModel)
	if model.addField != calendarAddFieldKind {
		t.Fatalf("expected tab to move to kind, got %v", model.addField)
	}

	updated, _ = model.updateAddMode(tea.KeyMsg{Type: tea.KeyShiftTab})
	model = updated.(calendarTUIModel)
	if model.addField != calendarAddFieldTitle {
		t.Fatalf("expected shift+tab to move back to title, got %v", model.addField)
	}
}

func TestUpdateAddModeSupportsMidStringEditing(t *testing.T) {
	model := calendarTUIModel{
		addMode:        true,
		addField:       calendarAddFieldTitle,
		addTitle:       "ab",
		addTitleCursor: 2,
		defaultKind:    "todo",
		addKind:        "todo",
	}

	updated, _ := model.updateAddMode(tea.KeyMsg{Type: tea.KeyLeft})
	model = updated.(calendarTUIModel)
	if model.addTitleCursor != 1 {
		t.Fatalf("expected cursor to move left, got %d", model.addTitleCursor)
	}

	updated, _ = model.updateAddMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
	model = updated.(calendarTUIModel)
	if model.addTitle != "aXb" || model.addTitleCursor != 2 {
		t.Fatalf("unexpected title edit: value=%q cursor=%d", model.addTitle, model.addTitleCursor)
	}
}

func TestBeginTagModeUsesMarkedTasksForBulkEdit(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open", DueOn: "2026-03-09", Tags: []string{"alpha"}})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open", DueOn: "2026-03-09", Tags: []string{"beta"}})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(first.ID)
	model.tagChoices = []string{"gamma"}
	model.markedTaskIDs = map[string]struct{}{first.ID: {}, second.ID: {}}

	model.beginTagMode()

	if !model.tagBulkMode {
		t.Fatal("expected beginTagMode to enter bulk mode for marked tasks")
	}
	if len(model.tagSelection) != 0 {
		t.Fatalf("expected no exact selection in bulk mode, got %+v", model.tagSelection)
	}
	for _, tag := range []string{"alpha", "beta", "gamma"} {
		if !containsTag(model.tagChoices, tag) {
			t.Fatalf("expected tag choices to include %q, got %+v", tag, model.tagChoices)
		}
	}
}

func TestUpdateTagModeUsesInputModeForNewTag(t *testing.T) {
	model := calendarTUIModel{
		tagMode:      true,
		tagChoices:   []string{"alpha"},
		tagSelection: []string{},
		tagIndex:     1,
	}
	updated, _ := model.updateTagMode(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(calendarTUIModel)
	if !model.tagInputMode {
		t.Fatal("expected enter on add-new-tag row to enter input mode")
	}

	updated, _ = model.updateTagMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = updated.(calendarTUIModel)
	if model.tagInputValue != "q" {
		t.Fatalf("expected q to be recorded as tag input, got %q", model.tagInputValue)
	}
	if !model.tagMode {
		t.Fatal("expected tag picker to stay open while typing")
	}
}

func TestUpdateTagModeCyclesBulkStatesWithSpace(t *testing.T) {
	model := calendarTUIModel{
		tagMode:       true,
		tagBulkMode:   true,
		tagChoices:    []string{"alpha"},
		tagBulkStates: map[string]calendarTagBulkState{},
		tagIndex:      2,
	}

	updated, _ := model.updateTagMode(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(calendarTUIModel)
	if got := model.tagBulkState("alpha"); got != calendarTagBulkStateAdd {
		t.Fatalf("expected add state, got %v", got)
	}

	updated, _ = model.updateTagMode(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(calendarTUIModel)
	if got := model.tagBulkState("alpha"); got != calendarTagBulkStateRemove {
		t.Fatalf("expected remove state, got %v", got)
	}

	updated, _ = model.updateTagMode(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(calendarTUIModel)
	if got := model.tagBulkState("alpha"); got != calendarTagBulkStateUnchanged {
		t.Fatalf("expected unchanged state, got %v", got)
	}
}

func TestUpdateTagModeSupportsMidStringEditing(t *testing.T) {
	model := calendarTUIModel{
		tagMode:        true,
		tagInputMode:   true,
		tagInputValue:  "ab",
		tagInputCursor: 2,
	}

	updated, _ := model.updateTagMode(tea.KeyMsg{Type: tea.KeyLeft})
	model = updated.(calendarTUIModel)
	if model.tagInputCursor != 1 {
		t.Fatalf("expected cursor to move left, got %d", model.tagInputCursor)
	}

	updated, _ = model.updateTagMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
	model = updated.(calendarTUIModel)
	if model.tagInputValue != "aXb" || model.tagInputCursor != 2 {
		t.Fatalf("unexpected tag input edit: value=%q cursor=%d", model.tagInputValue, model.tagInputCursor)
	}
}

func TestUpdateTagModeAddsNewTagAsBulkAdd(t *testing.T) {
	model := calendarTUIModel{
		tagMode:        true,
		tagBulkMode:    true,
		tagChoices:     []string{"alpha"},
		tagBulkStates:  map[string]calendarTagBulkState{},
		tagInputMode:   true,
		tagInputValue:  "beta",
		tagInputCursor: 4,
	}

	updated, _ := model.updateTagMode(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(calendarTUIModel)
	if !containsTag(model.tagChoices, "beta") {
		t.Fatalf("expected new tag in choices, got %+v", model.tagChoices)
	}
	if got := model.tagBulkState("beta"); got != calendarTagBulkStateAdd {
		t.Fatalf("expected new tag to default to add state, got %v", got)
	}
}

func TestUpdateTagModeUsesSpaceForToggleAndCtrlSForSave(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Task", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(task.ID)
	model.beginTagMode()
	model.tagChoices = []string{"alpha"}
	model.tagIndex = 2

	updated, _ := model.updateTagMode(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(calendarTUIModel)
	if len(model.tagSelection) != 0 {
		t.Fatalf("expected enter on a tag row to not toggle, got %+v", model.tagSelection)
	}

	updated, _ = model.updateTagMode(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(calendarTUIModel)
	if len(model.tagSelection) != 1 || model.tagSelection[0] != "alpha" {
		t.Fatalf("expected space to toggle tag selection, got %+v", model.tagSelection)
	}

	updated, _ = model.updateTagMode(tea.KeyMsg{Type: tea.KeyCtrlS})
	model = updated.(calendarTUIModel)
	if model.tagMode {
		t.Fatal("expected ctrl+s to save and close tag mode")
	}
	reloadedTask := model.taskByID[task.ID]
	if len(reloadedTask.Tags) != 1 || reloadedTask.Tags[0] != "alpha" {
		t.Fatalf("expected saved tag on task, got %+v", reloadedTask.Tags)
	}
}

func TestApplySelectedTaskTagDeltaUsesMarkedTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "First",
		Kind:   "todo",
		Status: "open",
		DueOn:  "2026-03-09",
		Tags:   []string{"alpha", "gamma"},
	})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Second",
		Kind:   "todo",
		Status: "open",
		DueOn:  "2026-03-09",
		Tags:   []string{"beta", "gamma"},
	})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(second.ID)
	model.markedTaskIDs = map[string]struct{}{first.ID: {}, second.ID: {}}

	if err := model.applySelectedTaskTagDelta(shelf.SetTaskInput{
		AddTags:    []string{"delta"},
		RemoveTags: []string{"gamma"},
	}); err != nil {
		t.Fatalf("applySelectedTaskTagDelta failed: %v", err)
	}

	reloadedFirst, err := shelf.EnsureTaskExists(root, first.ID)
	if err != nil {
		t.Fatalf("EnsureTaskExists failed for first: %v", err)
	}
	if !slices.Equal(reloadedFirst.Tags, []string{"alpha", "delta"}) {
		t.Fatalf("unexpected first tags: %+v", reloadedFirst.Tags)
	}
	reloadedSecond, err := shelf.EnsureTaskExists(root, second.ID)
	if err != nil {
		t.Fatalf("EnsureTaskExists failed for second: %v", err)
	}
	if !slices.Equal(reloadedSecond.Tags, []string{"beta", "delta"}) {
		t.Fatalf("unexpected second tags: %+v", reloadedSecond.Tags)
	}
	if model.markedCount() != 0 {
		t.Fatalf("expected marks cleared after bulk tag update, got %d", model.markedCount())
	}
	if model.selectedTaskID != first.ID {
		t.Fatalf("expected first updated task selected, got %s", model.selectedTaskID)
	}
	if model.message != "Updated tags for 2 tasks (+delta; -gamma)" {
		t.Fatalf("unexpected message: %s", model.message)
	}
	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if !containsTag(cfg.Tags, "delta") {
		t.Fatalf("expected config tags to include delta, got %+v", cfg.Tags)
	}
}

func TestSelectTaskByIDInTreeSyncsFocusedDateToTask(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Other", Kind: "todo", Status: "open", DueOn: "2026-03-09"}); err != nil {
		t.Fatalf("add other task failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Due", Kind: "todo", Status: "open", DueOn: "2026-03-12"})
	if err != nil {
		t.Fatalf("add task failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.switchMode(calendarModeTree)
	model.selectTaskByID(task.ID)
	if model.focusedDayLabel() != "2026-03-12" {
		t.Fatalf("expected focused day synced to task due date, got %s", model.focusedDayLabel())
	}
	section := model.focusedDaySection()
	if section == nil || len(section.Items) != 1 {
		t.Fatalf("expected selected day section rebuilt for synced date, got %+v", section)
	}
	if section.Items[0].Task.ID != task.ID {
		t.Fatalf("expected selected day contents synced to selected task, got %+v", section.Items)
	}
}

func TestCalendarSwitchModeKeepsSelectedTaskWhenPossible(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Today",
		Kind:   "todo",
		Status: "open",
		DueOn:  time.Now().Local().Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, startOfWeek(time.Now().Local()), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.selectTaskByID(task.ID)
	model.switchMode(calendarModeReview)
	if model.mode != calendarModeReview {
		t.Fatalf("unexpected mode: %s", model.mode)
	}
	if model.selectedTaskID != task.ID {
		t.Fatalf("expected selected task preserved, got %s", model.selectedTaskID)
	}
	model.switchMode(calendarModeTree)
	if model.mode != calendarModeTree {
		t.Fatalf("unexpected mode: %s", model.mode)
	}
	if model.selectedTaskID != task.ID {
		t.Fatalf("expected selected task preserved in tree mode, got %s", model.selectedTaskID)
	}
	model.switchMode(calendarModeBoard)
	if model.mode != calendarModeBoard {
		t.Fatalf("unexpected mode: %s", model.mode)
	}
	if model.selectedTaskID != task.ID {
		t.Fatalf("expected selected task preserved in board mode, got %s", model.selectedTaskID)
	}
}

func TestCalendarBoardModeMovesAcrossColumns(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	openTask, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Open task",
		Kind:   "todo",
		Status: "open",
		DueOn:  time.Now().Local().Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("add open failed: %v", err)
	}
	doneTask, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Done task",
		Kind:   "todo",
		Status: "done",
		DueOn:  time.Now().Local().Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("add done failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 7, []shelf.Status{"open", "done"}, calendarTUIOptions{
		Mode:   calendarModeBoard,
		Filter: shelf.TaskFilter{Statuses: []shelf.Status{"open", "done"}},
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	model.selectTaskByID(openTask.ID)
	if task, ok := model.selectedTask(); !ok || task.ID != openTask.ID {
		t.Fatalf("expected open task selected before column move: %+v ok=%t", task, ok)
	}
	model.moveBoardColumn(1)
	if task, ok := model.selectedTask(); !ok || task.ID != doneTask.ID {
		t.Fatalf("unexpected selected task after moveBoardColumn: %+v ok=%t", task, ok)
	}
}

func TestTreeModeMoveSelectionUnderAnotherTask(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parentA, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent A", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add parentA failed: %v", err)
	}
	child, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Child", Kind: "todo", Status: "open", Parent: parentA.ID})
	if err != nil {
		t.Fatalf("add child failed: %v", err)
	}
	parentB, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent B", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add parentB failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 30, []shelf.Status{"open", "in_progress", "blocked", "done", "cancelled"}, calendarTUIOptions{
		Mode:   calendarModeTree,
		Filter: shelf.TaskFilter{Statuses: []shelf.Status{"open", "in_progress", "blocked", "done", "cancelled"}},
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	model.selectTaskByID(child.ID)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	treeModel := updatedModel.(calendarTUIModel)
	if !treeModel.moveMode {
		t.Fatal("expected move mode to start")
	}
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	treeModel = updatedModel.(calendarTUIModel)
	if task, ok := treeModel.selectedTask(); !ok || task.ID != parentB.ID {
		t.Fatalf("expected move target parentB, got %+v ok=%t", task, ok)
	}
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	treeModel = updatedModel.(calendarTUIModel)
	store := shelf.NewTaskStore(root)
	tasks, err := store.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	byID := map[string]shelf.Task{}
	for _, task := range tasks {
		byID[task.ID] = task
	}
	if byID[child.ID].Parent != parentB.ID {
		t.Fatalf("expected child parent updated to parentB, got %q", byID[child.ID].Parent)
	}
	if treeModel.moveMode {
		t.Fatal("expected move mode to finish")
	}
}

func TestTreeModeMoveSelectionToRoot(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	child, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Child", Kind: "todo", Status: "open", Parent: parent.ID})
	if err != nil {
		t.Fatalf("add child failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 30, []shelf.Status{"open", "in_progress", "blocked", "done", "cancelled"}, calendarTUIOptions{
		Mode:   calendarModeTree,
		Filter: shelf.TaskFilter{Statuses: []shelf.Status{"open", "in_progress", "blocked", "done", "cancelled"}},
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	model.selectTaskByID(child.ID)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	treeModel := updatedModel.(calendarTUIModel)
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	treeModel = updatedModel.(calendarTUIModel)
	if treeModel.treeRowIndex != -1 {
		t.Fatalf("expected root move target, got treeRowIndex=%d", treeModel.treeRowIndex)
	}
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	treeModel = updatedModel.(calendarTUIModel)
	updated, err := shelf.EnsureTaskExists(root, child.ID)
	if err != nil {
		t.Fatalf("ensure task failed: %v", err)
	}
	if updated.Parent != "" {
		t.Fatalf("expected child moved to root, got parent=%q", updated.Parent)
	}
	if treeModel.moveMode {
		t.Fatal("expected move mode finished after move to root")
	}
}

func TestTreeModeMoveSelectionFreezesRangeMarksAtMoveStart(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open"}); err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open"}); err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Third", Kind: "todo", Status: "open"}); err != nil {
		t.Fatalf("add third failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Target", Kind: "todo", Status: "open"}); err != nil {
		t.Fatalf("add target failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 30, []shelf.Status{"open", "in_progress", "blocked", "done", "cancelled"}, calendarTUIOptions{
		Mode:   calendarModeTree,
		Filter: shelf.TaskFilter{Statuses: []shelf.Status{"open", "in_progress", "blocked", "done", "cancelled"}},
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	if len(model.treeRows) != 4 {
		t.Fatalf("expected 4 tree rows, got %d", len(model.treeRows))
	}
	selectedIDs := []string{
		model.treeRows[0].Task.ID,
		model.treeRows[1].Task.ID,
		model.treeRows[2].Task.ID,
	}
	targetID := model.treeRows[3].Task.ID
	model.selectTaskByID(selectedIDs[0])
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	treeModel := updatedModel.(calendarTUIModel)
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	treeModel = updatedModel.(calendarTUIModel)
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	treeModel = updatedModel.(calendarTUIModel)
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	treeModel = updatedModel.(calendarTUIModel)
	if treeModel.rangeMarkMode {
		t.Fatal("expected m to stop range select before move mode")
	}
	if len(treeModel.moveSourceIDs) != 3 {
		t.Fatalf("expected 3 frozen move sources, got %d", len(treeModel.moveSourceIDs))
	}
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	treeModel = updatedModel.(calendarTUIModel)
	if treeModel.markedCount() != 3 {
		t.Fatalf("expected marks frozen after move start, got %d", treeModel.markedCount())
	}
	if treeModel.isMarkedTask(targetID) {
		t.Fatal("expected move target not to be added to marked tasks after move start")
	}
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	treeModel = updatedModel.(calendarTUIModel)
	store := shelf.NewTaskStore(root)
	tasks, err := store.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	byID := map[string]shelf.Task{}
	for _, task := range tasks {
		byID[task.ID] = task
	}
	for _, movedID := range selectedIDs {
		if byID[movedID].Parent != targetID {
			t.Fatalf("expected %s moved under target, got parent=%q", movedID, byID[movedID].Parent)
		}
	}
	if byID[targetID].Parent != "" {
		t.Fatalf("expected target to stay at root, got parent=%q", byID[targetID].Parent)
	}
	if treeModel.moveMode {
		t.Fatal("expected move mode finished after apply")
	}
}

func TestTreeModeCollapseAndExpandCurrentNode(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	parent, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Parent", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	child, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Child", Kind: "todo", Status: "open", Parent: parent.ID})
	if err != nil {
		t.Fatalf("add child failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 30, []shelf.Status{"open"}, calendarTUIOptions{
		Mode:   calendarModeTree,
		Filter: shelf.TaskFilter{Statuses: []shelf.Status{"open"}},
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	model.selectTaskByID(parent.ID)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	treeModel := updatedModel.(calendarTUIModel)
	if len(treeModel.treeRows) != 1 {
		t.Fatalf("expected child hidden after collapse, got %d rows", len(treeModel.treeRows))
	}
	if !treeModel.treeRows[0].Collapsed {
		t.Fatalf("expected parent row collapsed, got %+v", treeModel.treeRows[0])
	}
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	treeModel = updatedModel.(calendarTUIModel)
	if len(treeModel.treeRows) != 2 {
		t.Fatalf("expected child visible after expand, got %d rows", len(treeModel.treeRows))
	}
	if treeModel.treeRows[0].Collapsed {
		t.Fatalf("expected parent row expanded, got %+v", treeModel.treeRows[0])
	}
	if treeModel.treeRows[1].Task.ID != child.ID {
		t.Fatalf("expected child row restored, got %+v", treeModel.treeRows[1])
	}
}

func TestTreeModeBulkStatusChangeUsesMarkedTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 30, []shelf.Status{"open", "done"}, calendarTUIOptions{
		Mode:   calendarModeTree,
		Filter: shelf.TaskFilter{Statuses: []shelf.Status{"open", "done"}},
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	model.selectTaskByID(first.ID)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	treeModel := updatedModel.(calendarTUIModel)
	treeModel.selectTaskByID(second.ID)
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	treeModel = updatedModel.(calendarTUIModel)
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	treeModel = updatedModel.(calendarTUIModel)
	store := shelf.NewTaskStore(root)
	tasks, err := store.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	for _, task := range tasks {
		if task.ID == first.ID || task.ID == second.ID {
			if task.Status != "done" {
				t.Fatalf("expected marked tasks to become done, got %+v", task)
			}
		}
	}
	if treeModel.markedCount() != 0 {
		t.Fatalf("expected marks cleared after bulk status change, got %d", treeModel.markedCount())
	}
}

func TestTreeModeRangeSelectMarksContinuousTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	third, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Third", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add third failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 30, []shelf.Status{"open", "done"}, calendarTUIOptions{
		Mode:   calendarModeTree,
		Filter: shelf.TaskFilter{Statuses: []shelf.Status{"open", "done"}},
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	model.selectTaskByID(first.ID)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	treeModel := updatedModel.(calendarTUIModel)
	if !treeModel.rangeMarkMode {
		t.Fatal("expected range mark mode enabled")
	}
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	treeModel = updatedModel.(calendarTUIModel)
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	treeModel = updatedModel.(calendarTUIModel)
	for _, taskID := range []string{first.ID, second.ID, third.ID} {
		if !treeModel.isMarkedTask(taskID) {
			t.Fatalf("expected task %s marked in range", taskID)
		}
	}
}

func TestTreeModeRangeSelectPreservesExistingMarksWhenRestarted(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open"}); err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open"}); err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Third", Kind: "todo", Status: "open"}); err != nil {
		t.Fatalf("add third failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 30, []shelf.Status{"open", "done"}, calendarTUIOptions{
		Mode:   calendarModeTree,
		Filter: shelf.TaskFilter{Statuses: []shelf.Status{"open", "done"}},
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	if len(model.treeRows) != 3 {
		t.Fatalf("expected 3 tree rows, got %d", len(model.treeRows))
	}
	orderedIDs := []string{
		model.treeRows[0].Task.ID,
		model.treeRows[1].Task.ID,
		model.treeRows[2].Task.ID,
	}
	model.selectTaskByID(orderedIDs[0])
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	treeModel := updatedModel.(calendarTUIModel)
	treeModel.selectTaskByID(orderedIDs[1])
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	treeModel = updatedModel.(calendarTUIModel)
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	treeModel = updatedModel.(calendarTUIModel)
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	treeModel = updatedModel.(calendarTUIModel)
	for _, taskID := range orderedIDs {
		if !treeModel.isMarkedTask(taskID) {
			t.Fatalf("expected existing marks preserved after finishing range select")
		}
	}
	treeModel.selectTaskByID(orderedIDs[2])
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	treeModel = updatedModel.(calendarTUIModel)
	for _, taskID := range orderedIDs {
		if !treeModel.isMarkedTask(taskID) {
			t.Fatalf("expected existing marks preserved when restarting range select")
		}
	}
}

func TestTreeModeUClearsAllMarks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 30, []shelf.Status{"open", "done"}, calendarTUIOptions{
		Mode:   calendarModeTree,
		Filter: shelf.TaskFilter{Statuses: []shelf.Status{"open", "done"}},
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	model.selectTaskByID(first.ID)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	treeModel := updatedModel.(calendarTUIModel)
	treeModel.selectTaskByID(second.ID)
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	treeModel = updatedModel.(calendarTUIModel)
	updatedModel, _ = treeModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	treeModel = updatedModel.(calendarTUIModel)
	if treeModel.markedCount() != 0 {
		t.Fatalf("expected all marks cleared, got %d", treeModel.markedCount())
	}
}

func TestLeaveToNormalModeClearsTransientStateWithoutDroppingMarks(t *testing.T) {
	model := calendarTUIModel{
		mode:          calendarModeTree,
		rangeMarkMode: true,
		rangeAnchorID: "01A",
		rangeBaseIDs:  map[string]struct{}{"01A": {}},
		markedTaskIDs: map[string]struct{}{"01A": {}, "01B": {}},
		showHelp:      true,
	}
	if !model.leaveToNormalMode() {
		t.Fatal("expected leaveToNormalMode to report a state change")
	}
	if model.rangeMarkMode {
		t.Fatal("expected leaveToNormalMode to leave range mode")
	}
	if model.showHelp {
		t.Fatal("expected leaveToNormalMode to close help")
	}
	if model.markedCount() != 2 {
		t.Fatalf("expected leaveToNormalMode to keep marks, got %d", model.markedCount())
	}
}

func TestBoardModeBulkStatusChangeUsesMarkedTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	openTask, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Open", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add open failed: %v", err)
	}
	blockedTask, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Blocked", Kind: "todo", Status: "blocked"})
	if err != nil {
		t.Fatalf("add blocked failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 30, []shelf.Status{"open", "blocked", "done"}, calendarTUIOptions{
		Mode:   calendarModeBoard,
		Filter: shelf.TaskFilter{Statuses: []shelf.Status{"open", "blocked", "done"}},
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	if task, ok := model.selectedTask(); !ok || task.ID != openTask.ID {
		t.Fatalf("unexpected initial selected task: %+v ok=%t", task, ok)
	}
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	boardModel := updatedModel.(calendarTUIModel)
	updatedModel, _ = boardModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	boardModel = updatedModel.(calendarTUIModel)
	if task, ok := boardModel.selectedTask(); !ok || task.ID != blockedTask.ID {
		t.Fatalf("expected blocked task selected, got %+v ok=%t", task, ok)
	}
	updatedModel, _ = boardModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	boardModel = updatedModel.(calendarTUIModel)
	updatedModel, _ = boardModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	boardModel = updatedModel.(calendarTUIModel)
	store := shelf.NewTaskStore(root)
	tasks, err := store.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	for _, task := range tasks {
		if task.ID == openTask.ID || task.ID == blockedTask.ID {
			if task.Status != "done" {
				t.Fatalf("expected marked board tasks to become done, got %+v", task)
			}
		}
	}
	if boardModel.markedCount() != 0 {
		t.Fatalf("expected board marks cleared after bulk status change, got %d", boardModel.markedCount())
	}
}

func TestBoardModeRangeSelectMarksContinuousTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	openTask, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Open", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add open failed: %v", err)
	}
	blockedTask, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Blocked", Kind: "todo", Status: "blocked"})
	if err != nil {
		t.Fatalf("add blocked failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 30, []shelf.Status{"open", "blocked", "done"}, calendarTUIOptions{
		Mode:   calendarModeBoard,
		Filter: shelf.TaskFilter{Statuses: []shelf.Status{"open", "blocked", "done"}},
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	boardModel := updatedModel.(calendarTUIModel)
	if !boardModel.rangeMarkMode {
		t.Fatal("expected board range mark mode enabled")
	}
	updatedModel, _ = boardModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	boardModel = updatedModel.(calendarTUIModel)
	if !boardModel.isMarkedTask(openTask.ID) || !boardModel.isMarkedTask(blockedTask.ID) {
		t.Fatalf("expected range select to mark both board tasks")
	}
}

func TestCalendarHelpToggle(t *testing.T) {
	model := calendarTUIModel{}
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	toggled := updatedModel.(calendarTUIModel)
	if !toggled.showHelp {
		t.Fatal("expected help overlay to toggle on")
	}
	updatedModel, _ = toggled.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	toggled = updatedModel.(calendarTUIModel)
	if toggled.showHelp {
		t.Fatal("expected help overlay to toggle off")
	}
}

func TestCalendarQClosesHelpBeforeQuit(t *testing.T) {
	model := calendarTUIModel{showHelp: true}
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	toggled := updatedModel.(calendarTUIModel)
	if toggled.showHelp {
		t.Fatal("expected q to close help when help is visible")
	}
	if cmd != nil {
		t.Fatal("expected q on help to close help without quitting")
	}
}

func TestCalendarEscDoesNotQuit(t *testing.T) {
	model := calendarTUIModel{}
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	toggled := updatedModel.(calendarTUIModel)
	if toggled.showHelp {
		t.Fatal("expected esc in normal mode to leave help closed")
	}
	if cmd != nil {
		t.Fatal("expected esc in normal mode to not quit")
	}
}

func TestCalendarEscClosesHelpWithoutQuit(t *testing.T) {
	model := calendarTUIModel{showHelp: true}
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	toggled := updatedModel.(calendarTUIModel)
	if toggled.showHelp {
		t.Fatal("expected esc to close help when help is visible")
	}
	if cmd != nil {
		t.Fatal("expected esc on help to close help without quitting")
	}
}

func TestRenderCockpitHeaderIsSingleLine(t *testing.T) {
	model := calendarTUIModel{
		mode:     calendarModeCalendar,
		days:     []calendarDay{{Date: "2026-03-09"}, {Date: "2026-03-10"}},
		statuses: []shelf.Status{"open", "in_progress", "blocked"},
		width:    120,
	}
	header := renderCockpitHeader(model, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local))
	if strings.Contains(header, "\n") {
		t.Fatalf("expected single-line header, got: %q", header)
	}
	if !strings.Contains(header, "?:help") {
		t.Fatalf("expected compact help hint, got: %q", header)
	}
}

func TestRenderCockpitHeaderShowsTransientModeHint(t *testing.T) {
	model := calendarTUIModel{
		mode:          calendarModeTree,
		days:          []calendarDay{{Date: "2026-03-09"}, {Date: "2026-03-10"}},
		statuses:      []shelf.Status{"open", "in_progress", "blocked"},
		width:         200,
		showHelp:      true,
		rangeMarkMode: true,
	}
	header := renderCockpitHeader(model, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local))
	if !strings.Contains(header, "Mode") || !strings.Contains(header, "help") || !strings.Contains(header, "range") {
		t.Fatalf("expected transient mode labels in header, got: %q", header)
	}
	if !strings.Contains(header, "Ctrl+[: normal") {
		t.Fatalf("expected normal-mode hint in header, got: %q", header)
	}
}

func TestReviewMainPaneUsesContextStripInsteadOfMonthGrid(t *testing.T) {
	today := time.Now().Local().Format("2006-01-02")
	model := calendarTUIModel{
		mode: calendarModeReview,
		days: []calendarDay{
			{Date: today},
		},
		visibleTasks: []shelf.Task{
			{ID: "01A", Title: "Inbox", Kind: "inbox", Status: "open"},
			{ID: "01B", Title: "Today", Kind: "todo", Status: "open", DueOn: today},
		},
		sections: []calendarSection{
			{ID: calendarSectionFocusedDay, Title: "Selected Day"},
			{ID: calendarSectionInbox, Title: "Inbox", Items: []calendarSectionItem{{Task: shelf.Task{ID: "01A", Title: "Inbox"}}}},
			{ID: calendarSectionToday, Title: "Today", Items: []calendarSectionItem{{Task: shelf.Task{ID: "01B", Title: "Today"}}}},
		},
		sectionRows: map[calendarSectionID]int{},
	}
	month := calendarMonthView{Label: "March 2026"}
	rendered := renderCalendarMainPane(model, month, 90, 18, true)
	if strings.Contains(rendered, "March 2026") {
		t.Fatalf("review pane should not render month grid label: %q", rendered)
	}
	if !strings.Contains(rendered, "Date ") || !strings.Contains(rendered, "Inbox 1") {
		t.Fatalf("review pane should render context strip: %q", rendered)
	}
}

func TestCalendarViewUsesSidebarForFocusedDayTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Focused task",
		Kind:   "todo",
		Status: "open",
		DueOn:  "2026-03-09",
	}); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.width = 120
	model.height = 24
	rendered := model.View()
	if !strings.Contains(rendered, "Selected Day: 2026-03-09") {
		t.Fatalf("calendar view should render selected day sidebar: %q", rendered)
	}
	if !strings.Contains(rendered, "n/p: task switch") {
		t.Fatalf("calendar view should show selected day switch hint: %q", rendered)
	}
	if strings.Index(rendered, "Selected Day: 2026-03-09") > strings.Index(rendered, "Inspector") {
		t.Fatalf("selected day pane should be above inspector in calendar sidebar: %q", rendered)
	}
}

func TestCalendarNPSwitchesFocusedDayTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	initialTask, ok := model.selectedTask()
	if !ok {
		t.Fatalf("expected an initial selected task, got %+v ok=%t", initialTask, ok)
	}
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	calendarModel := updatedModel.(calendarTUIModel)
	nextTask, ok := calendarModel.selectedTask()
	if !ok {
		t.Fatalf("expected n to select another focused-day task, got %+v ok=%t", nextTask, ok)
	}
	if nextTask.ID == initialTask.ID {
		t.Fatalf("expected n to move selection to a different task, still on %+v", nextTask)
	}
	updatedModel, _ = calendarModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	calendarModel = updatedModel.(calendarTUIModel)
	task, ok := calendarModel.selectedTask()
	if !ok || task.ID != initialTask.ID {
		t.Fatalf("expected p to return to the initial focused-day task, got %+v ok=%t", task, ok)
	}
	taskIDs := map[string]struct{}{
		first.ID:  {},
		second.ID: {},
	}
	if _, ok := taskIDs[initialTask.ID]; !ok {
		t.Fatalf("unexpected initial selected task ID: %s", initialTask.ID)
	}
	if _, ok := taskIDs[nextTask.ID]; !ok {
		t.Fatalf("unexpected next selected task ID: %s", nextTask.ID)
	}
}

func TestReviewViewUsesSidebarCalendar(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	today := time.Now().Local().Format("2006-01-02")
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Today task",
		Kind:   "todo",
		Status: "open",
		DueOn:  today,
	}); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 7, []shelf.Status{"open", "in_progress", "blocked"}, calendarTUIOptions{
		Mode:   calendarModeReview,
		ShowID: false,
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	model.width = 120
	model.height = 24
	rendered := model.View()
	if !strings.Contains(rendered, "Calendar") {
		t.Fatalf("non-calendar mode should render sidebar calendar: %q", rendered)
	}
	if !strings.Contains(rendered, "Sun") {
		t.Fatalf("non-calendar sidebar calendar should render the month grid: %q", rendered)
	}
	if !strings.Contains(rendered, "selection synced") {
		t.Fatalf("sidebar calendar should show synced hint: %q", rendered)
	}
}

func TestSidebarCalendarNavigationMovesFocusedDate(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	today := time.Now().Local().Format("2006-01-02")
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Today task",
		Kind:   "todo",
		Status: "open",
		DueOn:  today,
	}); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, startOfWeek(time.Now().Local()), 14, []shelf.Status{"open", "in_progress", "blocked"}, calendarTUIOptions{
		Mode:   calendarModeReview,
		ShowID: false,
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	original := model.focusedDayLabel()
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	reviewModel := updatedModel.(calendarTUIModel)
	if reviewModel.pane != calendarPaneInspector {
		t.Fatalf("expected sidebar pane active after tab, got %v", reviewModel.pane)
	}
	updatedModel, _ = reviewModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	reviewModel = updatedModel.(calendarTUIModel)
	if reviewModel.focusedDayLabel() == original {
		t.Fatalf("expected sidebar calendar navigation to move focused day, still %s", reviewModel.focusedDayLabel())
	}
}

func TestSidebarCalendarNavigationSyncsMainSelection(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "First day",
		Kind:   "todo",
		Status: "open",
		DueOn:  "2026-03-09",
	})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Second day",
		Kind:   "todo",
		Status: "open",
		DueOn:  "2026-03-10",
	})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, calendarTUIOptions{
		Mode:   calendarModeReview,
		ShowID: false,
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	model.selectTaskByID(first.ID)
	model.pane = calendarPaneInspector
	model.moveSidebarFocusByDays(1)
	if model.focusedDayLabel() != "2026-03-10" {
		t.Fatalf("expected focused day to move to 2026-03-10, got %s", model.focusedDayLabel())
	}
	if model.selectedTaskID != second.ID {
		t.Fatalf("expected main selection to follow sidebar date change, got %s", model.selectedTaskID)
	}
}

func TestReviewSelectTaskPrefersOperationalSectionAndSyncsSidebarDate(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	today := time.Now().Local().Format("2006-01-02")
	tomorrow := time.Now().Local().AddDate(0, 0, 1).Format("2006-01-02")
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Today", Kind: "todo", Status: "open", DueOn: today})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Tomorrow", Kind: "todo", Status: "open", DueOn: tomorrow})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, time.Now().Local(), 7, []shelf.Status{"open"}, calendarTUIOptions{
		Mode: calendarModeReview,
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	model.selectTaskByID(first.ID)
	if model.sectionIndex == 0 {
		t.Fatalf("expected review selection to prefer operational sections over Selected Day")
	}
	model.selectTaskByID(second.ID)
	if model.selectedTaskID != second.ID {
		t.Fatalf("expected main review move to select second task, got %s", model.selectedTaskID)
	}
	if model.sectionIndex == 0 {
		t.Fatalf("expected selected task to stay in an operational section, got section %d", model.sectionIndex)
	}
	if model.focusedDayLabel() != tomorrow {
		t.Fatalf("expected sidebar date to sync to tomorrow, got %s", model.focusedDayLabel())
	}
}

func TestReviewMoveWithinBlockedSectionKeepsCurrentTab(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	today := time.Now().Local().Format("2006-01-02")
	tomorrow := time.Now().Local().AddDate(0, 0, 1).Format("2006-01-02")
	todayBlocked, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Today blocked", Kind: "todo", Status: "blocked", DueOn: today})
	if err != nil {
		t.Fatalf("add today blocked failed: %v", err)
	}
	tomorrowBlocked, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Tomorrow blocked", Kind: "todo", Status: "blocked", DueOn: tomorrow})
	if err != nil {
		t.Fatalf("add tomorrow blocked failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, time.Now().Local(), 7, []shelf.Status{"blocked"}, calendarTUIOptions{
		Mode: calendarModeReview,
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	blockedIndex := -1
	for i, section := range model.sections {
		if section.ID == calendarSectionBlocked {
			blockedIndex = i
			break
		}
	}
	if blockedIndex < 0 {
		t.Fatal("expected blocked section")
	}
	model.sectionIndex = blockedIndex
	model.sectionRows[calendarSectionBlocked] = 1
	model.selectedTaskID = tomorrowBlocked.ID

	model.moveSectionRow(-1)

	if model.sectionIndex != blockedIndex {
		t.Fatalf("expected blocked tab to stay selected, got section %d", model.sectionIndex)
	}
	if model.selectedTaskID != todayBlocked.ID {
		t.Fatalf("expected blocked selection to move to today task, got %s", model.selectedTaskID)
	}
}

func TestNowSelectedDayNavigationKeepsSelectedDayTab(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	today := time.Now().Local().Format("2006-01-02")
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open", DueOn: today})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open", DueOn: today})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, time.Now().Local(), 7, []shelf.Status{"open"}, calendarTUIOptions{
		Mode: calendarModeNow,
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	model.sectionIndex = 0
	initialTask, ok := model.selectedTask()
	if !ok {
		t.Fatalf("expected initial Selected Day task")
	}
	model.sectionRows[calendarSectionFocusedDay] = 0
	model.selectedTaskID = initialTask.ID

	model.moveSelectedDayTask(1)

	if model.sectionIndex != 0 {
		t.Fatalf("expected Selected Day tab to stay selected, got section %d", model.sectionIndex)
	}
	if model.selectedTaskID == initialTask.ID {
		t.Fatalf("expected selected day move to choose another task, still on %s", model.selectedTaskID)
	}
	validIDs := map[string]struct{}{
		first.ID:  {},
		second.ID: {},
	}
	if _, ok := validIDs[model.selectedTaskID]; !ok {
		t.Fatalf("unexpected selected day move target: %s", model.selectedTaskID)
	}
}

func TestBoardSelectionSyncsSidebarDate(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open", DueOn: "2026-03-10"})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, calendarTUIOptions{
		Mode: calendarModeBoard,
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	model.selectTaskByID(first.ID)
	model.moveBoardRow(1)
	if model.selectedTaskID != second.ID {
		t.Fatalf("expected board row move to select second task, got %s", model.selectedTaskID)
	}
	if model.focusedDayLabel() != "2026-03-10" {
		t.Fatalf("expected sidebar date to sync to 2026-03-10, got %s", model.focusedDayLabel())
	}
}

func TestTreeModeSelectedDayContentsRefreshWhenFocusedDateChanges(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	first, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "First", Kind: "todo", Status: "open", DueOn: "2026-03-09"})
	if err != nil {
		t.Fatalf("add first failed: %v", err)
	}
	second, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Second", Kind: "todo", Status: "open", DueOn: "2026-03-10"})
	if err != nil {
		t.Fatalf("add second failed: %v", err)
	}
	model, err := newCalendarTUIModelWithOptions(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open"}, calendarTUIOptions{
		Mode: calendarModeTree,
	})
	if err != nil {
		t.Fatalf("newCalendarTUIModelWithOptions failed: %v", err)
	}
	model.selectTaskByID(first.ID)
	model.pane = calendarPaneInspector
	model.moveSidebarFocusByDays(1)

	section := model.focusedDaySection()
	if section == nil || len(section.Items) != 1 {
		t.Fatalf("expected selected day section to rebuild for new date, got %+v", section)
	}
	if section.Items[0].Task.ID != second.ID {
		t.Fatalf("expected selected day contents to switch to second task, got %+v", section.Items)
	}
}

func TestCalendarTJumpToToday(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Today task",
		Kind:   "todo",
		Status: "open",
		DueOn:  time.Now().Local().Format("2006-01-02"),
	}); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, startOfWeek(time.Now().Local()).AddDate(0, 0, -7), 21, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	calendarModel := updatedModel.(calendarTUIModel)
	if calendarModel.focusedDayLabel() != time.Now().Local().Format("2006-01-02") {
		t.Fatalf("expected t to jump to today, got %s", calendarModel.focusedDayLabel())
	}
}

func TestCtrlHLCyclesModes(t *testing.T) {
	model := calendarTUIModel{
		mode:          calendarModeCalendar,
		sectionRows:   map[calendarSectionID]int{},
		boardRowIndex: map[int]int{},
		readiness:     map[string]shelf.TaskReadiness{},
		taskByID:      map[string]shelf.Task{},
		titleByID:     map[string]string{},
		outboundCount: map[string]int{},
		inboundCount:  map[string]int{},
		days:          []calendarDay{{Date: time.Now().Local().Format("2006-01-02")}},
	}
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	calendarModel := updatedModel.(calendarTUIModel)
	if calendarModel.mode != calendarModeTree {
		t.Fatalf("expected ctrl+l to move to next mode, got %s", calendarModel.mode)
	}
	updatedModel, _ = calendarModel.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	calendarModel = updatedModel.(calendarTUIModel)
	if calendarModel.mode != calendarModeCalendar {
		t.Fatalf("expected ctrl+h to move to previous mode, got %s", calendarModel.mode)
	}
}

func TestCalendarViewFitsWindowWidth(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Focused task",
		Kind:   "todo",
		Status: "open",
		DueOn:  "2026-03-09",
	}); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	model, err := newCalendarTUIModel(root, time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local), 7, []shelf.Status{"open", "in_progress", "blocked"}, false)
	if err != nil {
		t.Fatalf("newCalendarTUIModel failed: %v", err)
	}
	model.width = 120
	model.height = 24
	rendered := model.View()
	for _, line := range strings.Split(strings.TrimSuffix(rendered, "\n"), "\n") {
		if lipgloss.Width(line) > model.width {
			t.Fatalf("view line exceeds width %d: %d %q", model.width, lipgloss.Width(line), line)
		}
	}
}

func TestNowMainPaneShowsThreeSectionsAtOnce(t *testing.T) {
	today := time.Now().Local().Format("2006-01-02")
	model := calendarTUIModel{
		mode: calendarModeNow,
		days: []calendarDay{{Date: today}},
		sections: []calendarSection{
			{ID: calendarSectionFocusedDay, Title: "Selected Day", Items: []calendarSectionItem{{Task: shelf.Task{ID: "01A", Title: "Focus"}}}},
			{ID: calendarSectionOverdue, Title: "Overdue", Items: []calendarSectionItem{{Task: shelf.Task{ID: "01B", Title: "Late"}}}},
			{ID: calendarSectionToday, Title: "Today", Items: []calendarSectionItem{{Task: shelf.Task{ID: "01C", Title: "Today"}}}},
		},
		sectionRows:  map[calendarSectionID]int{},
		sectionIndex: 1,
	}
	month := calendarMonthView{Label: "March 2026"}
	rendered := renderCalendarMainPane(model, month, 120, 18, true)
	for _, want := range []string{"Selected Day 1", "Overdue 1", "Today 1"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("now pane should render all three sections, missing %q in %q", want, rendered)
		}
	}
	if strings.Contains(rendered, "Selected Day 1  Overdue 1") {
		t.Fatalf("now pane should render separate columns, got %q", rendered)
	}
}
