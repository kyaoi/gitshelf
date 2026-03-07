package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

type calendarMode string

const (
	calendarModeCalendar calendarMode = "calendar"
	calendarModeReview   calendarMode = "review"
	calendarModeNow      calendarMode = "now"
	calendarModeTree     calendarMode = "tree"
	calendarModeBoard    calendarMode = "board"
)

type calendarPane int

const (
	calendarPaneMain calendarPane = iota
	calendarPaneInspector
)

type calendarSectionID string

const (
	calendarSectionFocusedDay calendarSectionID = "focused_day"
	calendarSectionInbox      calendarSectionID = "inbox"
	calendarSectionOverdue    calendarSectionID = "overdue"
	calendarSectionToday      calendarSectionID = "today"
	calendarSectionBlocked    calendarSectionID = "blocked"
	calendarSectionReady      calendarSectionID = "ready"
)

type calendarTUIOptions struct {
	Mode         calendarMode
	ShowID       bool
	SectionLimit int
	Filter       shelf.TaskFilter
}

type calendarSectionItem struct {
	Task     shelf.Task
	Subtitle string
	Reason   string
}

type calendarSection struct {
	ID    calendarSectionID
	Title string
	Items []calendarSectionItem
}

type cockpitTreeRow struct {
	Task  shelf.Task
	Label string
	Meta  string
}

type calendarTUIModel struct {
	rootDir       string
	mode          calendarMode
	startDate     time.Time
	daysCount     int
	statuses      []shelf.Status
	sectionLimit  int
	filter        shelf.TaskFilter
	defaultKind   shelf.Kind
	defaultStatus shelf.Status
	showID        bool

	visibleTasks  []shelf.Task
	allTasks      []shelf.Task
	readiness     map[string]shelf.TaskReadiness
	taskByID      map[string]shelf.Task
	titleByID     map[string]string
	outboundCount map[string]int
	inboundCount  map[string]int

	days           []calendarDay
	dayIndex       int
	sections       []calendarSection
	sectionIndex   int
	sectionRows    map[calendarSectionID]int
	treeRows       []cockpitTreeRow
	treeRowIndex   int
	boardColumns   []boardColumn
	boardColumnIdx int
	boardRowIndex  map[int]int
	selectedTaskID string
	pane           calendarPane
	width          int
	height         int
	bodyScroll     int
	message        string
	showHelp       bool
	showTaskBody   bool
	snoozeMode     bool
	snoozeIndex    int
	addMode        bool
	addTitle       string
}

func runCalendarTUI(rootDir string, startDate time.Time, daysCount int, statuses []shelf.Status, showID bool) error {
	return runCalendarModeTUI(rootDir, startDate, daysCount, statuses, calendarTUIOptions{
		Mode:   calendarModeCalendar,
		ShowID: showID,
	})
}

func runCalendarModeTUI(rootDir string, startDate time.Time, daysCount int, statuses []shelf.Status, opts calendarTUIOptions) error {
	model, err := newCalendarTUIModelWithOptions(rootDir, startDate, daysCount, statuses, opts)
	if err != nil {
		return err
	}
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func newCalendarTUIModel(rootDir string, startDate time.Time, daysCount int, statuses []shelf.Status, showID bool) (calendarTUIModel, error) {
	return newCalendarTUIModelWithOptions(rootDir, startDate, daysCount, statuses, calendarTUIOptions{
		Mode:   calendarModeCalendar,
		ShowID: showID,
	})
}

func newCalendarTUIModelWithOptions(rootDir string, startDate time.Time, daysCount int, statuses []shelf.Status, opts calendarTUIOptions) (calendarTUIModel, error) {
	cfg, err := shelf.LoadConfig(rootDir)
	if err != nil {
		return calendarTUIModel{}, err
	}
	if opts.Mode == "" {
		opts.Mode = calendarModeCalendar
	}
	model := calendarTUIModel{
		rootDir:       rootDir,
		mode:          opts.Mode,
		startDate:     startDate,
		daysCount:     daysCount,
		sectionLimit:  opts.SectionLimit,
		defaultKind:   cfg.DefaultKind,
		defaultStatus: cfg.DefaultStatus,
		showID:        opts.ShowID,
		pane:          calendarPaneMain,
		sectionRows:   map[calendarSectionID]int{},
		boardRowIndex: map[int]int{},
		readiness:     map[string]shelf.TaskReadiness{},
		taskByID:      map[string]shelf.Task{},
		titleByID:     map[string]string{},
		outboundCount: map[string]int{},
		inboundCount:  map[string]int{},
	}
	model.filter = opts.Filter
	model.filter.Limit = 0
	if len(model.filter.Statuses) == 0 {
		model.filter.Statuses = append([]shelf.Status{}, statuses...)
	}
	model.statuses = append([]shelf.Status{}, model.filter.Statuses...)
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
		if m.addMode {
			return m.updateAddMode(msg)
		}
		if m.snoozeMode {
			return m.updateSnoozeMode(msg)
		}
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "tab":
			m.nextPane()
			return m, nil
		case "shift+tab":
			m.prevPane()
			return m, nil
		case "pgdown", "ctrl+d":
			m.scrollBody(m.bodyPageStep())
			return m, nil
		case "pgup", "ctrl+u":
			m.scrollBody(-m.bodyPageStep())
			return m, nil
		case "home":
			m.bodyScroll = 0
			return m, nil
		case "end":
			m.bodyScroll = 1 << 30
			return m, nil
		case "left", "h":
			if m.usesSidebarCalendarNav() {
				m.moveFocusByDays(-1)
			} else if m.mode == calendarModeCalendar {
				m.moveFocusByDays(-1)
			} else if m.mode == calendarModeBoard {
				m.moveBoardColumn(-1)
			} else {
				m.moveSectionSelection(-1)
			}
			return m, nil
		case "right", "l":
			if m.usesSidebarCalendarNav() {
				m.moveFocusByDays(1)
			} else if m.mode == calendarModeCalendar {
				m.moveFocusByDays(1)
			} else if m.mode == calendarModeBoard {
				m.moveBoardColumn(1)
			} else {
				m.moveSectionSelection(1)
			}
			return m, nil
		case "up":
			if m.usesSidebarCalendarNav() {
				m.moveFocusByDays(-7)
				return m, nil
			}
			if m.mode == calendarModeBoard {
				m.moveBoardRow(-1)
				return m, nil
			}
			m.moveSectionRow(-1)
			return m, nil
		case "down":
			if m.usesSidebarCalendarNav() {
				m.moveFocusByDays(7)
				return m, nil
			}
			if m.mode == calendarModeBoard {
				m.moveBoardRow(1)
				return m, nil
			}
			m.moveSectionRow(1)
			return m, nil
		case "k":
			if m.usesSidebarCalendarNav() {
				m.moveFocusByDays(-7)
			} else if m.mode == calendarModeCalendar {
				m.moveFocusByDays(-7)
			} else if m.mode == calendarModeBoard {
				m.moveBoardRow(-1)
			} else {
				m.moveSectionRow(-1)
			}
			return m, nil
		case "j":
			if m.usesSidebarCalendarNav() {
				m.moveFocusByDays(7)
			} else if m.mode == calendarModeCalendar {
				m.moveFocusByDays(7)
			} else if m.mode == calendarModeBoard {
				m.moveBoardRow(1)
			} else {
				m.moveSectionRow(1)
			}
			return m, nil
		case "g":
			if m.mode == calendarModeCalendar {
				m.dayIndex = 0
				m.rebuildSections()
			} else if m.mode == calendarModeBoard {
				m.jumpBoardRowStart()
			} else {
				m.jumpSectionRowStart()
			}
			return m, nil
		case "G":
			if m.mode == calendarModeCalendar {
				if len(m.days) > 0 {
					m.dayIndex = len(m.days) - 1
					m.rebuildSections()
				}
			} else if m.mode == calendarModeBoard {
				m.jumpBoardRowEnd()
			} else {
				m.jumpSectionRowEnd()
			}
			return m, nil
		case "[", "H":
			if m.mode == calendarModeCalendar || m.usesSidebarCalendarNav() {
				m.moveFocusByMonths(-1)
			}
			return m, nil
		case "]", "L":
			if m.mode == calendarModeCalendar || m.usesSidebarCalendarNav() {
				m.moveFocusByMonths(1)
			}
			return m, nil
		case "n":
			if m.mode == calendarModeBoard {
				m.moveBoardColumn(1)
			} else if m.mode == calendarModeCalendar {
				m.moveFocusedDayTask(1)
			} else {
				m.moveSectionSelection(1)
			}
			return m, nil
		case "p":
			if m.mode == calendarModeBoard {
				m.moveBoardColumn(-1)
			} else if m.mode == calendarModeCalendar {
				m.moveFocusedDayTask(-1)
			} else {
				m.moveSectionSelection(-1)
			}
			return m, nil
		case "1", "2", "3", "4", "5", "6":
			m.jumpToSection(int(msg.String()[0] - '1'))
			return m, nil
		case "C":
			m.switchMode(calendarModeCalendar)
			return m, nil
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "T":
			m.switchMode(calendarModeTree)
			return m, nil
		case "B":
			m.switchMode(calendarModeBoard)
			return m, nil
		case "R":
			m.switchMode(calendarModeReview)
			return m, nil
		case "N":
			m.switchMode(calendarModeNow)
			return m, nil
		case "enter", "v":
			if _, ok := m.selectedTask(); ok {
				m.showTaskBody = !m.showTaskBody
			}
			return m, nil
		case "e":
			return m.openEditorForSelectedTask()
		case "a":
			m.addMode = true
			m.addTitle = ""
			m.message = "新規 task title を入力"
			return m, nil
		case "z":
			if _, ok := m.selectedTask(); !ok {
				m.message = "選択中の task がありません"
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

	focused, _ := m.focusedDate()
	month := buildCalendarMonthView(m.days, focused)
	mainWidth, gapWidth, inspectorWidth := m.layoutColumns()

	main := renderCalendarMainPane(m, month, mainWidth, m.pane == calendarPaneMain)
	selectedTask, selectedTaskOK := m.selectedTask()
	right := renderCalendarInspectorPane(selectedTask, selectedTaskOK, m.showID, m.showTaskBody, m.taskByID, m.readiness, m.outboundCount, m.inboundCount, inspectorWidth)
	if m.mode == calendarModeCalendar {
		right = renderCalendarSidebarPane(m, selectedTask, selectedTaskOK, inspectorWidth)
	} else {
		right = renderCalendarSecondarySidebarPane(m, selectedTask, selectedTaskOK, inspectorWidth, m.pane == calendarPaneInspector)
	}

	gap := lipgloss.NewStyle().Width(gapWidth).Render("")
	body := lipgloss.JoinHorizontal(lipgloss.Top, main, gap, right)
	topParts := []string{
		renderCockpitHeader(m, focused),
		renderCalendarModeTabs(m.mode),
	}
	bottomParts := make([]string, 0, 4)
	if m.showHelp {
		bottomParts = append(bottomParts, renderCockpitHelpOverlay(m.mode))
	}
	if m.snoozeMode {
		bottomParts = append(bottomParts, renderCalendarSnoozePicker(m.snoozeIndex))
	}
	if m.addMode {
		bottomParts = append(bottomParts, renderCalendarAddComposer(m.focusedDayLabel(), m.defaultKind, m.defaultStatus, m.addTitle))
	}
	if strings.TrimSpace(m.message) != "" {
		bottomParts = append(bottomParts, lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Render(m.message))
	}
	return renderCalendarViewport(topParts, body, bottomParts, m.height, m.bodyScroll) + "\n"
}

func (m calendarTUIModel) layoutWidths() (int, int) {
	main, _, inspector := m.layoutColumns()
	return main, inspector
}

func (m calendarTUIModel) layoutColumns() (int, int, int) {
	if m.width <= 0 {
		if m.mode == calendarModeCalendar {
			return 91, 1, 48
		}
		return 88, 1, 54
	}
	usable := max(96, m.width-2)
	if m.mode == calendarModeCalendar {
		gap := max(1, usable/100)
		main := usable * 65 / 100
		inspector := usable - main - gap
		if main < 56 {
			main = 56
			inspector = usable - main - gap
		}
		if inspector < 36 {
			inspector = 36
			main = usable - inspector - gap
		}
		return main, gap, inspector
	}
	inspectorRatio := 34
	inspector := max(36, usable*inspectorRatio/100)
	gap := 1
	main := usable - inspector - gap
	if main < 56 {
		main = 56
		inspector = usable - main - gap
	}
	return main, gap, inspector
}

func (m calendarTUIModel) usesSidebarCalendarNav() bool {
	return m.mode != calendarModeCalendar && m.pane == calendarPaneInspector
}

func (m calendarTUIModel) bodyPageStep() int {
	if m.height <= 0 {
		return 8
	}
	return max(3, m.height/3)
}

func (m *calendarTUIModel) scrollBody(delta int) {
	m.bodyScroll += delta
	if m.bodyScroll < 0 {
		m.bodyScroll = 0
	}
}

func (m *calendarTUIModel) nextPane() {
	m.pane = (m.pane + 1) % 2
}

func (m *calendarTUIModel) prevPane() {
	m.pane--
	if m.pane < 0 {
		m.pane = calendarPaneInspector
	}
}

func (m calendarTUIModel) paneLabel() string {
	if m.pane == calendarPaneInspector {
		return "inspector"
	}
	return "main"
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

func (m calendarTUIModel) updateAddMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.addMode = false
		m.addTitle = ""
		m.message = "新規 task 作成をキャンセルしました"
		return m, nil
	case "enter":
		title := strings.TrimSpace(m.addTitle)
		if title == "" {
			m.message = "title は必須です"
			return m, nil
		}
		if err := m.createTaskOnFocusedDay(title); err != nil {
			m.message = err.Error()
			return m, nil
		}
		m.addMode = false
		m.addTitle = ""
		return m, nil
	case "backspace":
		if len(m.addTitle) > 0 {
			runes := []rune(m.addTitle)
			m.addTitle = string(runes[:len(runes)-1])
		}
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			m.addTitle += msg.String()
			return m, nil
		}
	}
	return m, nil
}

