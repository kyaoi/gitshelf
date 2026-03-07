package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

type calendarMonthCell struct {
	Date           time.Time
	InCurrentMonth bool
	InRange        bool
	TaskCount      int
	DominantStatus shelf.Status
}

type calendarMonthView struct {
	Label string
	Weeks [][]calendarMonthCell
}

type calendarEditorFinishedMsg struct {
	Err error
}

type calendarTUIModel struct {
	rootDir      string
	startDate    time.Time
	daysCount    int
	statuses     []shelf.Status
	showID       bool
	days         []calendarDay
	dayIndex     int
	taskIndex    int
	width        int
	height       int
	message      string
	showTaskBody bool
	snoozeMode   bool
	snoozeIndex  int
}

func runCalendarTUI(rootDir string, startDate time.Time, daysCount int, statuses []shelf.Status, showID bool) error {
	model, err := newCalendarTUIModel(rootDir, startDate, daysCount, statuses, showID)
	if err != nil {
		return err
	}
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func newCalendarTUIModel(rootDir string, startDate time.Time, daysCount int, statuses []shelf.Status, showID bool) (calendarTUIModel, error) {
	model := calendarTUIModel{
		rootDir:   rootDir,
		startDate: startDate,
		daysCount: daysCount,
		statuses:  append([]shelf.Status{}, statuses...),
		showID:    showID,
		dayIndex:  0,
		taskIndex: 0,
	}
	if err := model.reload(); err != nil {
		return calendarTUIModel{}, err
	}
	return model, nil
}

func (m calendarTUIModel) Init() tea.Cmd {
	return nil
}

func (m calendarTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case calendarEditorFinishedMsg:
		if msg.Err != nil {
			m.message = msg.Err.Error()
			return m, nil
		}
		if err := m.reload(); err != nil {
			m.message = err.Error()
			return m, nil
		}
		m.message = "task updated"
		return m, nil
	case tea.KeyMsg:
		if m.snoozeMode {
			return m.updateSnoozeMode(msg)
		}
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "left", "h":
			m.dayIndex = max(0, m.dayIndex-1)
			m.clampTaskIndex()
			return m, nil
		case "right", "l":
			m.dayIndex = min(len(m.days)-1, m.dayIndex+1)
			m.clampTaskIndex()
			return m, nil
		case "up", "k":
			m.dayIndex = max(0, m.dayIndex-7)
			m.clampTaskIndex()
			return m, nil
		case "down", "j":
			m.dayIndex = min(len(m.days)-1, m.dayIndex+7)
			m.clampTaskIndex()
			return m, nil
		case "g":
			m.dayIndex = 0
			m.clampTaskIndex()
			return m, nil
		case "G":
			if len(m.days) > 0 {
				m.dayIndex = len(m.days) - 1
			}
			m.clampTaskIndex()
			return m, nil
		case "[", "H":
			m.dayIndex = moveCalendarIndexByMonth(m.days, m.dayIndex, -1)
			m.clampTaskIndex()
			return m, nil
		case "]", "L":
			m.dayIndex = moveCalendarIndexByMonth(m.days, m.dayIndex, 1)
			m.clampTaskIndex()
			return m, nil
		case "tab", "n":
			m.moveTaskSelection(1)
			return m, nil
		case "shift+tab", "p":
			m.moveTaskSelection(-1)
			return m, nil
		case "enter", "v":
			if _, ok := m.selectedTask(); ok {
				m.showTaskBody = !m.showTaskBody
			}
			return m, nil
		case "e":
			return m.openEditorForSelectedTask()
		case "z":
			if _, ok := m.selectedTask(); !ok {
				m.message = "選択中の日に task がありません"
				return m, nil
			}
			m.snoozeMode = true
			m.snoozeIndex = 0
			m.message = "期限変更プリセットを選択"
			return m, nil
		case "r":
			if err := m.reload(); err != nil {
				m.message = err.Error()
			} else {
				m.message = "reloaded"
			}
			return m, nil
		case "o":
			return m.applyStatusChange("open")
		case "i":
			return m.applyStatusChange("in_progress")
		case "b":
			return m.applyStatusChange("blocked")
		case "d":
			return m.applyStatusChange("done")
		case "c":
			return m.applyStatusChange("cancelled")
		}
	}
	return m, nil
}

