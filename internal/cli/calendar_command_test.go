package cli

import (
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
	rendered := renderCalendarCell(cell, "2026-03-09", 14)
	if got := lipgloss.Width(rendered); got != 14 {
		t.Fatalf("unexpected rendered width: %d", got)
	}
	if strings.Contains(rendered, "2026-03-09") {
		t.Fatalf("calendar cell should not render full date anymore: %q", rendered)
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

	sections := buildCalendarSections(calendarModeReview, focused, rootTasks, readiness, titles, 0)
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
	sections := buildCalendarSections(calendarModeToday, &calendarDay{Date: today, Tasks: []shelf.Task{tasks[2], tasks[3]}}, tasks, map[string]shelf.TaskReadiness{}, map[string]string{}, 1)
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
	rows := flattenCockpitTreeRows(nodes, "", true, false)
	if len(rows) != 2 {
		t.Fatalf("unexpected row count: %d", len(rows))
	}
	if !strings.Contains(rows[0].Label, "Parent") || rows[0].Meta != "todo/open" {
		t.Fatalf("unexpected parent row: %+v", rows[0])
	}
	if !strings.Contains(rows[1].Label, "Child") || !strings.Contains(rows[1].Label, "└─") || rows[1].Meta != "memo/blocked" {
		t.Fatalf("unexpected child row: %+v", rows[1])
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
	if task, ok := model.selectedTask(); !ok || task.ID != openTask.ID {
		t.Fatalf("unexpected initial selected task: %+v ok=%t", task, ok)
	}
	model.moveBoardColumn(1)
	if task, ok := model.selectedTask(); !ok || task.ID != doneTask.ID {
		t.Fatalf("unexpected selected task after moveBoardColumn: %+v ok=%t", task, ok)
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
			{ID: calendarSectionFocusedDay, Title: "Focused Day"},
			{ID: calendarSectionInbox, Title: "Inbox", Items: []calendarSectionItem{{Task: shelf.Task{ID: "01A", Title: "Inbox"}}}},
			{ID: calendarSectionToday, Title: "Today", Items: []calendarSectionItem{{Task: shelf.Task{ID: "01B", Title: "Today"}}}},
		},
		sectionRows: map[calendarSectionID]int{},
	}
	month := calendarMonthView{Label: "March 2026"}
	rendered := renderCalendarMainPane(model, month, 90, true)
	if strings.Contains(rendered, "March 2026") {
		t.Fatalf("review pane should not render month grid label: %q", rendered)
	}
	if !strings.Contains(rendered, "Focus ") || !strings.Contains(rendered, "Inbox 1") {
		t.Fatalf("review pane should render context strip: %q", rendered)
	}
}