func (m *calendarTUIModel) reload() error {
	selectedTaskID := m.selectedTaskID
	filter := m.filter
	filter.Limit = 0
	visibleTasks, err := shelf.ListTasks(m.rootDir, filter)
	if err != nil {
		return err
	}
	allTasks, err := shelf.NewTaskStore(m.rootDir).List()
	if err != nil {
		return err
	}
	readiness, err := shelf.BuildTaskReadiness(m.rootDir)
	if err != nil {
		return err
	}
	outboundCount, inboundCount, err := buildCalendarLinkCounts(m.rootDir, allTasks)
	if err != nil {
		return err
	}

	m.visibleTasks = visibleTasks
	m.allTasks = allTasks
	m.readiness = readiness
	m.outboundCount = outboundCount
	m.inboundCount = inboundCount
	m.taskByID = make(map[string]shelf.Task, len(allTasks))
	m.titleByID = make(map[string]string, len(allTasks))
	for _, task := range allTasks {
		m.taskByID[task.ID] = task
		m.titleByID[task.ID] = task.Title
	}

	m.days = buildCalendarDays(visibleTasks, m.startDate, m.daysCount)
	if len(m.days) == 0 {
		m.dayIndex = 0
		m.sections = nil
		m.treeRows = nil
		m.selectedTaskID = ""
		return nil
	}
	if m.dayIndex >= len(m.days) {
		m.dayIndex = len(m.days) - 1
	}
	if m.dayIndex < 0 {
		m.dayIndex = 0
	}
	m.selectedTaskID = selectedTaskID
	m.rebuildModeState()
	return nil
}

func (m *calendarTUIModel) rebuildModeState() {
	if m.mode == calendarModeTree {
		m.rebuildTreeRows()
		return
	}
	if m.mode == calendarModeBoard {
		m.rebuildBoardColumns()
		return
	}
	m.rebuildSections()
}

func (m *calendarTUIModel) switchMode(mode calendarMode) {
	if m.mode == mode {
		return
	}
	m.mode = mode
	m.sectionIndex = 0
	m.treeRowIndex = 0
	m.bodyScroll = 0
	m.rebuildModeState()
	m.message = fmt.Sprintf("mode: %s", mode)
}

func (m *calendarTUIModel) rebuildBoardColumns() {
	m.boardColumns = buildBoardColumns(m.statuses, m.visibleTasks)
	for i := range m.boardColumns {
		if _, ok := m.boardRowIndex[i]; !ok {
			m.boardRowIndex[i] = 0
		}
	}
	if len(m.boardColumns) == 0 {
		m.boardColumnIdx = 0
		m.selectedTaskID = ""
		return
	}
	if m.selectedTaskID != "" {
		for colIdx, column := range m.boardColumns {
			for rowIdx, task := range column.Tasks {
				if task.ID == m.selectedTaskID {
					m.boardColumnIdx = colIdx
					m.boardRowIndex[colIdx] = rowIdx
					return
				}
			}
		}
	}
	m.clampBoardSelection()
}

func buildCalendarLinkCounts(rootDir string, tasks []shelf.Task) (map[string]int, map[string]int, error) {
	edgeStore := shelf.NewEdgeStore(rootDir)
	outboundCount := make(map[string]int, len(tasks))
	inboundCount := make(map[string]int, len(tasks))
	for _, task := range tasks {
		outbound, err := edgeStore.ListOutbound(task.ID)
		if err != nil {
			return nil, nil, err
		}
		outboundCount[task.ID] = len(outbound)
		for _, edge := range outbound {
			inboundCount[edge.To]++
		}
	}
	return outboundCount, inboundCount, nil
}