func (m calendarTUIModel) View() string {
	if len(m.days) == 0 {
		return "No calendar days.\n"
	}

	focused := m.days[m.dayIndex]
	focusedDate, _ := time.ParseInLocation("2006-01-02", focused.Date, time.Now().Location())
	month := buildCalendarMonthView(m.days, focusedDate)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	subStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))

	header := []string{
		titleStyle.Render(fmt.Sprintf("Calendar %s .. %s", m.days[0].Date, m.days[len(m.days)-1].Date)),
		helpStyle.Render("h/l: 日  j/k: 週  [/]: 月  n/p: task  o/i/b/d/c: status  Enter: 詳細  e: edit  z: snooze  r: reload  q: quit"),
		subStyle.Render(fmt.Sprintf("Focused: %s", focusedDate.Format("Mon 2006-01-02"))),
	}

	parts := []string{
		strings.Join(header, "\n"),
		renderCalendarLegend(),
		renderCalendarMonth(month, focused.Date, m.width),
		renderCalendarDayDetails(focused, m.showID, m.taskIndex, m.showTaskBody),
	}
	if m.snoozeMode {
		parts = append(parts, renderCalendarSnoozePicker(m.snoozeIndex))
	}
	if strings.TrimSpace(m.message) != "" {
		parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Render(m.message))
	}

	return strings.Join(parts, "\n\n") + "\n"
}

func (m calendarTUIModel) updateSnoozeMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	options := calendarSnoozeOptions()
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		m.snoozeMode = false
		m.message = "期限変更をキャンセルしました"
		return m, nil
	case "up", "k":
		if m.snoozeIndex > 0 {
			m.snoozeIndex--
		}
		return m, nil
	case "down", "j":
		if m.snoozeIndex < len(options)-1 {
			m.snoozeIndex++
		}
		return m, nil
	case "enter":
		if len(options) == 0 {
			return m, nil
		}
		if err := m.applySnoozeOption(options[m.snoozeIndex]); err != nil {
			m.message = err.Error()
			m.snoozeMode = false
			return m, nil
		}
		m.snoozeMode = false
		return m, nil
	}
	return m, nil
}

func (m *calendarTUIModel) reload() error {
	tasks, err := shelf.ListTasks(m.rootDir, shelf.TaskFilter{
		Statuses: m.statuses,
		Limit:    0,
	})
	if err != nil {
		return err
	}
	m.days = buildCalendarDays(tasks, m.startDate, m.daysCount)
	if len(m.days) == 0 {
		m.dayIndex = 0
		m.taskIndex = 0
		return nil
	}
	if m.dayIndex >= len(m.days) {
		m.dayIndex = len(m.days) - 1
	}
	m.clampTaskIndex()
	return nil
}

func (m *calendarTUIModel) clampTaskIndex() {
	if len(m.days) == 0 {
		m.dayIndex = 0
		m.taskIndex = 0
		return
	}
	if m.dayIndex < 0 {
		m.dayIndex = 0
	}
	if m.dayIndex >= len(m.days) {
		m.dayIndex = len(m.days) - 1
	}
	tasks := m.days[m.dayIndex].Tasks
	if len(tasks) == 0 {
		m.taskIndex = 0
		m.showTaskBody = false
		return
	}
	if m.taskIndex < 0 {
		m.taskIndex = 0
	}
	if m.taskIndex >= len(tasks) {
		m.taskIndex = len(tasks) - 1
	}
}

func (m *calendarTUIModel) moveTaskSelection(delta int) {
	day := m.focusedDay()
	if day == nil || len(day.Tasks) == 0 {
		m.taskIndex = 0
		return
	}
	next := m.taskIndex + delta
	if next < 0 {
		next = len(day.Tasks) - 1
	}
	if next >= len(day.Tasks) {
		next = 0
	}
	m.taskIndex = next
}

func (m calendarTUIModel) focusedDay() *calendarDay {
	if len(m.days) == 0 || m.dayIndex < 0 || m.dayIndex >= len(m.days) {
		return nil
	}
	return &m.days[m.dayIndex]
}

func (m calendarTUIModel) selectedTask() (shelf.Task, bool) {
	day := m.focusedDay()
	if day == nil || len(day.Tasks) == 0 {
		return shelf.Task{}, false
	}
	if m.taskIndex < 0 || m.taskIndex >= len(day.Tasks) {
		return shelf.Task{}, false
	}
	return day.Tasks[m.taskIndex], true
}