func (m *calendarTUIModel) rebuildSections() {
	prevSectionID := calendarSectionFocusedDay
	if section := m.currentSection(); section != nil {
		prevSectionID = section.ID
	}
	m.sections = buildCalendarSections(m.mode, m.focusedDay(), m.visibleTasks, m.readiness, m.titleByID, m.sectionLimit)
	if len(m.sections) == 0 {
		m.sectionIndex = 0
		m.selectedTaskID = ""
		return
	}
	if m.selectedTaskID != "" {
		if secIdx, rowIdx, ok := findCalendarSectionTask(m.sections, m.selectedTaskID); ok {
			m.sectionIndex = secIdx
			m.sectionRows[m.sections[secIdx].ID] = rowIdx
			return
		}
	}
	for i, section := range m.sections {
		if section.ID == prevSectionID {
			m.sectionIndex = i
			break
		}
	}
	m.clampSectionSelection()
	if task, ok := m.selectedTask(); ok {
		m.selectedTaskID = task.ID
		return
	}
	m.selectFirstAvailableTask()
}

func (m *calendarTUIModel) rebuildTreeRows() {
	opts, err := treeOptionsFromFilter(shelf.TreeOptions{}, m.filter)
	if err != nil {
		m.treeRows = nil
		m.message = err.Error()
		m.selectedTaskID = ""
		return
	}
	nodes, err := shelf.BuildTree(m.rootDir, opts)
	if err != nil {
		m.treeRows = nil
		m.message = err.Error()
		m.selectedTaskID = ""
		return
	}
	m.treeRows = flattenCockpitTreeRows(nodes, "", true, m.showID)
	if len(m.treeRows) == 0 {
		m.treeRowIndex = 0
		m.selectedTaskID = ""
		return
	}
	if m.selectedTaskID != "" {
		for i, row := range m.treeRows {
			if row.Task.ID == m.selectedTaskID {
				m.treeRowIndex = i
				m.selectedTaskID = row.Task.ID
				return
			}
		}
	}
	if m.treeRowIndex < 0 {
		m.treeRowIndex = 0
	}
	if m.treeRowIndex >= len(m.treeRows) {
		m.treeRowIndex = len(m.treeRows) - 1
	}
	m.selectedTaskID = m.treeRows[m.treeRowIndex].Task.ID
}

func flattenCockpitTreeRows(nodes []shelf.TreeNode, prefix string, isRoot bool, showID bool) []cockpitTreeRow {
	rows := make([]cockpitTreeRow, 0)
	for i, node := range nodes {
		isLast := i == len(nodes)-1
		branch := "├─ "
		nextPrefix := prefix + "│  "
		if isLast {
			branch = "└─ "
			nextPrefix = prefix + "   "
		}
		if isRoot {
			branch = ""
		}
		label := node.Task.Title
		if showID {
			label = fmt.Sprintf("[%s] %s", shelf.ShortID(node.Task.ID), label)
		}
		meta := fmt.Sprintf("%s/%s", node.Task.Kind, node.Task.Status)
		if strings.TrimSpace(node.Task.DueOn) != "" {
			meta += "  due=" + node.Task.DueOn
		}
		label = fmt.Sprintf("%s%s%s", prefix, branch, label)
		rows = append(rows, cockpitTreeRow{
			Task:  node.Task,
			Label: label,
			Meta:  meta,
		})
		rows = append(rows, flattenCockpitTreeRows(node.Children, nextPrefix, false, showID)...)
	}
	return rows
}

func buildCalendarSections(mode calendarMode, focusedDay *calendarDay, tasks []shelf.Task, readiness map[string]shelf.TaskReadiness, titleByID map[string]string, sectionLimit int) []calendarSection {
	sections := make([]calendarSection, 0, 6)
	for _, descriptor := range calendarSectionsForMode(mode) {
		var items []calendarSectionItem
		switch descriptor.ID {
		case calendarSectionFocusedDay:
			items = buildFocusedDaySectionItems(focusedDay, titleByID)
		case calendarSectionInbox:
			items = buildInboxSectionItems(tasks, titleByID)
		case calendarSectionOverdue:
			items = buildOverdueSectionItems(tasks, titleByID)
		case calendarSectionToday:
			items = buildTodaySectionItems(tasks, titleByID)
		case calendarSectionBlocked:
			items = buildBlockedSectionItems(tasks, readiness, titleByID)
		case calendarSectionReady:
			items = buildReadySectionItems(tasks, readiness, titleByID)
		}
		if sectionLimit > 0 && descriptor.ID != calendarSectionFocusedDay && len(items) > sectionLimit {
			items = items[:sectionLimit]
		}
		sections = append(sections, calendarSection{ID: descriptor.ID, Title: descriptor.Title, Items: items})
	}
	return sections
}

func calendarSectionsForMode(mode calendarMode) []calendarSection {
	switch mode {
	case calendarModeReview:
		return []calendarSection{{ID: calendarSectionFocusedDay, Title: "Focused Day"}, {ID: calendarSectionInbox, Title: "Inbox"}, {ID: calendarSectionOverdue, Title: "Overdue"}, {ID: calendarSectionToday, Title: "Today"}, {ID: calendarSectionBlocked, Title: "Blocked"}, {ID: calendarSectionReady, Title: "Ready"}}
	case calendarModeNow:
		return []calendarSection{{ID: calendarSectionFocusedDay, Title: "Focused Day"}, {ID: calendarSectionOverdue, Title: "Overdue"}, {ID: calendarSectionToday, Title: "Today"}}
	default:
		return []calendarSection{{ID: calendarSectionFocusedDay, Title: "Focused Day"}, {ID: calendarSectionOverdue, Title: "Overdue"}, {ID: calendarSectionToday, Title: "Today"}, {ID: calendarSectionReady, Title: "Ready"}}
	}
}

func buildFocusedDaySectionItems(day *calendarDay, titleByID map[string]string) []calendarSectionItem {
	if day == nil {
		return nil
	}
	items := make([]calendarSectionItem, 0, len(day.Tasks))
	for _, task := range day.Tasks {
		items = append(items, newCalendarSectionItem(task, titleByID, ""))
	}
	return items
}

func buildInboxSectionItems(tasks []shelf.Task, titleByID map[string]string) []calendarSectionItem {
	filtered := make([]shelf.Task, 0)
	for _, task := range tasks {
		if task.Kind == "inbox" && task.Status == "open" {
			filtered = append(filtered, task)
		}
	}
	sortTasksForCalendarSection(filtered)
	return buildCalendarSectionItems(filtered, titleByID, nil)
}

func buildOverdueSectionItems(tasks []shelf.Task, titleByID map[string]string) []calendarSectionItem {
	today := time.Now().Local().Format("2006-01-02")
	filtered := make([]shelf.Task, 0)
	for _, task := range tasks {
		if task.DueOn != "" && task.DueOn < today {
			filtered = append(filtered, task)
		}
	}
	sortTasksForCalendarSection(filtered)
	return buildCalendarSectionItems(filtered, titleByID, nil)
}

func buildTodaySectionItems(tasks []shelf.Task, titleByID map[string]string) []calendarSectionItem {
	today := time.Now().Local().Format("2006-01-02")
	filtered := make([]shelf.Task, 0)
	for _, task := range tasks {
		if task.DueOn == today {
			filtered = append(filtered, task)
		}
	}
	sortTasksForCalendarSection(filtered)
	return buildCalendarSectionItems(filtered, titleByID, nil)
}

func buildBlockedSectionItems(tasks []shelf.Task, readiness map[string]shelf.TaskReadiness, titleByID map[string]string) []calendarSectionItem {
	filtered := make([]shelf.Task, 0)
	reasons := map[string]string{}
	for _, task := range tasks {
		info := readiness[task.ID]
		if task.Status == "blocked" || info.BlockedByDeps {
			filtered = append(filtered, task)
			reasons[task.ID] = strings.Join(reviewBlockedBy(task, info, titleByID), "; ")
		}
	}
	sortTasksForCalendarSection(filtered)
	return buildCalendarSectionItems(filtered, titleByID, reasons)
}

func buildReadySectionItems(tasks []shelf.Task, readiness map[string]shelf.TaskReadiness, titleByID map[string]string) []calendarSectionItem {
	filtered := make([]shelf.Task, 0)
	for _, task := range tasks {
		if task.Kind == "inbox" {
			continue
		}
		if readiness[task.ID].Ready {
			filtered = append(filtered, task)
		}
	}
	sortTasksForCalendarSection(filtered)
	return buildCalendarSectionItems(filtered, titleByID, nil)
}

func buildCalendarSectionItems(tasks []shelf.Task, titleByID map[string]string, reasons map[string]string) []calendarSectionItem {
	items := make([]calendarSectionItem, 0, len(tasks))
	for _, task := range tasks {
		reason := ""
		if reasons != nil {
			reason = reasons[task.ID]
		}
		items = append(items, newCalendarSectionItem(task, titleByID, reason))
	}
	return items
}

func newCalendarSectionItem(task shelf.Task, titleByID map[string]string, reason string) calendarSectionItem {
	parentTitle := "root"
	if strings.TrimSpace(task.Parent) != "" {
		if title := strings.TrimSpace(titleByID[task.Parent]); title != "" {
			parentTitle = title
		} else {
			parentTitle = "(missing)"
		}
	}
	dueText := "-"
	if strings.TrimSpace(task.DueOn) != "" {
		dueText = task.DueOn
	}
	return calendarSectionItem{
		Task:     task,
		Subtitle: fmt.Sprintf("%s/%s  due=%s  parent=%s", task.Kind, task.Status, dueText, parentTitle),
		Reason:   reason,
	}
}

func sortTasksForCalendarSection(tasks []shelf.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		leftDue := strings.TrimSpace(tasks[i].DueOn)
		rightDue := strings.TrimSpace(tasks[j].DueOn)
		switch {
		case leftDue == "" && rightDue != "":
			return false
		case leftDue != "" && rightDue == "":
			return true
		case leftDue != rightDue:
			return leftDue < rightDue
		default:
			return tasks[i].ID < tasks[j].ID
		}
	})
}

func findCalendarSectionTask(sections []calendarSection, taskID string) (int, int, bool) {
	for secIdx, section := range sections {
		for rowIdx, item := range section.Items {
			if item.Task.ID == taskID {
				return secIdx, rowIdx, true
			}
		}
	}
	return 0, 0, false
}

func (m *calendarTUIModel) selectFirstAvailableTask() {
	for secIdx, section := range m.sections {
		if len(section.Items) == 0 {
			continue
		}
		m.sectionIndex = secIdx
		m.sectionRows[section.ID] = 0
		m.selectedTaskID = section.Items[0].Task.ID
		return
	}
	m.selectedTaskID = ""
}

func (m *calendarTUIModel) clampSectionSelection() {
	if len(m.sections) == 0 {
		m.sectionIndex = 0
		m.selectedTaskID = ""
		return
	}
	if m.sectionIndex < 0 {
		m.sectionIndex = 0
	}
	if m.sectionIndex >= len(m.sections) {
		m.sectionIndex = len(m.sections) - 1
	}
	section := m.sections[m.sectionIndex]
	row := m.sectionRows[section.ID]
	if len(section.Items) == 0 {
		m.sectionRows[section.ID] = 0
		m.selectedTaskID = ""
		return
	}
	if row < 0 {
		row = 0
	}
	if row >= len(section.Items) {
		row = len(section.Items) - 1
	}
	m.sectionRows[section.ID] = row
	m.selectedTaskID = section.Items[row].Task.ID
}

func (m *calendarTUIModel) moveSectionSelection(delta int) {
	if m.mode == calendarModeTree || m.mode == calendarModeBoard {
		return
	}
	if len(m.sections) == 0 {
		return
	}
	m.sectionIndex += delta
	if m.sectionIndex < 0 {
		m.sectionIndex = len(m.sections) - 1
	}
	if m.sectionIndex >= len(m.sections) {
		m.sectionIndex = 0
	}
	m.clampSectionSelection()
}

func (m *calendarTUIModel) moveSectionRow(delta int) {
	if m.mode == calendarModeTree {
		m.moveTreeRow(delta)
		return
	}
	section := m.currentSection()
	if section == nil || len(section.Items) == 0 {
		return
	}
	row := m.sectionRows[section.ID] + delta
	if row < 0 {
		row = len(section.Items) - 1
	}
	if row >= len(section.Items) {
		row = 0
	}
	m.sectionRows[section.ID] = row
	m.selectedTaskID = section.Items[row].Task.ID
}

func (m *calendarTUIModel) jumpSectionRowStart() {
	if m.mode == calendarModeTree {
		if len(m.treeRows) == 0 {
			return
		}
		m.treeRowIndex = 0
		m.selectedTaskID = m.treeRows[0].Task.ID
		return
	}
	section := m.currentSection()
	if section == nil || len(section.Items) == 0 {
		return
	}
	m.sectionRows[section.ID] = 0
	m.selectedTaskID = section.Items[0].Task.ID
}

func (m *calendarTUIModel) jumpSectionRowEnd() {
	if m.mode == calendarModeTree {
		if len(m.treeRows) == 0 {
			return
		}
		m.treeRowIndex = len(m.treeRows) - 1
		m.selectedTaskID = m.treeRows[m.treeRowIndex].Task.ID
		return
	}
	section := m.currentSection()
	if section == nil || len(section.Items) == 0 {
		return
	}
	row := len(section.Items) - 1
	m.sectionRows[section.ID] = row
	m.selectedTaskID = section.Items[row].Task.ID
}

func (m *calendarTUIModel) jumpToSection(index int) {
	if m.mode == calendarModeTree || m.mode == calendarModeBoard {
		return
	}
	if index < 0 || index >= len(m.sections) {
		return
	}
	m.sectionIndex = index
	m.clampSectionSelection()
}

func (m *calendarTUIModel) currentSection() *calendarSection {
	if m.mode == calendarModeTree || m.mode == calendarModeBoard {
		return nil
	}
	if len(m.sections) == 0 || m.sectionIndex < 0 || m.sectionIndex >= len(m.sections) {
		return nil
	}
	return &m.sections[m.sectionIndex]
}

func (m *calendarTUIModel) focusedDaySection() *calendarSection {
	for i := range m.sections {
		if m.sections[i].ID == calendarSectionFocusedDay {
			return &m.sections[i]
		}
	}
	return nil
}

func (m *calendarTUIModel) moveTreeRow(delta int) {
	if len(m.treeRows) == 0 {
		return
	}
	m.treeRowIndex += delta
	if m.treeRowIndex < 0 {
		m.treeRowIndex = len(m.treeRows) - 1
	}
	if m.treeRowIndex >= len(m.treeRows) {
		m.treeRowIndex = 0
	}
	m.selectedTaskID = m.treeRows[m.treeRowIndex].Task.ID
}

func (m *calendarTUIModel) moveFocusedDayTask(delta int) {
	section := m.focusedDaySection()
	if section == nil || len(section.Items) == 0 {
		m.message = "focused day に task がありません"
		return
	}
	row := m.sectionRows[section.ID] + delta
	if row < 0 {
		row = len(section.Items) - 1
	}
	if row >= len(section.Items) {
		row = 0
	}
	m.sectionRows[section.ID] = row
	m.selectedTaskID = section.Items[row].Task.ID
}

func (m *calendarTUIModel) clampBoardSelection() {
	if len(m.boardColumns) == 0 {
		m.boardColumnIdx = 0
		m.selectedTaskID = ""
		return
	}
	if m.boardColumnIdx < 0 {
		m.boardColumnIdx = 0
	}
	if m.boardColumnIdx >= len(m.boardColumns) {
		m.boardColumnIdx = len(m.boardColumns) - 1
	}
	column := m.boardColumns[m.boardColumnIdx]
	maxRow := len(column.Tasks) - 1
	if maxRow < 0 {
		m.boardRowIndex[m.boardColumnIdx] = 0
		m.selectedTaskID = ""
		return
	}
	row := m.boardRowIndex[m.boardColumnIdx]
	if row < 0 {
		row = 0
	}
	if row > maxRow {
		row = maxRow
	}
	m.boardRowIndex[m.boardColumnIdx] = row
	m.selectedTaskID = column.Tasks[row].ID
}

func (m *calendarTUIModel) moveBoardColumn(delta int) {
	if len(m.boardColumns) == 0 {
		return
	}
	m.boardColumnIdx += delta
	if m.boardColumnIdx < 0 {
		m.boardColumnIdx = len(m.boardColumns) - 1
	}
	if m.boardColumnIdx >= len(m.boardColumns) {
		m.boardColumnIdx = 0
	}
	m.clampBoardSelection()
}

func (m *calendarTUIModel) moveBoardRow(delta int) {
	if len(m.boardColumns) == 0 {
		return
	}
	column := m.boardColumns[m.boardColumnIdx]
	if len(column.Tasks) == 0 {
		return
	}
	row := m.boardRowIndex[m.boardColumnIdx] + delta
	if row < 0 {
		row = len(column.Tasks) - 1
	}
	if row >= len(column.Tasks) {
		row = 0
	}
	m.boardRowIndex[m.boardColumnIdx] = row
	m.selectedTaskID = column.Tasks[row].ID
}