func (m calendarTUIModel) openEditorForSelectedTask() (tea.Model, tea.Cmd) {
	task, ok := m.selectedTask()
	if !ok {
		m.message = "選択中の日に task がありません"
		return m, nil
	}
	editorCmd, err := resolveEditorCommand(os.LookupEnv)
	if err != nil {
		m.message = err.Error()
		return m, nil
	}
	args := strings.Fields(strings.TrimSpace(editorCmd))
	if len(args) == 0 {
		m.message = "editor command is empty"
		return m, nil
	}
	taskPath := filepath.Join(shelf.TasksDir(m.rootDir), task.ID+".md")
	args = append(args, taskPath)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return calendarEditorFinishedMsg{Err: normalizeEditorExecError(err)}
		}
		return calendarEditorFinishedMsg{}
	})
}

func (m *calendarTUIModel) applySnoozeOption(option snoozePreset) error {
	task, ok := m.selectedTask()
	if !ok {
		return fmt.Errorf("選択中の日に task がありません")
	}

	var (
		nextDue string
		err     error
	)
	if option.Mode == snoozeModeTo {
		nextDue, err = shelf.NormalizeDueOn(option.Value)
	} else {
		nextDue, err = applyByDays(task.DueOn, option.Value)
	}
	if err != nil {
		return err
	}

	if err := withWriteLock(m.rootDir, func() error {
		if err := prepareUndoSnapshot(m.rootDir, "calendar-snooze"); err != nil {
			return err
		}
		_, err := shelf.SetTask(m.rootDir, task.ID, shelf.SetTaskInput{DueOn: &nextDue})
		return err
	}); err != nil {
		return err
	}
	if err := m.reload(); err != nil {
		return err
	}
	m.message = fmt.Sprintf("Snoozed %s to %s", task.Title, nextDue)
	return nil
}

func (m calendarTUIModel) applyStatusChange(nextStatus shelf.Status) (tea.Model, tea.Cmd) {
	task, ok := m.selectedTask()
	if !ok {
		m.message = "no task selected"
		return m, nil
	}
	err := withWriteLock(m.rootDir, func() error {
		if err := prepareUndoSnapshot(m.rootDir, "calendar-status"); err != nil {
			return err
		}
		_, err := shelf.SetTask(m.rootDir, task.ID, shelf.SetTaskInput{Status: &nextStatus})
		return err
	})
	if err != nil {
		m.message = err.Error()
		return m, nil
	}
	if err := m.reload(); err != nil {
		m.message = err.Error()
		return m, nil
	}
	m.message = fmt.Sprintf("%s -> %s", task.Title, nextStatus)
	return m, nil
}

func buildCalendarMonthView(days []calendarDay, focusDate time.Time) calendarMonthView {
	taskCounts := map[string]int{}
	inRange := map[string]struct{}{}
	dominant := map[string]shelf.Status{}
	for _, day := range days {
		taskCounts[day.Date] = len(day.Tasks)
		inRange[day.Date] = struct{}{}
		dominant[day.Date] = dominantCalendarStatus(day.Tasks)
	}

	monthStart := time.Date(focusDate.Year(), focusDate.Month(), 1, 0, 0, 0, 0, focusDate.Location())
	monthEnd := monthStart.AddDate(0, 1, -1)
	gridStart := startOfWeek(monthStart)
	gridEnd := startOfWeek(monthEnd).AddDate(0, 0, 6)

	weeks := make([][]calendarMonthCell, 0, 6)
	current := gridStart
	for !current.After(gridEnd) {
		week := make([]calendarMonthCell, 0, 7)
		for i := 0; i < 7; i++ {
			key := current.Format("2006-01-02")
			_, ok := inRange[key]
			week = append(week, calendarMonthCell{
				Date:           current,
				InCurrentMonth: current.Month() == focusDate.Month(),
				InRange:        ok,
				TaskCount:      taskCounts[key],
				DominantStatus: dominant[key],
			})
			current = current.AddDate(0, 0, 1)
		}
		weeks = append(weeks, week)
	}

	return calendarMonthView{
		Label: focusDate.Format("January 2006"),
		Weeks: weeks,
	}
}

func renderCalendarLegend() string {
	item := func(color lipgloss.Color, label string) string {
		return lipgloss.NewStyle().Foreground(color).Render("■ " + label)
	}
	return strings.Join([]string{
		item(lipgloss.Color("203"), "blocked"),
		item(lipgloss.Color("220"), "in_progress"),
		item(lipgloss.Color("81"), "open"),
		item(lipgloss.Color("78"), "done"),
		item(lipgloss.Color("245"), "cancelled"),
	}, "  ")
}