func (m *calendarTUIModel) jumpBoardRowStart() {
	if len(m.boardColumns) == 0 {
		return
	}
	column := m.boardColumns[m.boardColumnIdx]
	if len(column.Tasks) == 0 {
		return
	}
	m.boardRowIndex[m.boardColumnIdx] = 0
	m.selectedTaskID = column.Tasks[0].ID
}

func (m *calendarTUIModel) jumpBoardRowEnd() {
	if len(m.boardColumns) == 0 {
		return
	}
	column := m.boardColumns[m.boardColumnIdx]
	if len(column.Tasks) == 0 {
		return
	}
	row := len(column.Tasks) - 1
	m.boardRowIndex[m.boardColumnIdx] = row
	m.selectedTaskID = column.Tasks[row].ID
}

func (m *calendarTUIModel) moveFocusByDays(delta int) {
	target, err := m.focusedDate()
	if err != nil {
		m.message = err.Error()
		return
	}
	m.moveFocusToDate(target.AddDate(0, 0, delta))
}

func (m *calendarTUIModel) moveFocusByMonths(delta int) {
	target, err := m.focusedDate()
	if err != nil {
		m.message = err.Error()
		return
	}
	m.moveFocusToDate(target.AddDate(0, delta, 0))
}

func (m *calendarTUIModel) moveFocusToDate(target time.Time) {
	newStart, newIndex := planCalendarWindow(m.startDate, m.daysCount, target)
	if !sameDate(newStart, m.startDate) {
		m.startDate = newStart
		if err := m.reload(); err != nil {
			m.message = err.Error()
			return
		}
	}
	m.dayIndex = newIndex
	m.rebuildModeState()
}

func (m calendarTUIModel) focusedDay() *calendarDay {
	if len(m.days) == 0 || m.dayIndex < 0 || m.dayIndex >= len(m.days) {
		return nil
	}
	return &m.days[m.dayIndex]
}

func (m calendarTUIModel) focusedDate() (time.Time, error) {
	day := m.focusedDay()
	if day == nil {
		return time.Time{}, fmt.Errorf("選択中の日付がありません")
	}
	return time.ParseInLocation("2006-01-02", day.Date, time.Now().Location())
}

func (m calendarTUIModel) focusedDayLabel() string {
	day := m.focusedDay()
	if day == nil {
		return ""
	}
	return day.Date
}

func (m calendarTUIModel) selectedTask() (shelf.Task, bool) {
	if m.mode == calendarModeCalendar {
		section := m.focusedDaySection()
		if section == nil || len(section.Items) == 0 {
			return shelf.Task{}, false
		}
		row := m.sectionRows[section.ID]
		if row < 0 || row >= len(section.Items) {
			row = 0
		}
		return section.Items[row].Task, true
	}
	if m.mode == calendarModeTree {
		if len(m.treeRows) == 0 || m.treeRowIndex < 0 || m.treeRowIndex >= len(m.treeRows) {
			return shelf.Task{}, false
		}
		return m.treeRows[m.treeRowIndex].Task, true
	}
	if m.mode == calendarModeBoard {
		if len(m.boardColumns) == 0 || m.boardColumnIdx < 0 || m.boardColumnIdx >= len(m.boardColumns) {
			return shelf.Task{}, false
		}
		column := m.boardColumns[m.boardColumnIdx]
		if len(column.Tasks) == 0 {
			return shelf.Task{}, false
		}
		row := m.boardRowIndex[m.boardColumnIdx]
		if row < 0 || row >= len(column.Tasks) {
			return shelf.Task{}, false
		}
		return column.Tasks[row], true
	}
	section := m.currentSection()
	if section == nil || len(section.Items) == 0 {
		return shelf.Task{}, false
	}
	row := m.sectionRows[section.ID]
	if row < 0 || row >= len(section.Items) {
		return shelf.Task{}, false
	}
	return section.Items[row].Task, true
}

func (m calendarTUIModel) openEditorForSelectedTask() (tea.Model, tea.Cmd) {
	task, ok := m.selectedTask()
	if !ok {
		m.message = "選択中の task がありません"
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

func (m *calendarTUIModel) createTaskOnFocusedDay(title string) error {
	day := m.focusedDay()
	if day == nil {
		return fmt.Errorf("選択中の日付がありません")
	}
	input := shelf.AddTaskInput{Title: title, Kind: m.defaultKind, Status: m.defaultStatus, DueOn: day.Date}
	created := shelf.Task{}
	if err := withWriteLock(m.rootDir, func() error {
		if err := prepareUndoSnapshot(m.rootDir, "calendar-add"); err != nil {
			return err
		}
		var err error
		created, err = shelf.AddTask(m.rootDir, input)
		return err
	}); err != nil {
		return err
	}
	if calendarStatusIncluded(m.statuses, created.Status) {
		if err := m.reload(); err != nil {
			return err
		}
		m.selectTaskByID(created.ID)
		m.message = fmt.Sprintf("Created %s on %s", created.Title, created.DueOn)
		return nil
	}
	m.visibleTasks = append(m.visibleTasks, created)
	m.taskByID[created.ID] = created
	m.titleByID[created.ID] = created.Title
	m.insertTaskOnFocusedDay(created)
	m.rebuildModeState()
	m.selectTaskByID(created.ID)
	m.message = fmt.Sprintf("Created %s on %s (current filter excludes it; visible until reload)", created.Title, created.DueOn)
	return nil
}

func (m *calendarTUIModel) applySnoozeOption(option snoozePreset) error {
	task, ok := m.selectedTask()
	if !ok {
		return fmt.Errorf("選択中の task がありません")
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
	m.selectTaskByID(task.ID)
	m.message = fmt.Sprintf("Snoozed %s to %s", task.Title, nextDue)
	return nil
}

func (m calendarTUIModel) applyStatusChange(nextStatus shelf.Status) (tea.Model, tea.Cmd) {
	task, ok := m.selectedTask()
	if !ok {
		m.message = "no task selected"
		return m, nil
	}
	updatedTask := task
	err := withWriteLock(m.rootDir, func() error {
		if err := prepareUndoSnapshot(m.rootDir, "calendar-status"); err != nil {
			return err
		}
		var err error
		updatedTask, err = shelf.SetTask(m.rootDir, task.ID, shelf.SetTaskInput{Status: &nextStatus})
		return err
	})
	if err != nil {
		m.message = err.Error()
		return m, nil
	}
	if calendarStatusIncluded(m.statuses, nextStatus) {
		if err := m.reload(); err != nil {
			m.message = err.Error()
			return m, nil
		}
		m.selectTaskByID(updatedTask.ID)
		m.message = fmt.Sprintf("%s -> %s", task.Title, nextStatus)
		return m, nil
	}
	m.replaceTaskInVisibleState(updatedTask)
	m.rebuildModeState()
	m.selectTaskByID(updatedTask.ID)
	m.message = fmt.Sprintf("%s -> %s (current filter excludes it; visible until reload)", task.Title, nextStatus)
	return m, nil
}

func calendarStatusIncluded(statuses []shelf.Status, target shelf.Status) bool {
	for _, status := range statuses {
		if status == target {
			return true
		}
	}
	return false
}

func (m *calendarTUIModel) replaceTaskInVisibleState(task shelf.Task) {
	for i := range m.visibleTasks {
		if m.visibleTasks[i].ID == task.ID {
			m.visibleTasks[i] = task
			break
		}
	}
	m.taskByID[task.ID] = task
	m.titleByID[task.ID] = task.Title
	m.updateTaskInDays(task)
}

func (m *calendarTUIModel) updateTaskInDays(task shelf.Task) {
	for dayIndex := range m.days {
		for taskIndex := range m.days[dayIndex].Tasks {
			if m.days[dayIndex].Tasks[taskIndex].ID == task.ID {
				m.days[dayIndex].Tasks[taskIndex] = task
			}
		}
	}
}

func (m *calendarTUIModel) insertTaskOnFocusedDay(task shelf.Task) {
	day := m.focusedDay()
	if day == nil {
		return
	}
	day.Tasks = append(day.Tasks, task)
}

func (m *calendarTUIModel) selectTaskByID(taskID string) {
	m.selectedTaskID = taskID
	if m.mode == calendarModeTree {
		for i, row := range m.treeRows {
			if row.Task.ID == taskID {
				m.treeRowIndex = i
				return
			}
		}
		return
	}
	if m.mode == calendarModeBoard {
		for colIdx, column := range m.boardColumns {
			for rowIdx, task := range column.Tasks {
				if task.ID == taskID {
					m.boardColumnIdx = colIdx
					m.boardRowIndex[colIdx] = rowIdx
					return
				}
			}
		}
		m.clampBoardSelection()
		return
	}
	if secIdx, rowIdx, ok := findCalendarSectionTask(m.sections, taskID); ok {
		m.sectionIndex = secIdx
		m.sectionRows[m.sections[secIdx].ID] = rowIdx
		return
	}
	m.clampSectionSelection()
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
			week = append(week, calendarMonthCell{Date: current, InCurrentMonth: current.Month() == focusDate.Month(), InRange: ok, TaskCount: taskCounts[key], DominantStatus: dominant[key]})
			current = current.AddDate(0, 0, 1)
		}
		weeks = append(weeks, week)
	}

	return calendarMonthView{Label: focusDate.Format("January 2006"), Weeks: weeks}
}

func renderCockpitHeader(m calendarTUIModel, focused time.Time) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	parts := []string{
		titleStyle.Render("Daily Cockpit"),
		accentStyle.Render("[" + strings.ToUpper(string(m.mode)) + "]"),
		metaStyle.Render("Focus " + focused.Format("Mon 2006-01-02")),
		metaStyle.Render(fmt.Sprintf("Range %s..%s", m.days[0].Date, m.days[len(m.days)-1].Date)),
		metaStyle.Render("Filter " + formatCalendarStatusFilter(m.statuses)),
		metaStyle.Render("?:help"),
	}
	return trimLine(strings.Join(parts, "  "), max(48, m.width-2))
}

func renderCalendarViewport(topParts []string, body string, bottomParts []string, totalHeight int, scroll int) string {
	sections := make([]string, 0, 3)
	topBlock := strings.Join(topParts, "\n\n")
	bottomBlock := strings.Join(bottomParts, "\n\n")
	bodyBlock := body
	if totalHeight > 0 {
		bodyHeight := availableViewportHeight(totalHeight, topBlock, bottomBlock)
		bodyBlock, scroll = clipRenderedBlock(body, bodyHeight, scroll)
		_ = scroll
	}
	if strings.TrimSpace(topBlock) != "" {
		sections = append(sections, topBlock)
	}
	sections = append(sections, bodyBlock)
	if strings.TrimSpace(bottomBlock) != "" {
		sections = append(sections, bottomBlock)
	}
	return strings.Join(sections, "\n\n")
}

func availableViewportHeight(totalHeight int, topBlock string, bottomBlock string) int {
	if totalHeight <= 0 {
		return 0
	}
	height := totalHeight
	if strings.TrimSpace(topBlock) != "" {
		height -= lipgloss.Height(topBlock)
		height--
	}
	if strings.TrimSpace(bottomBlock) != "" {
		height -= lipgloss.Height(bottomBlock)
		height--
	}
	return max(3, height)
}

func clipRenderedBlock(block string, height int, scroll int) (string, int) {
	if height <= 0 {
		return block, 0
	}
	lines := strings.Split(strings.TrimSuffix(block, "\n"), "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	maxScroll := max(0, len(lines)-height)
	if scroll < 0 {
		scroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}
	end := min(len(lines), scroll+height)
	return strings.Join(lines[scroll:end], "\n"), maxScroll
}

func renderCockpitHelpOverlay(mode calendarMode) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("141")).
		Padding(0, 1)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	lines := []string{
		titleStyle.Render("Help"),
		mutedStyle.Render(fmt.Sprintf("mode=%s", mode)),
		"Tab: pane  C/T/B/R/N: mode  ?: close",
		"h/l: move  j/k: rows or weeks  n/p: task or tabs or columns",
		"PgUp/PgDn or Ctrl+U/D: scroll body  Home/End: top/bottom",
		"sidebar: in non-calendar modes, focus the right pane to move dates",
		"o/i/b/d/c: status  a: add  e: edit  z: snooze  r: reload",
		"Enter: details  q/Esc/Ctrl+C: quit",
	}
	return boxStyle.Render(strings.Join(lines, "\n"))
}

func renderCalendarLegend() string {
	item := func(color lipgloss.Color, label string) string {
		return lipgloss.NewStyle().Foreground(color).Render("■ " + label)
	}
	return strings.Join([]string{item(lipgloss.Color("203"), "blocked"), item(lipgloss.Color("220"), "in_progress"), item(lipgloss.Color("81"), "open"), item(lipgloss.Color("78"), "done"), item(lipgloss.Color("245"), "cancelled")}, "  ")
}

func renderCalendarModeTabs(mode calendarMode) string {
	tabs := []struct {
		mode  calendarMode
		label string
	}{
		{calendarModeCalendar, "Calendar"},
		{calendarModeTree, "Tree"},
		{calendarModeBoard, "Board"},
		{calendarModeReview, "Review"},
		{calendarModeNow, "Now"},
	}
	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("81")).
		Bold(true).
		Padding(0, 1)
	idleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 1)
	parts := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		style := idleStyle
		if tab.mode == mode {
			style = activeStyle
		}
		parts = append(parts, style.Render(tab.label))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

func renderCalendarSectionTabs(sections []calendarSection, selected int, width int) string {
	if len(sections) == 0 {
		return ""
	}
	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("24")).
		Bold(true).
		Padding(0, 1)
	idleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Padding(0, 1)
	parts := make([]string, 0, len(sections))
	for i, section := range sections {
		label := fmt.Sprintf("%s %d", section.Title, len(section.Items))
		style := idleStyle
		if i == selected {
			style = activeStyle
		}
		parts = append(parts, style.Render(label))
	}
	return trimLine(lipgloss.JoinHorizontal(lipgloss.Left, parts...), max(12, width))
}

func renderCalendarMainPane(m calendarTUIModel, month calendarMonthView, width int, active bool) string {
	borderColor := lipgloss.Color("240")
	if active {
		borderColor = lipgloss.Color("45")
	}
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(width)

	if m.mode == calendarModeTree {
		parts := []string{
			renderCockpitContextStrip(m, max(24, width-4)),
			renderCockpitTreePane(m.treeRows, m.treeRowIndex, max(20, width-4)),
		}
		return boxStyle.Render(strings.Join(parts, "\n\n"))
	}
	if m.mode == calendarModeBoard {
		parts := []string{
			renderCockpitContextStrip(m, max(24, width-4)),
			renderCockpitBoardPane(m.boardColumns, m.boardColumnIdx, m.boardRowIndex, m.showID, max(20, width-4)),
		}
		return boxStyle.Render(strings.Join(parts, "\n\n"))
	}
	parts := make([]string, 0, 4)
	if m.mode != calendarModeCalendar {
		parts = append(parts, renderCockpitContextStrip(m, max(24, width-4)))
	}
	if m.mode != calendarModeCalendar && m.mode != calendarModeNow {
		parts = append(parts, renderCalendarSectionTabs(m.sections, m.sectionIndex, max(20, width-4)))
	}
	if m.mode == calendarModeCalendar {
		parts = append(parts, renderCalendarLegend())
		parts = append(parts, renderCalendarMonth(month, m.focusedDayLabel(), max(56, width-4), false))
		return boxStyle.Render(strings.Join(parts, "\n\n"))
	}
	if m.mode == calendarModeNow {
		parts = append(parts, renderCalendarTriptychSections(m.sections, m.sectionIndex, m.sectionRows, m.showID, max(36, width-4)))
		return boxStyle.Render(strings.Join(parts, "\n\n"))
	}
	parts = append(parts, renderCalendarActiveSection(m.currentSection(), m.sectionRows, m.showID, max(20, width-4)))
	return boxStyle.Render(strings.Join(parts, "\n\n"))
}

func renderCockpitContextStrip(m calendarTUIModel, width int) string {
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	parts := []string{
		"Focus " + m.focusedDayLabel(),
	}
	switch m.mode {
	case calendarModeReview:
		parts = append(parts,
			fmt.Sprintf("Inbox %d", countSectionItems(m.sections, calendarSectionInbox)),
			fmt.Sprintf("Overdue %d", countSectionItems(m.sections, calendarSectionOverdue)),
			fmt.Sprintf("Today %d", countSectionItems(m.sections, calendarSectionToday)),
			fmt.Sprintf("Blocked %d", countSectionItems(m.sections, calendarSectionBlocked)),
			fmt.Sprintf("Ready %d", countSectionItems(m.sections, calendarSectionReady)),
		)
	case calendarModeNow:
		parts = append(parts,
			fmt.Sprintf("Overdue %d", countSectionItems(m.sections, calendarSectionOverdue)),
			fmt.Sprintf("Today %d", countSectionItems(m.sections, calendarSectionToday)),
		)
	default:
		parts = append(parts, fmt.Sprintf("Visible %d", len(m.visibleTasks)))
	}
	return mutedStyle.Render(trimLine(strings.Join(parts, "  ·  "), width))
}

func countSectionItems(sections []calendarSection, target calendarSectionID) int {
	for _, section := range sections {
		if section.ID == target {
			return len(section.Items)
		}
	}
	return 0
}