func renderCalendarMonth(month calendarMonthView, focusedDate string, width int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	dayHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244"))
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 1)

	cellWidth := 14
	if width > 0 {
		cellWidth = max(10, min(16, (width-10)/7))
	}

	headers := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	headerCells := make([]string, 0, len(headers))
	for _, header := range headers {
		headerCells = append(headerCells, dayHeaderStyle.Width(cellWidth).Align(lipgloss.Center).Render(header))
	}

	rows := []string{
		titleStyle.Render(month.Label),
		lipgloss.JoinHorizontal(lipgloss.Top, headerCells...),
	}
	for _, week := range month.Weeks {
		cells := make([]string, 0, len(week))
		for _, cell := range week {
			cells = append(cells, renderCalendarCell(cell, focusedDate, cellWidth))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}

	return containerStyle.Render(strings.Join(rows, "\n"))
}

func renderCalendarCell(cell calendarMonthCell, focusedDate string, cellWidth int) string {
	key := cell.Date.Format("2006-01-02")
	today := time.Now().Format("2006-01-02")

	contentWidth := max(6, cellWidth)
	style := lipgloss.NewStyle().
		Width(contentWidth).
		Height(3)

	switch {
	case key == focusedDate:
		style = style.Background(lipgloss.Color("25")).Foreground(lipgloss.Color("255")).Bold(true)
	case cell.InRange && cell.TaskCount > 0:
		style = style.Background(calendarStatusColor(cell.DominantStatus)).Foreground(lipgloss.Color("255"))
	case cell.InRange:
		style = style.Background(lipgloss.Color("235")).Foreground(lipgloss.Color("250"))
	case cell.InCurrentMonth:
		style = style.Foreground(lipgloss.Color("240"))
	default:
		style = style.Foreground(lipgloss.Color("238"))
	}

	dayLabel := fmt.Sprintf("%2d", cell.Date.Day())
	if key == today {
		dayLabel += " •"
	}

	countLine := ""
	if cell.TaskCount > 0 {
		if cell.TaskCount == 1 {
			countLine = "1 due"
		} else {
			countLine = fmt.Sprintf("%d due", cell.TaskCount)
		}
	}

	lines := []string{
		padOrTrim(dayLabel, contentWidth),
		padOrTrim(countLine, contentWidth),
		padOrTrim(cell.Date.Format("2006-01-02"), contentWidth),
	}
	return style.Render(strings.Join(lines, "\n"))
}

func renderCalendarDayDetails(day calendarDay, showID bool, selectedIndex int, showTaskBody bool) string {
	listStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)

	lines := []string{
		titleStyle.Render(fmt.Sprintf("%s (%d due)", day.Date, len(day.Tasks))),
	}
	if len(day.Tasks) == 0 {
		lines = append(lines, mutedStyle.Render("(no due tasks)"))
		return listStyle.Render(strings.Join(lines, "\n"))
	}

	for i, task := range day.Tasks {
		label := task.Title
		if showID {
			label = fmt.Sprintf("[%s] %s", shelf.ShortID(task.ID), label)
		}
		line := fmt.Sprintf("  %s (%s/%s)", label, uiKind(task.Kind), uiStatus(task.Status))
		if i == selectedIndex {
			line = selectedStyle.Render("> " + line)
		} else {
			line = "  " + line
		}
		lines = append(lines, line)
	}

	parts := []string{listStyle.Render(strings.Join(lines, "\n"))}
	if selectedIndex >= 0 && selectedIndex < len(day.Tasks) {
		parts = append(parts, renderCalendarTaskPreview(day.Tasks[selectedIndex], showID, showTaskBody))
	}
	return strings.Join(parts, "\n\n")
}