func renderCalendarSidebarPane(m calendarTUIModel, task shelf.Task, ok bool, width int) string {
	focusedSection := m.focusedDaySection()
	dayList := renderCalendarActiveSection(focusedSection, m.sectionRows, m.showID, width)
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("45")).
		Padding(0, 1).
		Width(width)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	focusedPane := boxStyle.Render(strings.Join([]string{
		titleStyle.Render("Focused Day"),
		metaStyle.Render("n/p: task switch"),
		dayList,
	}, "\n"))
	inspector := renderCalendarInspectorPane(task, ok, m.showID, m.showTaskBody, m.taskByID, m.readiness, m.outboundCount, m.inboundCount, width)
	return lipgloss.JoinVertical(lipgloss.Left, focusedPane, "", inspector)
}

func renderCalendarSecondarySidebarPane(m calendarTUIModel, task shelf.Task, ok bool, width int, active bool) string {
	calendarPane := renderCalendarMiniSidebar(m, width, active)
	inspector := renderCalendarInspectorPane(task, ok, m.showID, m.showTaskBody, m.taskByID, m.readiness, m.outboundCount, m.inboundCount, width)
	return lipgloss.JoinVertical(lipgloss.Left, calendarPane, "", inspector)
}

func renderCalendarMiniSidebar(m calendarTUIModel, width int, active bool) string {
	focused, err := m.focusedDate()
	if err != nil {
		focused = time.Now().Local()
	}
	month := buildCalendarMonthView(m.days, focused)
	borderColor := lipgloss.Color("240")
	if active {
		borderColor = lipgloss.Color("45")
	}
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(width)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	return boxStyle.Render(strings.Join([]string{
		titleStyle.Render("Calendar"),
		metaStyle.Render("Tab then h/j/k/l/[ ]: date move"),
		renderCalendarMonth(month, m.focusedDayLabel(), max(28, width-4), true),
	}, "\n"))
}

func renderCockpitTreePane(rows []cockpitTreeRow, selected int, width int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	lines := []string{titleStyle.Render("Tree")}
	if len(rows) == 0 {
		lines = append(lines, mutedStyle.Render("(none)"))
		return strings.Join(lines, "\n")
	}
	for i, row := range rows {
		label := trimLine(row.Label, max(20, width))
		meta := trimLine(row.Meta, max(18, width-2))
		if i == selected {
			lines = append(lines, selectedStyle.Render("> "+label))
			lines = append(lines, mutedStyle.Render("  "+meta))
		} else {
			lines = append(lines, "  "+label)
			lines = append(lines, mutedStyle.Render("  "+meta))
		}
	}
	return strings.Join(lines, "\n")
}

func renderCockpitBoardPane(columns []boardColumn, selectedColumn int, rowIndex map[int]int, showID bool, width int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	if len(columns) == 0 {
		return titleStyle.Render("Board") + "\n" + mutedStyle.Render("(none)")
	}
	columnGap := 3
	columnWidth := max(18, (width-(len(columns)-1)*columnGap)/max(1, len(columns)))
	if columnWidth < 16 {
		columnWidth = 16
	}
	rendered := make([]string, 0, len(columns))
	for colIdx, column := range columns {
		header := fmt.Sprintf("%s %d", column.Status, len(column.Tasks))
		if colIdx == selectedColumn {
			header = selectedStyle.Render(header)
		} else {
			header = titleStyle.Render(header)
		}
		lines := []string{header}
		if len(column.Tasks) == 0 {
			lines = append(lines, mutedStyle.Render("(none)"))
		}
		currentRow := rowIndex[colIdx]
		for taskIdx, task := range column.Tasks {
			label := task.Title
			if showID {
				label = fmt.Sprintf("[%s] %s", shelf.ShortID(task.ID), label)
			}
			meta := trimLine(renderBoardTaskMeta(task), max(14, columnWidth-2))
			prefix := "  "
			if colIdx == selectedColumn && taskIdx == currentRow {
				prefix = "> "
				lines = append(lines, selectedStyle.Render(trimLine(prefix+label, max(14, columnWidth-2))))
				lines = append(lines, mutedStyle.Render("  "+meta))
			} else {
				lines = append(lines, prefix+trimLine(label, max(14, columnWidth-2)))
				lines = append(lines, mutedStyle.Render("  "+meta))
			}
		}
		style := lipgloss.NewStyle().Width(columnWidth)
		rendered = append(rendered, style.Render(strings.Join(lines, "\n")))
	}
	return joinFixedColumns(rendered, " │ ")
}

func renderCalendarGridPane(month calendarMonthView, focusedDate string, width int, active bool) string {
	borderColor := lipgloss.Color("240")
	if active {
		borderColor = lipgloss.Color("45")
	}
	containerStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Padding(0, 1).Width(width)
	content := renderCalendarMonth(month, focusedDate, max(42, width-4), true)
	return containerStyle.Render(content)
}

func renderCalendarActiveSection(section *calendarSection, sectionRows map[calendarSectionID]int, showID bool, width int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)

	if section == nil {
		return mutedStyle.Render("No section")
	}
	lines := []string{titleStyle.Render(fmt.Sprintf("%s %d", section.Title, len(section.Items)))}
	if len(section.Items) == 0 {
		lines = append(lines, mutedStyle.Render("(none)"))
		return strings.Join(lines, "\n")
	}
	row := sectionRows[section.ID]
	if row < 0 {
		row = 0
	}
	if row >= len(section.Items) {
		row = len(section.Items) - 1
	}
	for i, item := range section.Items {
		label := item.Task.Title
		if showID {
			label = fmt.Sprintf("[%s] %s", shelf.ShortID(item.Task.ID), label)
		}
		rendered := "  " + trimLine(label, max(18, width))
		if i == row {
			rendered = selectedStyle.Render("> " + trimLine(label, max(18, width)))
		}
		lines = append(lines, rendered)
		lines = append(lines, mutedStyle.Render("  "+trimLine(item.Subtitle, max(18, width-2))))
		if i == row && strings.TrimSpace(item.Reason) != "" {
			lines = append(lines, mutedStyle.Render("  "+trimLine(item.Reason, max(18, width-2))))
		}
	}
	return strings.Join(lines, "\n")
}

func renderCalendarTriptychSections(sections []calendarSection, selected int, sectionRows map[calendarSectionID]int, showID bool, width int) string {
	if len(sections) == 0 {
		return ""
	}
	gap := 3
	columnWidth := max(22, (width-gap*2)/3)
	rendered := make([]string, 0, 3)
	for i, section := range sections {
		rendered = append(rendered, lipgloss.NewStyle().Width(columnWidth).Render(
			renderCalendarSectionColumn(&section, i == selected, sectionRows, showID, columnWidth),
		))
	}
	return joinFixedColumns(rendered, " │ ")
}

func renderCalendarSectionColumn(section *calendarSection, active bool, sectionRows map[calendarSectionID]int, showID bool, width int) string {
	if section == nil {
		return ""
	}
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("250")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Width(max(12, width))
	if active {
		titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("24")).
			Padding(0, 1).
			Width(max(12, width))
	}
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	header := trimLine(fmt.Sprintf("%s %d", section.Title, len(section.Items)), max(8, width-2))
	lines := []string{titleStyle.Render(header)}
	if len(section.Items) == 0 {
		lines = append(lines, mutedStyle.Render("(none)"))
		return strings.Join(lines, "\n")
	}
	row := sectionRows[section.ID]
	if row < 0 {
		row = 0
	}
	if row >= len(section.Items) {
		row = len(section.Items) - 1
	}
	for i, item := range section.Items {
		label := item.Task.Title
		if showID {
			label = fmt.Sprintf("[%s] %s", shelf.ShortID(item.Task.ID), label)
		}
		if i == row && active {
			lines = append(lines, selectedStyle.Render("> "+trimLine(label, max(18, width))))
		} else {
			lines = append(lines, "  "+trimLine(label, max(18, width)))
		}
		lines = append(lines, mutedStyle.Render("  "+trimLine(item.Subtitle, max(18, width-2))))
		if i == row && active && strings.TrimSpace(item.Reason) != "" {
			lines = append(lines, mutedStyle.Render("  "+trimLine(item.Reason, max(18, width-2))))
		}
	}
	return strings.Join(lines, "\n")
}

func joinFixedColumns(columns []string, separator string) string {
	if len(columns) == 0 {
		return ""
	}
	if len(columns) == 1 {
		return columns[0]
	}
	separatorView := lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render(separator)
	parts := make([]string, 0, len(columns)*2-1)
	for i, column := range columns {
		if i > 0 {
			parts = append(parts, separatorView)
		}
		parts = append(parts, column)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func renderBoardTaskMeta(task shelf.Task) string {
	meta := fmt.Sprintf("%s/%s", task.Kind, task.Status)
	if strings.TrimSpace(task.DueOn) != "" {
		meta += "  due=" + task.DueOn
	}
	return meta
}

func renderCalendarMonth(month calendarMonthView, focusedDate string, width int, compact bool) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	dayHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244"))
	maxCellWidth := 12
	if !compact {
		maxCellWidth = 13
	}
	cellWidth := max(8, min(maxCellWidth, (width-2)/7))
	headers := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	headerCells := make([]string, 0, len(headers))
	for _, header := range headers {
		headerCells = append(headerCells, dayHeaderStyle.Width(cellWidth).Align(lipgloss.Center).Render(header))
	}
	rows := []string{titleStyle.Render(month.Label), lipgloss.JoinHorizontal(lipgloss.Top, headerCells...)}
	for _, week := range month.Weeks {
		cells := make([]string, 0, len(week))
		for _, cell := range week {
			cells = append(cells, renderCalendarCell(cell, focusedDate, cellWidth, compact))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	return strings.Join(rows, "\n")
}

func renderCalendarCell(cell calendarMonthCell, focusedDate string, cellWidth int, compact bool) string {
	key := cell.Date.Format("2006-01-02")
	today := time.Now().Format("2006-01-02")
	contentWidth := max(6, cellWidth)
	height := 2
	if !compact {
		height = 4
	}
	style := lipgloss.NewStyle().Width(contentWidth).Height(height)
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
		dayLabel += "*"
	}
	countLine := ""
	if cell.TaskCount > 0 {
		countLine = fmt.Sprintf("• %d", cell.TaskCount)
	} else if cell.InRange {
		countLine = "·"
	}
	lines := []string{padOrTrim(dayLabel, contentWidth)}
	if !compact {
		lines = append(lines, padOrTrim("", contentWidth))
	}
	lines = append(lines, padOrTrim(countLine, contentWidth))
	return style.Render(strings.Join(lines, "\n"))
}

func renderCalendarSectionsPane(sections []calendarSection, sectionIndex int, sectionRows map[calendarSectionID]int, showID bool, width int, active bool) string {
	borderColor := lipgloss.Color("240")
	if active {
		borderColor = lipgloss.Color("45")
	}
	boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Padding(0, 1).Width(width)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	sectionTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))

	lines := []string{titleStyle.Render("Daily Cockpit")}
	for secIdx, section := range sections {
		heading := fmt.Sprintf("%d. %s (%d)", secIdx+1, section.Title, len(section.Items))
		if secIdx == sectionIndex {
			heading = sectionTitleStyle.Render("> " + heading)
		} else {
			heading = sectionTitleStyle.Render("  " + heading)
		}
		lines = append(lines, heading)
		if len(section.Items) == 0 {
			lines = append(lines, mutedStyle.Render("    (none)"))
			continue
		}
		for rowIdx, item := range section.Items {
			label := item.Task.Title
			if showID {
				label = fmt.Sprintf("[%s] %s", shelf.ShortID(item.Task.ID), label)
			}
			line := fmt.Sprintf("    %s", trimLine(label, max(16, width-8)))
			if secIdx == sectionIndex && rowIdx == sectionRows[section.ID] {
				line = selectedStyle.Render("  > " + trimLine(label, max(16, width-8)))
			}
			lines = append(lines, line)
			subtitle := "      " + trimLine(item.Subtitle, max(16, width-8))
			lines = append(lines, mutedStyle.Render(subtitle))
			if strings.TrimSpace(item.Reason) != "" {
				lines = append(lines, mutedStyle.Render("      "+trimLine(item.Reason, max(16, width-8))))
			}
		}
	}
	return boxStyle.Render(strings.Join(lines, "\n"))
}

func renderCalendarInspectorPane(task shelf.Task, ok bool, showID bool, showTaskBody bool, taskByID map[string]shelf.Task, readiness map[string]shelf.TaskReadiness, outboundCount map[string]int, inboundCount map[string]int, width int) string {
	boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1).Width(width)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	lines := []string{titleStyle.Render("Inspector")}
	if !ok {
		lines = append(lines, mutedStyle.Render("No task selected"), mutedStyle.Render("main pane から task を選択してください"))
		return boxStyle.Render(strings.Join(lines, "\n"))
	}
	label := task.Title
	if showID {
		label = fmt.Sprintf("[%s] %s", shelf.ShortID(task.ID), label)
	}
	lines = append(lines,
		titleStyle.Render(label),
		fmt.Sprintf("kind=%s  status=%s", uiKind(task.Kind), uiStatus(task.Status)),
		fmt.Sprintf("due=%s  repeat=%s", uiDue(task.DueOn), formatCalendarRepeat(task.RepeatEvery)),
	)
	if len(task.Tags) > 0 {
		lines = append(lines, fmt.Sprintf("tags=%s", strings.Join(task.Tags, ", ")))
	}
	path := buildTaskPath(task, taskByID)
	lines = append(lines, fmt.Sprintf("%s %s", labelStyle.Render("path:"), trimLine(path, max(20, width-6))))
	body := strings.TrimSpace(task.Body)
	if body == "" {
		lines = append(lines, mutedStyle.Render("(empty body)"))
		return boxStyle.Render(strings.Join(lines, "\n"))
	}
	bodyLines := strings.Split(body, "\n")
	maxLines := 4
	if showTaskBody {
		maxLines = 12
		lines = append(lines, mutedStyle.Render("details expanded"))
		if task.EstimateMin > 0 || task.SpentMin > 0 {
			lines = append(lines, fmt.Sprintf("estimate=%s  spent=%s", shelf.FormatWorkMinutes(task.EstimateMin), shelf.FormatWorkMinutes(task.SpentMin)))
		}
		if strings.TrimSpace(task.TimerStart) != "" {
			lines = append(lines, fmt.Sprintf("timer=%s", task.TimerStart))
		}
		lines = append(lines, fmt.Sprintf("links=out:%d in:%d", outboundCount[task.ID], inboundCount[task.ID]))
		if info, exists := readiness[task.ID]; exists {
			if len(info.UnresolvedDependsOn) > 0 {
				labels := make([]string, 0, len(info.UnresolvedDependsOn))
				for _, depID := range info.UnresolvedDependsOn {
					if title := strings.TrimSpace(taskByID[depID].Title); title != "" {
						labels = append(labels, title)
					} else {
						labels = append(labels, depID)
					}
				}
				lines = append(lines, fmt.Sprintf("depends_on=%s", strings.Join(labels, ", ")))
			}
			lines = append(lines, fmt.Sprintf("ready=%t", info.Ready))
		}
	} else {
		lines = append(lines, mutedStyle.Render("compact details"))
	}
	if len(bodyLines) > maxLines {
		bodyLines = append(bodyLines[:maxLines], "...")
	}
	lines = append(lines, bodyLines...)
	return boxStyle.Render(strings.Join(lines, "\n"))
}

func formatCalendarRepeat(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func formatCalendarStatusFilter(statuses []shelf.Status) string {
	labels := make([]string, 0, len(statuses))
	for _, status := range statuses {
		labels = append(labels, string(status))
	}
	return strings.Join(labels, ",")
}

func renderCalendarSnoozePicker(selected int) string {
	boxStyle := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("141")).Padding(0, 1)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	lines := []string{titleStyle.Render("Snooze Presets"), helpStyle.Render("j/k: 移動  Enter: 決定  Esc/q: 戻る")}
	for i, option := range calendarSnoozeOptions() {
		line := "  " + option.Label
		if i == selected {
			line = selectedStyle.Render("> " + option.Label)
		}
		lines = append(lines, line)
	}
	return boxStyle.Render(strings.Join(lines, "\n"))
}

func renderCalendarAddComposer(date string, defaultKind shelf.Kind, defaultStatus shelf.Status, title string) string {
	boxStyle := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("81")).Padding(0, 1)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	lines := []string{titleStyle.Render("Add Task"), helpStyle.Render("type: title  Enter: create  Esc: cancel"), fmt.Sprintf("due=%s  kind=%s  status=%s", date, defaultKind, defaultStatus), "Title: " + title + "_"}
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

func planCalendarWindow(startDate time.Time, dayCount int, targetDate time.Time) (time.Time, int) {
	if dayCount <= 0 {
		return startDate, 0
	}
	start := normalizeDate(startDate)
	target := normalizeDate(targetDate)
	end := start.AddDate(0, 0, dayCount-1)
	if !target.Before(start) && !target.After(end) {
		return start, int(target.Sub(start).Hours() / 24)
	}
	if target.Before(start) {
		return target, 0
	}
	return target.AddDate(0, 0, -(dayCount - 1)), dayCount - 1
}

func normalizeDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func sameDate(a time.Time, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
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
	if len(runes) <= width {
		return value
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}