func renderCalendarTaskPreview(task shelf.Task, showID bool, showTaskBody bool) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	label := task.Title
	if showID {
		label = fmt.Sprintf("[%s] %s", shelf.ShortID(task.ID), label)
	}

	lines := []string{
		titleStyle.Render(label),
		fmt.Sprintf("kind=%s  status=%s  due=%s", uiKind(task.Kind), uiStatus(task.Status), uiDue(task.DueOn)),
	}
	if len(task.Tags) > 0 {
		lines = append(lines, fmt.Sprintf("tags=%s", strings.Join(task.Tags, ", ")))
	}
	if strings.TrimSpace(task.Parent) != "" {
		lines = append(lines, fmt.Sprintf("parent=%s", task.Parent))
	}

	body := strings.TrimSpace(task.Body)
	if body == "" {
		lines = append(lines, mutedStyle.Render("(empty body)"))
		return boxStyle.Render(strings.Join(lines, "\n"))
	}

	bodyLines := strings.Split(body, "\n")
	maxLines := 3
	if showTaskBody {
		maxLines = 12
		lines = append(lines, mutedStyle.Render("full body preview"))
	} else {
		lines = append(lines, mutedStyle.Render("compact body preview"))
	}
	if len(bodyLines) > maxLines {
		bodyLines = append(bodyLines[:maxLines], "...")
	}
	lines = append(lines, strings.Join(bodyLines, "\n"))
	return boxStyle.Render(strings.Join(lines, "\n"))
}

func renderCalendarSnoozePicker(selected int) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("141")).
		Padding(0, 1)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	lines := []string{
		titleStyle.Render("Snooze Presets"),
		helpStyle.Render("j/k: 移動  Enter: 決定  Esc/q: 戻る"),
	}
	for i, option := range calendarSnoozeOptions() {
		line := "  " + option.Label
		if i == selected {
			line = selectedStyle.Render("> " + option.Label)
		}
		lines = append(lines, line)
	}
	return boxStyle.Render(strings.Join(lines, "\n"))
}

func calendarSnoozeOptions() []snoozePreset {
	options := make([]snoozePreset, 0, len(snoozeInteractivePresets()))
	for _, option := range snoozeInteractivePresets() {
		if option.NeedsInput {
			continue
		}
		options = append(options, option)
	}
	return options
}

func dominantCalendarStatus(tasks []shelf.Task) shelf.Status {
	best := shelf.Status("")
	bestScore := -1
	for _, task := range tasks {
		score := calendarStatusPriority(task.Status)
		if score > bestScore {
			best = task.Status
			bestScore = score
		}
	}
	return best
}

func calendarStatusPriority(status shelf.Status) int {
	switch status {
	case "blocked":
		return 5
	case "in_progress":
		return 4
	case "open":
		return 3
	case "done":
		return 2
	case "cancelled":
		return 1
	default:
		return 0
	}
}

func calendarStatusColor(status shelf.Status) lipgloss.Color {
	switch status {
	case "blocked":
		return lipgloss.Color("160")
	case "in_progress":
		return lipgloss.Color("172")
	case "open":
		return lipgloss.Color("24")
	case "done":
		return lipgloss.Color("28")
	case "cancelled":
		return lipgloss.Color("238")
	default:
		return lipgloss.Color("237")
	}
}

func moveCalendarIndexByMonth(days []calendarDay, currentIndex int, delta int) int {
	if len(days) == 0 {
		return 0
	}
	if currentIndex < 0 {
		currentIndex = 0
	}
	if currentIndex >= len(days) {
		currentIndex = len(days) - 1
	}
	currentDate, err := time.ParseInLocation("2006-01-02", days[currentIndex].Date, time.Now().Location())
	if err != nil {
		return currentIndex
	}
	target := currentDate.AddDate(0, delta, 0)
	startDate, err := time.ParseInLocation("2006-01-02", days[0].Date, time.Now().Location())
	if err != nil {
		return currentIndex
	}
	endDate, err := time.ParseInLocation("2006-01-02", days[len(days)-1].Date, time.Now().Location())
	if err != nil {
		return currentIndex
	}
	if target.Before(startDate) {
		return 0
	}
	if target.After(endDate) {
		return len(days) - 1
	}
	diff := int(target.Sub(startDate).Hours() / 24)
	if diff < 0 {
		return 0
	}
	if diff >= len(days) {
		return len(days) - 1
	}
	return diff
}

func padOrTrim(value string, width int) string {
	value = trimLine(value, width)
	return lipgloss.NewStyle().Width(width).Render(value)
}

func trimLine(value string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(value)
	if lipgloss.Width(value) <= width || len(runes) <= width {
		return value
	}
	if width <= 1 {
		return string(runes[:1])
	}
	if len(runes) <= width {
		return value
	}
	return string(runes[:width-1]) + "…"
}
