package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/kyaoi/gitshelf/internal/interactive"
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

type calendarKindTarget int

const (
	calendarKindTargetTask calendarKindTarget = iota
	calendarKindTargetAdd
)

type calendarAddField int

const (
	calendarAddFieldTitle calendarAddField = iota
	calendarAddFieldKind
)

type calendarTextPromptPurpose int

const (
	calendarTextPromptNone calendarTextPromptPurpose = iota
	calendarTextPromptTag
)

type calendarTagBulkState int

const (
	calendarTagBulkStateUnchanged calendarTagBulkState = iota
	calendarTagBulkStateAdd
	calendarTagBulkStateRemove
)

type calendarFilterSection int

const (
	calendarFilterIncludeStatuses calendarFilterSection = iota
	calendarFilterExcludeStatuses
	calendarFilterIncludeKinds
	calendarFilterExcludeKinds
)

type calendarLinkAction int

const (
	calendarLinkActionAdd calendarLinkAction = iota
	calendarLinkActionRemove
)

type calendarLinkCandidate struct {
	TaskID      string
	Title       string
	Label       string
	Path        string
	Type        shelf.LinkType
	Parent      string
	HasChildren bool
	Collapsed   bool
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
	Task        shelf.Task
	ParentTitle string
	Subtitle    string
	Reason      string
}

type calendarSection struct {
	ID    calendarSectionID
	Title string
	Items []calendarSectionItem
}

type cockpitTreeRow struct {
	Task         shelf.Task
	Label        string
	Meta         string
	DueOn        string
	DueInherited bool
	HasChildren  bool
	Collapsed    bool
}

type calendarTUIModel struct {
	rootDir       string
	mode          calendarMode
	startDate     time.Time
	daysCount     int
	statuses      []shelf.Status
	statusChoices []shelf.Status
	sectionLimit  int
	filter        shelf.TaskFilter
	defaultKind   shelf.Kind
	defaultStatus shelf.Status
	kindChoices   []shelf.Kind
	tagChoices    []string
	linkChoices   []shelf.LinkType
	blockingLink  shelf.LinkType
	copySeparator string
	copyPresets   []shelf.CopyPreset
	showID        bool

	visibleTasks  []shelf.Task
	allTasks      []shelf.Task
	readiness     map[string]shelf.TaskReadiness
	taskByID      map[string]shelf.Task
	titleByID     map[string]string
	effectiveDue  map[string]string
	outboundCount map[string]int
	inboundCount  map[string]int

	days                     []calendarDay
	dayIndex                 int
	sections                 []calendarSection
	sectionIndex             int
	sectionRows              map[calendarSectionID]int
	treeRows                 []cockpitTreeRow
	treeRowIndex             int
	collapsedTree            map[string]struct{}
	boardColumns             []boardColumn
	boardColumnIdx           int
	boardRowIndex            map[int]int
	selectedTaskID           string
	markedTaskIDs            map[string]struct{}
	rangeMarkMode            bool
	rangeAnchorID            string
	rangeBaseIDs             map[string]struct{}
	moveMode                 bool
	moveSourceIDs            []string
	pane                     calendarPane
	width                    int
	height                   int
	bodyScroll               int
	message                  string
	showHelp                 bool
	showTaskBody             bool
	snoozeMode               bool
	snoozeIndex              int
	linkMode                 bool
	linkAction               calendarLinkAction
	linkIndex                int
	linkQuery                string
	linkQueryCursor          int
	linkQueryMode            bool
	linkTypeIndex            int
	linkCollapsedTree        map[string]struct{}
	copyPresetMode           bool
	copyPresetIndex          int
	copyPresetFocus          calendarCopyPresetFocus
	copyPresetName           string
	copyPresetNameCursor     int
	copyPresetScope          shelf.CopyPresetScope
	copyPresetSubtreeStyle   shelf.CopySubtreeStyle
	copyPresetTemplate       string
	copyPresetTemplateCursor int
	copyPresetJoinWith       string
	copyPresetJoinWithCursor int
	filterMode               bool
	filterSection            calendarFilterSection
	filterIndex              int
	filterSnapshot           shelf.TaskFilter
	kindMode                 bool
	kindIndex                int
	kindTarget               calendarKindTarget
	tagMode                  bool
	tagIndex                 int
	tagSelection             []string
	tagBulkMode              bool
	tagBulkStates            map[string]calendarTagBulkState
	tagInputMode             bool
	tagInputValue            string
	tagInputCursor           int
	textPromptMode           bool
	textPromptTitle          string
	textPromptValue          string
	textPromptCursor         int
	textPromptPurpose        calendarTextPromptPurpose
	addMode                  bool
	addTitle                 string
	addTitleCursor           int
	addKind                  shelf.Kind
	addField                 calendarAddField
	addAtRoot                bool
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
		rootDir:           rootDir,
		mode:              opts.Mode,
		startDate:         startDate,
		daysCount:         daysCount,
		sectionLimit:      opts.SectionLimit,
		defaultKind:       cfg.DefaultKind,
		defaultStatus:     cfg.DefaultStatus,
		kindChoices:       append([]shelf.Kind{}, cfg.Kinds...),
		statusChoices:     append([]shelf.Status{}, cfg.Statuses...),
		tagChoices:        append([]string{}, cfg.Tags...),
		linkChoices:       append([]shelf.LinkType{}, cfg.LinkTypes.Names...),
		blockingLink:      cfg.BlockingLinkType(),
		copySeparator:     cfg.Commands.Cockpit.CopySeparator,
		copyPresets:       append([]shelf.CopyPreset{}, cfg.Commands.Cockpit.CopyPresets...),
		showID:            opts.ShowID,
		pane:              calendarPaneMain,
		sectionRows:       map[calendarSectionID]int{},
		boardRowIndex:     map[int]int{},
		markedTaskIDs:     map[string]struct{}{},
		rangeBaseIDs:      map[string]struct{}{},
		collapsedTree:     map[string]struct{}{},
		linkCollapsedTree: map[string]struct{}{},
		readiness:         map[string]shelf.TaskReadiness{},
		taskByID:          map[string]shelf.Task{},
		titleByID:         map[string]string{},
		effectiveDue:      map[string]string{},
		outboundCount:     map[string]int{},
		inboundCount:      map[string]int{},
		addKind:           cfg.DefaultKind,
	}
	model.filter = opts.Filter
	model.filter.Limit = 0
	if len(model.filter.Statuses) == 0 && len(statuses) > 0 {
		model.filter.Statuses = append([]shelf.Status{}, statuses...)
	}
	model.statuses = deriveActiveStatuses(model.statusChoices, model.filter)
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
		if m.linkMode {
			return m.updateLinkMode(msg)
		}
		if m.copyPresetMode {
			return m.updateCopyPresetMode(msg)
		}
		if m.filterMode {
			return m.updateFilterMode(msg)
		}
		if m.tagMode {
			return m.updateTagMode(msg)
		}
		if m.kindMode {
			return m.updateKindMode(msg)
		}
		if m.addMode {
			return m.updateAddMode(msg)
		}
		if m.snoozeMode {
			return m.updateSnoozeMode(msg)
		}
		if m.moveMode {
			switch msg.String() {
			case "esc", "q", "m", "ctrl+[":
				m.moveMode = false
				m.moveSourceIDs = nil
				m.message = "move をキャンセルしました"
				return m, nil
			case "enter":
				return m.applyMoveSelection()
			}
		}
		if m.showHelp && (msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+[") {
			m.showHelp = false
			return m, nil
		}
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "tab":
			if m.mode != calendarModeCalendar {
				m.nextPane()
			}
			return m, nil
		case "shift+tab":
			if m.mode != calendarModeCalendar {
				m.prevPane()
			}
			return m, nil
		case "ctrl+h":
			m.cycleMode(-1)
			return m, nil
		case "ctrl+l":
			m.cycleMode(1)
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
		case "t":
			m.moveFocusToDate(time.Now().Local())
			m.message = "jumped to today"
			return m, nil
		case "left", "h":
			if m.usesSidebarCalendarNav() {
				m.moveSidebarFocusByDays(-1)
			} else if m.mode == calendarModeCalendar {
				m.moveFocusByDays(-1)
			} else if m.mode == calendarModeTree {
				m.collapseCurrentTreeNode()
			} else if m.mode == calendarModeBoard {
				m.moveBoardColumn(-1)
			} else {
				m.moveSectionSelection(-1)
			}
			return m, nil
		case "right", "l":
			if m.usesSidebarCalendarNav() {
				m.moveSidebarFocusByDays(1)
			} else if m.mode == calendarModeCalendar {
				m.moveFocusByDays(1)
			} else if m.mode == calendarModeTree {
				m.expandCurrentTreeNode()
			} else if m.mode == calendarModeBoard {
				m.moveBoardColumn(1)
			} else {
				m.moveSectionSelection(1)
			}
			return m, nil
		case "up":
			if m.usesSidebarCalendarNav() {
				m.moveSidebarFocusByDays(-7)
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
				m.moveSidebarFocusByDays(7)
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
				m.moveSidebarFocusByDays(-7)
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
				m.moveSidebarFocusByDays(7)
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
		case "[":
			if m.mode == calendarModeCalendar || m.usesSidebarCalendarNav() {
				if m.usesSidebarCalendarNav() {
					m.moveSidebarFocusByMonths(-1)
				} else {
					m.moveFocusByMonths(-1)
				}
			}
			return m, nil
		case "]":
			if m.mode == calendarModeCalendar || m.usesSidebarCalendarNav() {
				if m.usesSidebarCalendarNav() {
					m.moveSidebarFocusByMonths(1)
				} else {
					m.moveFocusByMonths(1)
				}
			}
			return m, nil
		case "n":
			m.moveSelectedDayTask(1)
			return m, nil
		case "p":
			m.moveSelectedDayTask(-1)
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
		case "enter":
			if _, ok := m.selectedTask(); ok {
				m.showTaskBody = !m.showTaskBody
			}
			return m, nil
		case "V":
			if m.mode == calendarModeTree || m.mode == calendarModeBoard {
				m.toggleRangeSelectionMode()
				return m, nil
			}
			return m, nil
		case "u":
			if m.mode == calendarModeTree || m.mode == calendarModeBoard {
				m.clearMarkedSelection()
				m.message = "cleared all marks"
				return m, nil
			}
			return m, nil
		case "v":
			if m.mode == calendarModeTree || m.mode == calendarModeBoard {
				if m.rangeMarkMode {
					m.rangeMarkMode = false
					m.rangeAnchorID = ""
					m.rangeBaseIDs = map[string]struct{}{}
				}
				m.toggleMarkedSelection()
				return m, nil
			}
			if _, ok := m.selectedTask(); ok {
				m.showTaskBody = !m.showTaskBody
			}
			return m, nil
		case "e", "E":
			return m.openEditorForSelectedTask()
		case "y":
			if err := m.copySelectedTaskTitles(); err != nil {
				m.message = err.Error()
			}
			return m, nil
		case "Y":
			if err := m.copySelectedTaskSubtrees(); err != nil {
				m.message = err.Error()
			}
			return m, nil
		case "P":
			if err := m.copySelectedTaskPaths(); err != nil {
				m.message = err.Error()
			}
			return m, nil
		case "O":
			if err := m.copySelectedTaskBodies(); err != nil {
				m.message = err.Error()
			}
			return m, nil
		case "M":
			m.beginCopyPresetMode()
			return m, nil
		case "L":
			m.beginLinkMode(calendarLinkActionAdd)
			return m, nil
		case "U":
			m.beginLinkMode(calendarLinkActionRemove)
			return m, nil
		case "a":
			if _, ok := m.selectedTask(); !ok {
				m.beginAddMode(true)
				return m, nil
			}
			m.beginAddMode(false)
			return m, nil
		case "A":
			m.beginAddMode(true)
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
		case "K":
			m.beginKindMode(calendarKindTargetTask)
			return m, nil
		case "#":
			m.beginTagMode()
			return m, nil
		case "f":
			m.beginFilterMode()
			return m, nil
		case "m":
			return m.beginMoveSelection()
		case "r":
			if err := m.reload(); err != nil {
				m.message = err.Error()
			} else {
				m.message = "reloaded"
			}
			return m, nil
		case "x":
			return m.toggleArchivedState()
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
	topParts := []string{
		renderCockpitHeader(m, focused),
		renderCalendarModeTabs(m.mode, max(20, m.width-2)),
	}
	bottomParts := make([]string, 0, 1)
	if strings.TrimSpace(m.message) != "" {
		bottomParts = append(bottomParts, lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Render(m.message))
	}
	bodyHeight := availableViewportHeight(m.height, strings.Join(topParts, "\n\n"), strings.Join(bottomParts, "\n\n"))
	month := buildCalendarMonthView(m.days, focused)
	mainWidth, gapWidth, inspectorWidth := m.layoutColumns()

	main := renderCalendarMainPane(m, month, mainWidth, bodyHeight, m.pane == calendarPaneMain)
	selectedTask, selectedTaskOK := m.selectedTask()
	right := renderCalendarInspectorPane(selectedTask, selectedTaskOK, m.showID, m.showTaskBody, m.taskByID, m.readiness, m.outboundCount, m.inboundCount, m.blockingLink, inspectorWidth, bodyHeight)
	if m.mode == calendarModeCalendar {
		right = renderCalendarSidebarPane(m, selectedTask, selectedTaskOK, inspectorWidth, bodyHeight)
	} else {
		right = renderCalendarSecondarySidebarPane(m, selectedTask, selectedTaskOK, inspectorWidth, bodyHeight, m.pane == calendarPaneInspector)
	}

	gap := lipgloss.NewStyle().Width(gapWidth).Render("")
	body := lipgloss.JoinHorizontal(lipgloss.Top, main, gap, right)
	if popup := m.activePopup(); strings.TrimSpace(popup) != "" {
		body = renderCenteredPopupCanvas(popup, max(48, m.width-2), max(12, bodyHeight))
	}
	return renderCalendarViewport(topParts, body, bottomParts, m.height, m.bodyScroll) + "\n"
}

func (m calendarTUIModel) activePopup() string {
	popupWidth, popupHeight := m.popupDimensions()
	switch {
	case m.showHelp:
		return renderCockpitHelpOverlay(m.mode, popupWidth, popupHeight)
	case m.filterMode:
		return renderCalendarFilterPicker(m, popupWidth, popupHeight)
	case m.snoozeMode:
		return renderCalendarSnoozePicker(m.bulkActionPopupLabel(), m.snoozeIndex, popupWidth, popupHeight)
	case m.linkMode:
		query := m.linkQuery
		if m.linkQueryMode {
			query = interactive.RenderTextCursor(query, m.linkQueryCursor)
		}
		return renderCalendarLinkPicker(m.linkAction, m.linkTypeLabel(), query, m.linkQueryMode, m.selectedTaskPopupLabel(), m.currentLinkCandidates(), m.linkIndex, popupWidth, popupHeight)
	case m.copyPresetMode:
		return renderCalendarCopyPresetPopup(m, popupWidth, popupHeight)
	case m.kindMode:
		return renderCalendarKindPicker(m.bulkActionPopupLabel(), m.kindChoices, m.kindIndex, popupWidth, popupHeight)
	case m.tagMode:
		inputValue := m.tagInputValue
		if m.tagInputMode {
			inputValue = interactive.RenderTextCursor(inputValue, m.tagInputCursor)
		}
		return renderCalendarTagPicker(m.tagPopupLabel(), m.tagChoices, m.tagSelection, m.tagBulkMode, m.tagBulkStates, m.tagIndex, m.tagInputMode, inputValue, popupWidth, popupHeight)
	case m.textPromptMode:
		return renderCalendarTextPrompt(m.textPromptTitle, interactive.RenderTextCursor(m.textPromptValue, m.textPromptCursor))
	case m.addMode:
		title := m.addTitle
		if m.addField == calendarAddFieldTitle {
			title = interactive.RenderTextCursor(title, m.addTitleCursor)
		}
		return renderCalendarAddComposer(m.focusedDayLabel(), m.addKind, m.defaultStatus, title, m.addField, m.addTargetLabel(), m.addAtRoot, popupWidth, popupHeight)
	default:
		return ""
	}
}

func (m calendarTUIModel) popupDimensions() (int, int) {
	width := max(76, min(112, m.width*78/100))
	height := max(18, min(28, m.height*78/100))
	if m.width <= 0 {
		width = 92
	}
	if m.height <= 0 {
		height = 24
	}
	return width, height
}

func renderCenteredPopupCanvas(popup string, width int, height int) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, popup)
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

func (m *calendarTUIModel) leaveToNormalMode() bool {
	changed := false
	if m.addMode {
		m.addMode = false
		m.addTitle = ""
		m.addTitleCursor = 0
		m.addKind = m.defaultKind
		m.addField = calendarAddFieldTitle
		m.addAtRoot = false
		changed = true
	}
	if m.snoozeMode {
		m.snoozeMode = false
		changed = true
	}
	if m.linkMode {
		m.linkMode = false
		m.linkQuery = ""
		m.linkQueryCursor = 0
		m.linkQueryMode = false
		changed = true
	}
	if m.copyPresetMode {
		m.copyPresetMode = false
		m.copyPresetIndex = 0
		m.copyPresetFocus = calendarCopyPresetFocusPresets
		m.copyPresetName = ""
		m.copyPresetNameCursor = 0
		m.copyPresetScope = ""
		m.copyPresetSubtreeStyle = ""
		m.copyPresetTemplate = ""
		m.copyPresetTemplateCursor = 0
		m.copyPresetJoinWith = ""
		m.copyPresetJoinWithCursor = 0
		changed = true
	}
	if m.filterMode {
		m.filterMode = false
		changed = true
	}
	if m.kindMode {
		m.kindMode = false
		changed = true
	}
	if m.tagMode {
		m.clearTagModeState()
		changed = true
	}
	if m.textPromptMode {
		m.textPromptMode = false
		m.textPromptTitle = ""
		m.textPromptValue = ""
		m.textPromptCursor = 0
		m.textPromptPurpose = calendarTextPromptNone
		changed = true
	}
	if m.moveMode {
		m.moveMode = false
		m.moveSourceIDs = nil
		changed = true
	}
	if m.rangeMarkMode {
		m.rangeMarkMode = false
		m.rangeAnchorID = ""
		m.rangeBaseIDs = map[string]struct{}{}
		changed = true
	}
	if m.showHelp {
		m.showHelp = false
		changed = true
	}
	if changed {
		m.addTitle = ""
		m.addTitleCursor = 0
		m.message = "normal mode"
	}
	return changed
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
	case "esc", "q", "ctrl+[":
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

func (m *calendarTUIModel) beginFilterMode() {
	m.filterMode = true
	m.filterSection = calendarFilterIncludeStatuses
	m.filterIndex = 0
	m.filterSnapshot = cloneTaskFilter(m.filter)
	m.message = "filter を編集"
}

func (m calendarTUIModel) updateFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "ctrl+[":
		m.filter = cloneTaskFilter(m.filterSnapshot)
		m.statuses = deriveActiveStatuses(m.statusChoices, m.filter)
		m.filterMode = false
		m.message = "filter 編集をキャンセルしました"
		return m, nil
	case "left", "h":
		if m.filterSection > calendarFilterIncludeStatuses {
			m.filterSection--
		}
		m.clampFilterSelection()
		return m, nil
	case "right", "l", "tab":
		if m.filterSection < calendarFilterExcludeKinds {
			m.filterSection++
		} else {
			m.filterSection = calendarFilterIncludeStatuses
		}
		m.clampFilterSelection()
		return m, nil
	case "up", "k":
		if m.filterIndex > 0 {
			m.filterIndex--
		}
		return m, nil
	case "down", "j":
		if m.filterIndex < len(m.currentFilterOptions())-1 {
			m.filterIndex++
		}
		return m, nil
	case " ":
		if err := m.toggleCurrentFilterOption(); err != nil {
			m.message = err.Error()
			return m, nil
		}
		return m, nil
	case "enter":
		if err := m.reload(); err != nil {
			m.message = err.Error()
			return m, nil
		}
		m.filterMode = false
		m.message = "filter を更新"
		return m, nil
	}
	return m, nil
}

func cloneTaskFilter(filter shelf.TaskFilter) shelf.TaskFilter {
	clone := filter
	clone.Kinds = append([]shelf.Kind{}, filter.Kinds...)
	clone.Statuses = append([]shelf.Status{}, filter.Statuses...)
	clone.Tags = append([]string{}, filter.Tags...)
	clone.NotKinds = append([]shelf.Kind{}, filter.NotKinds...)
	clone.NotStatuses = append([]shelf.Status{}, filter.NotStatuses...)
	clone.NotTags = append([]string{}, filter.NotTags...)
	return clone
}

func (m *calendarTUIModel) clampFilterSelection() {
	options := m.currentFilterOptions()
	if len(options) == 0 {
		m.filterIndex = 0
		return
	}
	if m.filterIndex < 0 {
		m.filterIndex = 0
	}
	if m.filterIndex >= len(options) {
		m.filterIndex = len(options) - 1
	}
}

func (m calendarTUIModel) currentFilterOptions() []string {
	return m.filterOptionsForSection(m.filterSection)
}

func (m *calendarTUIModel) toggleCurrentFilterOption() error {
	options := m.currentFilterOptions()
	if len(options) == 0 || m.filterIndex < 0 || m.filterIndex >= len(options) {
		return nil
	}
	value := options[m.filterIndex]
	switch m.filterSection {
	case calendarFilterIncludeStatuses:
		m.filter.Statuses = toggleStatusFilter(m.filter.Statuses, shelf.Status(value))
	case calendarFilterExcludeStatuses:
		m.filter.NotStatuses = toggleStatusFilter(m.filter.NotStatuses, shelf.Status(value))
	case calendarFilterIncludeKinds:
		m.filter.Kinds = toggleKindFilter(m.filter.Kinds, shelf.Kind(value))
	case calendarFilterExcludeKinds:
		m.filter.NotKinds = toggleKindFilter(m.filter.NotKinds, shelf.Kind(value))
	}
	m.statuses = deriveActiveStatuses(m.statusChoices, m.filter)
	return nil
}

func toggleStatusFilter(values []shelf.Status, target shelf.Status) []shelf.Status {
	if slices.Contains(values, target) {
		return removeStatusFilter(values, target)
	}
	values = append(values, target)
	slices.Sort(values)
	return values
}

func removeStatusFilter(values []shelf.Status, target shelf.Status) []shelf.Status {
	filtered := make([]shelf.Status, 0, len(values))
	for _, value := range values {
		if value == target {
			continue
		}
		filtered = append(filtered, value)
	}
	return filtered
}

func toggleKindFilter(values []shelf.Kind, target shelf.Kind) []shelf.Kind {
	if slices.Contains(values, target) {
		filtered := make([]shelf.Kind, 0, len(values))
		for _, value := range values {
			if value == target {
				continue
			}
			filtered = append(filtered, value)
		}
		return filtered
	}
	values = append(values, target)
	slices.Sort(values)
	return values
}

func (m calendarTUIModel) updateAddMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+[":
		m.addMode = false
		m.addTitle = ""
		m.addTitleCursor = 0
		m.addKind = m.defaultKind
		m.addField = calendarAddFieldTitle
		m.addAtRoot = false
		m.message = "新規 task 作成をキャンセルしました"
		return m, nil
	case "tab":
		if m.addField == calendarAddFieldTitle {
			m.addField = calendarAddFieldKind
		} else {
			m.addField = calendarAddFieldTitle
		}
		return m, nil
	case "shift+tab":
		if m.addField == calendarAddFieldKind {
			m.addField = calendarAddFieldTitle
		} else {
			m.addField = calendarAddFieldKind
		}
		return m, nil
	case "enter":
		title := strings.TrimSpace(m.addTitle)
		if title == "" {
			m.message = "title は必須です"
			return m, nil
		}
		if err := m.createTaskFromAddMode(title); err != nil {
			m.message = err.Error()
			return m, nil
		}
		m.addMode = false
		m.addTitle = ""
		m.addTitleCursor = 0
		m.addKind = m.defaultKind
		m.addField = calendarAddFieldTitle
		m.addAtRoot = false
		return m, nil
	case "up":
		if m.addField == calendarAddFieldKind {
			m.addKind = m.prevAddKind()
		}
		return m, nil
	case "down":
		if m.addField == calendarAddFieldKind {
			m.addKind = m.nextAddKind()
		}
		return m, nil
	case "backspace":
		if m.addField == calendarAddFieldTitle {
			m.addTitle, m.addTitleCursor = interactive.DeleteRuneBeforeCursor(m.addTitle, m.addTitleCursor)
		}
		return m, nil
	case "left":
		if m.addField == calendarAddFieldTitle {
			m.addTitleCursor = interactive.MoveTextCursorLeft(m.addTitle, m.addTitleCursor)
		}
		return m, nil
	case "right":
		if m.addField == calendarAddFieldTitle {
			m.addTitleCursor = interactive.MoveTextCursorRight(m.addTitle, m.addTitleCursor)
		}
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			if m.addField == calendarAddFieldKind {
				switch msg.String() {
				case "k":
					m.addKind = m.prevAddKind()
					return m, nil
				case "j":
					m.addKind = m.nextAddKind()
					return m, nil
				}
			}
			if m.addField == calendarAddFieldTitle {
				for _, r := range msg.Runes {
					m.addTitle, m.addTitleCursor = interactive.InsertRuneAtCursor(m.addTitle, m.addTitleCursor, r)
				}
				return m, nil
			}
		}
	}
	return m, nil
}

func (m calendarTUIModel) nextAddKind() shelf.Kind {
	if len(m.kindChoices) == 0 {
		return m.addKind
	}
	index := 0
	for i, kind := range m.kindChoices {
		if kind == m.addKind {
			index = i
			break
		}
	}
	return m.kindChoices[(index+1)%len(m.kindChoices)]
}

func (m calendarTUIModel) prevAddKind() shelf.Kind {
	if len(m.kindChoices) == 0 {
		return m.addKind
	}
	index := 0
	for i, kind := range m.kindChoices {
		if kind == m.addKind {
			index = i
			break
		}
	}
	index--
	if index < 0 {
		index = len(m.kindChoices) - 1
	}
	return m.kindChoices[index]
}

func (m calendarTUIModel) addTargetLabel() string {
	if m.addAtRoot {
		return "Root"
	}
	task, ok := m.selectedTask()
	if !ok {
		return "Root"
	}
	return buildTaskPath(task, m.taskByID)
}

func (m calendarTUIModel) selectedTaskPopupLabel() string {
	task, ok := m.selectedTask()
	if !ok {
		return "none"
	}
	label := task.Title
	if m.showID {
		label = fmt.Sprintf("[%s] %s", shelf.ShortID(task.ID), label)
	}
	return label
}

func (m calendarTUIModel) bulkActionPopupLabel() string {
	if marked := m.markedCount(); marked > 0 {
		return fmt.Sprintf("%d marked tasks", marked)
	}
	return m.selectedTaskPopupLabel()
}

func (m calendarTUIModel) tagPopupLabel() string {
	if m.tagBulkMode {
		return m.bulkActionPopupLabel()
	}
	return m.selectedTaskPopupLabel()
}

func (m *calendarTUIModel) clearTagModeState() {
	m.tagMode = false
	m.tagSelection = nil
	m.tagBulkMode = false
	m.tagBulkStates = nil
	m.tagInputMode = false
	m.tagInputValue = ""
	m.tagInputCursor = 0
}

func (m calendarTUIModel) tagBulkState(tag string) calendarTagBulkState {
	if m.tagBulkStates == nil {
		return calendarTagBulkStateUnchanged
	}
	state, ok := m.tagBulkStates[tag]
	if !ok {
		return calendarTagBulkStateUnchanged
	}
	return state
}

func (m *calendarTUIModel) setTagBulkState(tag string, state calendarTagBulkState) {
	if state == calendarTagBulkStateUnchanged {
		if m.tagBulkStates != nil {
			delete(m.tagBulkStates, tag)
		}
		return
	}
	if m.tagBulkStates == nil {
		m.tagBulkStates = map[string]calendarTagBulkState{}
	}
	m.tagBulkStates[tag] = state
}

func nextCalendarTagBulkState(state calendarTagBulkState) calendarTagBulkState {
	switch state {
	case calendarTagBulkStateUnchanged:
		return calendarTagBulkStateAdd
	case calendarTagBulkStateAdd:
		return calendarTagBulkStateRemove
	default:
		return calendarTagBulkStateUnchanged
	}
}

func markerForCalendarTagBulkState(state calendarTagBulkState) string {
	switch state {
	case calendarTagBulkStateAdd:
		return "[+]"
	case calendarTagBulkStateRemove:
		return "[-]"
	default:
		return "[ ]"
	}
}

func (m *calendarTUIModel) beginKindMode(target calendarKindTarget) {
	if len(m.kindChoices) == 0 {
		m.message = "kind が設定されていません"
		return
	}
	current := m.defaultKind
	if target == calendarKindTargetAdd {
		current = m.addKind
	} else if task, ok := m.selectedTask(); ok {
		current = task.Kind
	} else {
		m.message = "選択中の task がありません"
		return
	}
	m.kindTarget = target
	m.kindIndex = 0
	for i, kind := range m.kindChoices {
		if kind == current {
			m.kindIndex = i
			break
		}
	}
	m.kindMode = true
	if target == calendarKindTargetAdd {
		m.message = "新規 task の kind を選択"
		return
	}
	m.message = "kind を選択"
}

func (m calendarTUIModel) updateKindMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "ctrl+[":
		m.kindMode = false
		m.message = "kind 変更をキャンセルしました"
		return m, nil
	case "up", "k":
		if m.kindIndex > 0 {
			m.kindIndex--
		}
		return m, nil
	case "down", "j":
		if m.kindIndex < len(m.kindChoices)-1 {
			m.kindIndex++
		}
		return m, nil
	case "enter":
		if len(m.kindChoices) == 0 {
			m.kindMode = false
			return m, nil
		}
		selected := m.kindChoices[m.kindIndex]
		if m.kindTarget == calendarKindTargetAdd {
			m.addKind = selected
			m.kindMode = false
			m.message = fmt.Sprintf("新規 task の kind を %s に設定", selected)
			return m, nil
		}
		if err := m.applySelectedTaskKind(selected); err != nil {
			m.kindMode = false
			m.message = err.Error()
			return m, nil
		}
		m.kindMode = false
		return m, nil
	}
	return m, nil
}

func (m *calendarTUIModel) beginTagMode() {
	task, ok := m.selectedTask()
	if !ok {
		m.message = "選択中の task がありません"
		return
	}
	choices := append([]string{}, m.tagChoices...)
	for _, tag := range task.Tags {
		if containsTag(choices, tag) {
			continue
		}
		choices = append(choices, tag)
	}
	m.tagChoices = shelf.NormalizeTags(choices)
	if m.markedCount() > 0 {
		for _, taskID := range m.activeTaskIDs() {
			markedTask, ok := m.taskByID[taskID]
			if !ok {
				continue
			}
			for _, tag := range markedTask.Tags {
				if containsTag(m.tagChoices, tag) {
					continue
				}
				m.tagChoices = append(m.tagChoices, tag)
			}
		}
		m.tagChoices = shelf.NormalizeTags(m.tagChoices)
		m.tagSelection = nil
		m.tagBulkMode = true
		m.tagBulkStates = map[string]calendarTagBulkState{}
	} else {
		m.tagSelection = append([]string{}, task.Tags...)
		m.tagBulkMode = false
		m.tagBulkStates = nil
	}
	m.tagIndex = 0
	m.tagMode = true
	m.tagInputMode = false
	m.tagInputValue = ""
	m.tagInputCursor = 0
	if m.tagBulkMode {
		m.message = "tag を一括編集"
		return
	}
	m.message = "tag を編集"
}

func (m calendarTUIModel) updateTagMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.tagInputMode {
		switch msg.String() {
		case "esc", "ctrl+[":
			m.tagInputMode = false
			m.tagInputValue = ""
			m.tagInputCursor = 0
			m.message = "tag 入力をキャンセルしました"
			return m, nil
		case "enter":
			value := strings.TrimSpace(m.tagInputValue)
			if value == "" {
				m.message = "tag は必須です"
				return m, nil
			}
			if !containsTag(m.tagChoices, value) {
				m.tagChoices = append(m.tagChoices, value)
				m.tagChoices = shelf.NormalizeTags(m.tagChoices)
			}
			if m.tagBulkMode {
				m.setTagBulkState(value, calendarTagBulkStateAdd)
			} else if !containsTag(m.tagSelection, value) {
				m.tagSelection = append(m.tagSelection, value)
				m.tagSelection = shelf.NormalizeTags(m.tagSelection)
			}
			m.tagIndex = 2
			for i, tag := range m.tagChoices {
				if tag == value {
					m.tagIndex = i + 2
					break
				}
			}
			m.tagInputMode = false
			m.tagInputValue = ""
			m.tagInputCursor = 0
			m.message = fmt.Sprintf("tag %s を追加", value)
			return m, nil
		case "backspace":
			m.tagInputValue, m.tagInputCursor = interactive.DeleteRuneBeforeCursor(m.tagInputValue, m.tagInputCursor)
			return m, nil
		case "left":
			m.tagInputCursor = interactive.MoveTextCursorLeft(m.tagInputValue, m.tagInputCursor)
			return m, nil
		case "right":
			m.tagInputCursor = interactive.MoveTextCursorRight(m.tagInputValue, m.tagInputCursor)
			return m, nil
		default:
			if msg.Type == tea.KeyRunes {
				for _, r := range msg.Runes {
					m.tagInputValue, m.tagInputCursor = interactive.InsertRuneAtCursor(m.tagInputValue, m.tagInputCursor, r)
				}
				return m, nil
			}
		}
		return m, nil
	}

	optionCount := len(m.tagChoices) + 2
	switch msg.String() {
	case "ctrl+s":
		if m.tagBulkMode {
			if err := m.applySelectedTaskTagDelta(m.currentTagBulkDelta()); err != nil {
				m.clearTagModeState()
				m.message = err.Error()
				return m, nil
			}
			m.clearTagModeState()
			return m, nil
		}
		if err := m.applySelectedTaskTags(append([]string{}, m.tagSelection...)); err != nil {
			m.clearTagModeState()
			m.message = err.Error()
			return m, nil
		}
		m.clearTagModeState()
		return m, nil
	case "esc", "q", "ctrl+[":
		m.clearTagModeState()
		m.message = "tag 変更をキャンセルしました"
		return m, nil
	case "up", "k":
		if m.tagIndex > 0 {
			m.tagIndex--
		}
		return m, nil
	case "down", "j":
		if m.tagIndex < optionCount-1 {
			m.tagIndex++
		}
		return m, nil
	case " ":
		if m.tagIndex < 2 {
			return m, nil
		}
		tag := m.tagChoices[m.tagIndex-2]
		if m.tagBulkMode {
			m.setTagBulkState(tag, nextCalendarTagBulkState(m.tagBulkState(tag)))
			return m, nil
		}
		if containsTag(m.tagSelection, tag) {
			m.tagSelection = removeTag(m.tagSelection, tag)
		} else {
			m.tagSelection = append(m.tagSelection, tag)
			m.tagSelection = shelf.NormalizeTags(m.tagSelection)
		}
		return m, nil
	case "enter":
		switch {
		case m.tagIndex == 0:
			if m.tagBulkMode {
				if err := m.applySelectedTaskTagDelta(m.currentTagBulkDelta()); err != nil {
					m.clearTagModeState()
					m.message = err.Error()
					return m, nil
				}
				m.clearTagModeState()
				return m, nil
			}
			if err := m.applySelectedTaskTags(append([]string{}, m.tagSelection...)); err != nil {
				m.clearTagModeState()
				m.message = err.Error()
				return m, nil
			}
			m.clearTagModeState()
			return m, nil
		case m.tagIndex == 1:
			m.tagInputMode = true
			m.tagInputValue = ""
			m.tagInputCursor = 0
			m.message = "新しい tag を入力"
			return m, nil
		}
	}
	return m, nil
}

func (m calendarTUIModel) updateTextPromptMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "ctrl+[":
		m.textPromptMode = false
		m.textPromptTitle = ""
		m.textPromptValue = ""
		m.textPromptCursor = 0
		m.textPromptPurpose = calendarTextPromptNone
		m.message = "入力をキャンセルしました"
		return m, nil
	case "enter":
		value := strings.TrimSpace(m.textPromptValue)
		if value == "" {
			m.message = "入力は必須です"
			return m, nil
		}
		switch m.textPromptPurpose {
		case calendarTextPromptTag:
			if !containsTag(m.tagChoices, value) {
				m.tagChoices = append(m.tagChoices, value)
				m.tagChoices = shelf.NormalizeTags(m.tagChoices)
			}
			if !containsTag(m.tagSelection, value) {
				m.tagSelection = append(m.tagSelection, value)
				m.tagSelection = shelf.NormalizeTags(m.tagSelection)
			}
			m.tagIndex = 2
			for i, tag := range m.tagChoices {
				if tag == value {
					m.tagIndex = i + 2
					break
				}
			}
			m.message = fmt.Sprintf("tag %s を追加", value)
		}
		m.textPromptMode = false
		m.textPromptTitle = ""
		m.textPromptValue = ""
		m.textPromptCursor = 0
		m.textPromptPurpose = calendarTextPromptNone
		return m, nil
	case "backspace":
		m.textPromptValue, m.textPromptCursor = interactive.DeleteRuneBeforeCursor(m.textPromptValue, m.textPromptCursor)
		return m, nil
	case "left":
		m.textPromptCursor = interactive.MoveTextCursorLeft(m.textPromptValue, m.textPromptCursor)
		return m, nil
	case "right":
		m.textPromptCursor = interactive.MoveTextCursorRight(m.textPromptValue, m.textPromptCursor)
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			for _, r := range msg.Runes {
				m.textPromptValue, m.textPromptCursor = interactive.InsertRuneAtCursor(m.textPromptValue, m.textPromptCursor, r)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *calendarTUIModel) reload() error {
	selectedTaskID := m.selectedTaskID
	cfg, err := shelf.LoadConfig(m.rootDir)
	if err != nil {
		return err
	}
	m.defaultKind = cfg.DefaultKind
	m.defaultStatus = cfg.DefaultStatus
	m.kindChoices = append([]shelf.Kind{}, cfg.Kinds...)
	m.statusChoices = append([]shelf.Status{}, cfg.Statuses...)
	m.tagChoices = append([]string{}, cfg.Tags...)
	m.linkChoices = append([]shelf.LinkType{}, cfg.LinkTypes.Names...)
	m.blockingLink = cfg.BlockingLinkType()
	m.copySeparator = cfg.Commands.Cockpit.CopySeparator
	m.copyPresets = append([]shelf.CopyPreset{}, cfg.Commands.Cockpit.CopyPresets...)
	m.statuses = deriveActiveStatuses(m.statusChoices, m.filter)
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
	m.effectiveDue = buildEffectiveDueMap(allTasks)
	m.outboundCount = outboundCount
	m.inboundCount = inboundCount
	m.taskByID = make(map[string]shelf.Task, len(allTasks))
	m.titleByID = make(map[string]string, len(allTasks))
	for _, task := range allTasks {
		m.taskByID[task.ID] = task
		m.titleByID[task.ID] = task.Title
	}
	m.pruneMarkedSelection(visibleTasks)

	m.days = buildCalendarDays(applyEffectiveDue(visibleTasks, m.effectiveDue), m.startDate, m.daysCount)
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

func (m *calendarTUIModel) pruneMarkedSelection(visibleTasks []shelf.Task) {
	if len(m.markedTaskIDs) == 0 {
		return
	}
	visible := make(map[string]struct{}, len(visibleTasks))
	for _, task := range visibleTasks {
		visible[task.ID] = struct{}{}
	}
	for taskID := range m.markedTaskIDs {
		if _, ok := visible[taskID]; !ok {
			delete(m.markedTaskIDs, taskID)
		}
	}
}

func (m *calendarTUIModel) rebuildModeState() {
	if m.mode == calendarModeTree {
		m.rebuildSections()
		m.rebuildTreeRows()
		return
	}
	if m.mode == calendarModeBoard {
		m.rebuildSections()
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
	m.rangeMarkMode = false
	m.rangeAnchorID = ""
	m.rangeBaseIDs = map[string]struct{}{}
	m.moveMode = false
	m.moveSourceIDs = nil
	m.rebuildModeState()
	m.message = fmt.Sprintf("mode: %s", mode)
}

func (m calendarTUIModel) markedCount() int {
	return len(m.markedTaskIDs)
}

func (m calendarTUIModel) isMarkedTask(taskID string) bool {
	_, ok := m.markedTaskIDs[taskID]
	return ok
}

func (m *calendarTUIModel) toggleMarkedSelection() {
	task, ok := m.selectedTask()
	if !ok {
		m.message = "選択中の task がありません"
		return
	}
	if m.markedTaskIDs == nil {
		m.markedTaskIDs = map[string]struct{}{}
	}
	if _, exists := m.markedTaskIDs[task.ID]; exists {
		delete(m.markedTaskIDs, task.ID)
		m.message = fmt.Sprintf("unmarked: %s", task.Title)
		return
	}
	m.markedTaskIDs[task.ID] = struct{}{}
	m.message = fmt.Sprintf("marked %d task(s)", len(m.markedTaskIDs))
}

func (m *calendarTUIModel) unmarkCurrentSelection() {
	task, ok := m.selectedTask()
	if !ok {
		m.message = "選択中の task がありません"
		return
	}
	if _, exists := m.markedTaskIDs[task.ID]; !exists {
		m.message = "current task is not marked"
		return
	}
	delete(m.markedTaskIDs, task.ID)
	delete(m.rangeBaseIDs, task.ID)
	if m.rangeAnchorID == task.ID {
		m.rangeMarkMode = false
		m.rangeAnchorID = ""
	}
	m.message = fmt.Sprintf("unmarked: %s", task.Title)
}

func (m *calendarTUIModel) toggleRangeSelectionMode() {
	task, ok := m.selectedTask()
	if !ok {
		m.message = "選択中の task がありません"
		return
	}
	if m.rangeMarkMode {
		m.rangeMarkMode = false
		m.rangeAnchorID = ""
		m.rangeBaseIDs = map[string]struct{}{}
		m.message = fmt.Sprintf("range select fixed (%d task)", m.markedCount())
		return
	}
	m.rangeMarkMode = true
	m.rangeAnchorID = task.ID
	m.rangeBaseIDs = cloneIDSet(m.markedTaskIDs)
	if m.rangeBaseIDs == nil {
		m.rangeBaseIDs = map[string]struct{}{}
	}
	m.rangeBaseIDs[task.ID] = struct{}{}
	if m.markedTaskIDs == nil {
		m.markedTaskIDs = map[string]struct{}{}
	}
	m.markedTaskIDs[task.ID] = struct{}{}
	m.message = "range select started"
}

func (m calendarTUIModel) activeTaskIDs() []string {
	if len(m.markedTaskIDs) > 0 {
		return m.orderedMarkedTaskIDs()
	}
	task, ok := m.selectedTask()
	if !ok {
		return nil
	}
	return []string{task.ID}
}

func (m calendarTUIModel) copySubtreeRootIDs() []string {
	taskIDs := m.activeTaskIDs()
	if len(taskIDs) <= 1 {
		return taskIDs
	}
	roots := make([]string, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		task, ok := m.taskByID[taskID]
		if !ok {
			continue
		}
		include := true
		parentID := task.Parent
		for parentID != "" {
			if slices.Contains(taskIDs, parentID) {
				include = false
				break
			}
			parent, ok := m.taskByID[parentID]
			if !ok {
				break
			}
			parentID = parent.Parent
		}
		if include {
			roots = append(roots, taskID)
		}
	}
	return roots
}

func (m calendarTUIModel) orderedMarkedTaskIDs() []string {
	ordered := make([]string, 0, len(m.markedTaskIDs))
	appendIfMarked := func(taskID string) {
		if _, ok := m.markedTaskIDs[taskID]; ok {
			ordered = append(ordered, taskID)
		}
	}
	switch m.mode {
	case calendarModeTree:
		for _, row := range m.treeRows {
			appendIfMarked(row.Task.ID)
		}
	case calendarModeBoard:
		for _, column := range m.boardColumns {
			for _, task := range column.Tasks {
				appendIfMarked(task.ID)
			}
		}
	default:
		for taskID := range m.markedTaskIDs {
			ordered = append(ordered, taskID)
		}
		sort.Strings(ordered)
	}
	return ordered
}

func (m *calendarTUIModel) clearMarkedSelection() {
	m.markedTaskIDs = map[string]struct{}{}
	m.rangeMarkMode = false
	m.rangeAnchorID = ""
	m.rangeBaseIDs = map[string]struct{}{}
}

func (m *calendarTUIModel) syncRangeSelection() {
	if !m.rangeMarkMode || strings.TrimSpace(m.rangeAnchorID) == "" {
		return
	}
	order := m.selectionOrder()
	if len(order) == 0 {
		return
	}
	currentID := m.currentSelectableTaskID()
	if strings.TrimSpace(currentID) == "" {
		return
	}
	start := -1
	end := -1
	for i, taskID := range order {
		if taskID == m.rangeAnchorID {
			start = i
		}
		if taskID == currentID {
			end = i
		}
	}
	if start == -1 || end == -1 {
		return
	}
	if start > end {
		start, end = end, start
	}
	m.markedTaskIDs = cloneIDSet(m.rangeBaseIDs)
	if m.markedTaskIDs == nil {
		m.markedTaskIDs = map[string]struct{}{}
	}
	for _, taskID := range order[start : end+1] {
		m.markedTaskIDs[taskID] = struct{}{}
	}
}

func cloneIDSet(src map[string]struct{}) map[string]struct{} {
	if len(src) == 0 {
		return map[string]struct{}{}
	}
	dst := make(map[string]struct{}, len(src))
	for id := range src {
		dst[id] = struct{}{}
	}
	return dst
}

func (m calendarTUIModel) selectionOrder() []string {
	switch m.mode {
	case calendarModeTree:
		ids := make([]string, 0, len(m.treeRows))
		for _, row := range m.treeRows {
			ids = append(ids, row.Task.ID)
		}
		return ids
	case calendarModeBoard:
		ids := make([]string, 0)
		for _, column := range m.boardColumns {
			for _, task := range column.Tasks {
				ids = append(ids, task.ID)
			}
		}
		return ids
	default:
		return nil
	}
}

func (m calendarTUIModel) currentSelectableTaskID() string {
	task, ok := m.selectedTask()
	if !ok {
		return ""
	}
	return task.ID
}

func (m *calendarTUIModel) cycleMode(delta int) {
	order := []calendarMode{
		calendarModeCalendar,
		calendarModeTree,
		calendarModeBoard,
		calendarModeReview,
		calendarModeNow,
	}
	index := 0
	for i, mode := range order {
		if mode == m.mode {
			index = i
			break
		}
	}
	index += delta
	if index < 0 {
		index = len(order) - 1
	}
	if index >= len(order) {
		index = 0
	}
	m.switchMode(order[index])
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

func buildEffectiveDueMap(tasks []shelf.Task) map[string]string {
	byID := make(map[string]shelf.Task, len(tasks))
	for _, task := range tasks {
		byID[task.ID] = task
	}
	effective := make(map[string]string, len(tasks))
	var resolve func(task shelf.Task, visited map[string]struct{}) string
	resolve = func(task shelf.Task, visited map[string]struct{}) string {
		if due, ok := effective[task.ID]; ok {
			return due
		}
		if strings.TrimSpace(task.DueOn) != "" {
			effective[task.ID] = task.DueOn
			return task.DueOn
		}
		parentID := strings.TrimSpace(task.Parent)
		if parentID == "" {
			effective[task.ID] = ""
			return ""
		}
		if _, seen := visited[parentID]; seen {
			effective[task.ID] = ""
			return ""
		}
		parent, ok := byID[parentID]
		if !ok {
			effective[task.ID] = ""
			return ""
		}
		nextVisited := cloneIDSet(visited)
		nextVisited[parentID] = struct{}{}
		due := resolve(parent, nextVisited)
		effective[task.ID] = due
		return due
	}
	for _, task := range tasks {
		resolve(task, map[string]struct{}{task.ID: {}})
	}
	return effective
}

func applyEffectiveDue(tasks []shelf.Task, effectiveDue map[string]string) []shelf.Task {
	display := make([]shelf.Task, 0, len(tasks))
	for _, task := range tasks {
		clone := task
		if due, ok := effectiveDue[task.ID]; ok && strings.TrimSpace(due) != "" {
			clone.DueOn = due
		}
		display = append(display, clone)
	}
	return display
}

func (m *calendarTUIModel) rebuildSections() {
	prevSectionID := calendarSectionFocusedDay
	if section := m.currentSection(); section != nil {
		prevSectionID = section.ID
	}
	m.sections = buildCalendarSections(m.mode, m.focusedDay(), applyEffectiveDue(m.visibleTasks, m.effectiveDue), m.readiness, m.titleByID, m.blockingLink, m.sectionLimit)
	if len(m.sections) == 0 {
		m.sectionIndex = 0
		m.selectedTaskID = ""
		return
	}
	if m.selectedTaskID != "" {
		if secIdx, rowIdx, ok := findCalendarSectionTaskInSection(m.sections, prevSectionID, m.selectedTaskID); ok {
			m.sectionIndex = secIdx
			m.sectionRows[m.sections[secIdx].ID] = rowIdx
			return
		}
		findSectionTask := findCalendarSectionTask
		if m.mode == calendarModeReview || m.mode == calendarModeNow {
			findSectionTask = findPreferredCalendarSectionTask
		}
		if secIdx, rowIdx, ok := findSectionTask(m.sections, m.selectedTaskID); ok {
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
	m.treeRows = flattenCockpitTreeRows(nodes, "", true, m.showID, m.collapsedTree, m.effectiveDue)
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

func flattenCockpitTreeRows(nodes []shelf.TreeNode, prefix string, isRoot bool, showID bool, collapsed map[string]struct{}, effectiveDue map[string]string) []cockpitTreeRow {
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
		hasChildren := len(node.Children) > 0
		collapsedNode := false
		if hasChildren {
			_, collapsedNode = collapsed[node.Task.ID]
		}
		marker := "  "
		if hasChildren && collapsedNode {
			marker = "[+] "
		} else if hasChildren {
			marker = "[-] "
		}
		meta := fmt.Sprintf("%s/%s", node.Task.Kind, node.Task.Status)
		dueOn := strings.TrimSpace(node.Task.DueOn)
		dueInherited := false
		if effective := strings.TrimSpace(effectiveDue[node.Task.ID]); effective != "" {
			dueOn = effective
			dueInherited = strings.TrimSpace(node.Task.DueOn) == ""
		}
		label = fmt.Sprintf("%s%s%s%s", prefix, branch, marker, label)
		rows = append(rows, cockpitTreeRow{
			Task:         node.Task,
			Label:        label,
			Meta:         meta,
			DueOn:        dueOn,
			DueInherited: dueInherited,
			HasChildren:  hasChildren,
			Collapsed:    collapsedNode,
		})
		if !collapsedNode {
			rows = append(rows, flattenCockpitTreeRows(node.Children, nextPrefix, false, showID, collapsed, effectiveDue)...)
		}
	}
	return rows
}

func buildCalendarSections(mode calendarMode, focusedDay *calendarDay, tasks []shelf.Task, readiness map[string]shelf.TaskReadiness, titleByID map[string]string, blockingLinkType shelf.LinkType, sectionLimit int) []calendarSection {
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
			items = buildBlockedSectionItems(tasks, readiness, titleByID, blockingLinkType)
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
		return []calendarSection{{ID: calendarSectionFocusedDay, Title: "Selected Day"}, {ID: calendarSectionInbox, Title: "Inbox"}, {ID: calendarSectionOverdue, Title: "Overdue"}, {ID: calendarSectionToday, Title: "Today"}, {ID: calendarSectionBlocked, Title: "Blocked"}, {ID: calendarSectionReady, Title: "Ready"}}
	case calendarModeNow:
		return []calendarSection{{ID: calendarSectionFocusedDay, Title: "Selected Day"}, {ID: calendarSectionOverdue, Title: "Overdue"}, {ID: calendarSectionToday, Title: "Today"}}
	default:
		return []calendarSection{{ID: calendarSectionFocusedDay, Title: "Selected Day"}, {ID: calendarSectionOverdue, Title: "Overdue"}, {ID: calendarSectionToday, Title: "Today"}, {ID: calendarSectionReady, Title: "Ready"}}
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

func buildBlockedSectionItems(tasks []shelf.Task, readiness map[string]shelf.TaskReadiness, titleByID map[string]string, blockingLinkType shelf.LinkType) []calendarSectionItem {
	filtered := make([]shelf.Task, 0)
	reasons := map[string]string{}
	for _, task := range tasks {
		info := readiness[task.ID]
		if task.Status == "blocked" || info.BlockedByDeps {
			filtered = append(filtered, task)
			reasons[task.ID] = strings.Join(reviewBlockedBy(task, info, titleByID, blockingLinkType), "; ")
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
		Task:        task,
		ParentTitle: parentTitle,
		Subtitle:    fmt.Sprintf("%s/%s  due=%s  parent=%s", task.Kind, task.Status, dueText, parentTitle),
		Reason:      reason,
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

func findCalendarSectionTaskInSection(sections []calendarSection, sectionID calendarSectionID, taskID string) (int, int, bool) {
	for secIdx, section := range sections {
		if section.ID != sectionID {
			continue
		}
		for rowIdx, item := range section.Items {
			if item.Task.ID == taskID {
				return secIdx, rowIdx, true
			}
		}
		return 0, 0, false
	}
	return 0, 0, false
}

func findPreferredCalendarSectionTask(sections []calendarSection, taskID string) (int, int, bool) {
	for secIdx, section := range sections {
		if section.ID == calendarSectionFocusedDay {
			continue
		}
		for rowIdx, item := range section.Items {
			if item.Task.ID == taskID {
				return secIdx, rowIdx, true
			}
		}
	}
	return findCalendarSectionTask(sections, taskID)
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
	m.syncSelectionFromMain(section.Items[row].Task.ID)
}

func (m *calendarTUIModel) jumpSectionRowStart() {
	if m.mode == calendarModeTree {
		if m.moveMode {
			m.treeRowIndex = -1
			m.syncRangeSelection()
			return
		}
		if len(m.treeRows) == 0 {
			return
		}
		m.treeRowIndex = 0
		m.syncSelectionFromMain(m.treeRows[0].Task.ID)
		m.syncRangeSelection()
		return
	}
	section := m.currentSection()
	if section == nil || len(section.Items) == 0 {
		return
	}
	m.sectionRows[section.ID] = 0
	m.syncSelectionFromMain(section.Items[0].Task.ID)
}

func (m *calendarTUIModel) jumpSectionRowEnd() {
	if m.mode == calendarModeTree {
		if len(m.treeRows) == 0 {
			return
		}
		m.treeRowIndex = len(m.treeRows) - 1
		m.syncSelectionFromMain(m.treeRows[m.treeRowIndex].Task.ID)
		m.syncRangeSelection()
		return
	}
	section := m.currentSection()
	if section == nil || len(section.Items) == 0 {
		return
	}
	row := len(section.Items) - 1
	m.sectionRows[section.ID] = row
	m.syncSelectionFromMain(section.Items[row].Task.ID)
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
	if m.moveMode {
		rowCount := len(m.treeRows) + 1
		m.treeRowIndex += delta
		if m.treeRowIndex < -1 {
			m.treeRowIndex = rowCount - 2
		}
		if m.treeRowIndex >= rowCount-1 {
			m.treeRowIndex = -1
		}
		if m.treeRowIndex >= 0 {
			m.syncSelectionFromMain(m.treeRows[m.treeRowIndex].Task.ID)
		}
		m.syncRangeSelection()
		return
	}
	m.treeRowIndex += delta
	if m.treeRowIndex < 0 {
		m.treeRowIndex = len(m.treeRows) - 1
	}
	if m.treeRowIndex >= len(m.treeRows) {
		m.treeRowIndex = 0
	}
	m.syncSelectionFromMain(m.treeRows[m.treeRowIndex].Task.ID)
	m.syncRangeSelection()
}

func (m *calendarTUIModel) collapseCurrentTreeNode() {
	if len(m.treeRows) == 0 || m.treeRowIndex < 0 || m.treeRowIndex >= len(m.treeRows) {
		return
	}
	row := m.treeRows[m.treeRowIndex]
	if row.HasChildren && !row.Collapsed {
		m.collapsedTree[row.Task.ID] = struct{}{}
		m.rebuildTreeRows()
		m.message = fmt.Sprintf("collapsed: %s", row.Task.Title)
		return
	}
	parentID := m.taskByID[row.Task.ID].Parent
	if strings.TrimSpace(parentID) == "" {
		return
	}
	m.selectTaskByID(parentID)
}

func (m *calendarTUIModel) expandCurrentTreeNode() {
	if len(m.treeRows) == 0 || m.treeRowIndex < 0 || m.treeRowIndex >= len(m.treeRows) {
		return
	}
	row := m.treeRows[m.treeRowIndex]
	if row.HasChildren && row.Collapsed {
		delete(m.collapsedTree, row.Task.ID)
		m.rebuildTreeRows()
		m.message = fmt.Sprintf("expanded: %s", row.Task.Title)
		return
	}
}

func (m *calendarTUIModel) moveFocusedDayTask(delta int) {
	section := m.focusedDaySection()
	if section == nil || len(section.Items) == 0 {
		m.message = "selected day に task がありません"
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

func (m *calendarTUIModel) moveSelectedDayTask(delta int) {
	section := m.focusedDaySection()
	if section == nil || len(section.Items) == 0 {
		m.message = "selected day に task がありません"
		return
	}
	row := m.sectionRows[section.ID] + delta
	if row < 0 {
		row = len(section.Items) - 1
	}
	if row >= len(section.Items) {
		row = 0
	}
	taskID := section.Items[row].Task.ID
	m.sectionRows[section.ID] = row
	if m.mode == calendarModeCalendar {
		m.selectedTaskID = taskID
		return
	}
	m.selectTaskByIDPreservingSection(taskID)
	m.sectionRows[section.ID] = row
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
	m.syncSelectionFromMain(column.Tasks[row].ID)
	m.syncRangeSelection()
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
	m.syncSelectionFromMain(column.Tasks[row].ID)
	m.syncRangeSelection()
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
	m.syncSelectionFromMain(column.Tasks[0].ID)
	m.syncRangeSelection()
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
	m.syncSelectionFromMain(column.Tasks[row].ID)
	m.syncRangeSelection()
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

func (m *calendarTUIModel) moveSidebarFocusByDays(delta int) {
	m.moveFocusByDays(delta)
	m.syncMainSelectionToFocusedDay()
}

func (m *calendarTUIModel) moveSidebarFocusByMonths(delta int) {
	m.moveFocusByMonths(delta)
	m.syncMainSelectionToFocusedDay()
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
	m.syncFocusedDayRow()
}

func (m *calendarTUIModel) syncMainSelectionToFocusedDay() {
	if m.mode == calendarModeCalendar {
		return
	}
	section := m.focusedDaySection()
	if section == nil || len(section.Items) == 0 {
		return
	}
	row := m.sectionRows[section.ID]
	if row < 0 {
		row = 0
	}
	if row >= len(section.Items) {
		row = len(section.Items) - 1
	}
	m.selectTaskByIDPreservingSection(section.Items[row].Task.ID)
}

func (m *calendarTUIModel) syncSelectionFromMain(taskID string) {
	m.selectedTaskID = taskID
	if m.mode != calendarModeCalendar {
		m.syncFocusedDateToTask(taskID)
	}
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

func (m calendarTUIModel) openEditorForSelectedTaskEdges() (tea.Model, tea.Cmd) {
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
	edgePath := filepath.Join(shelf.EdgesDir(m.rootDir), task.ID+".toml")
	if _, err := os.Stat(edgePath); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(edgePath, []byte(""), 0o644); err != nil {
			m.message = err.Error()
			return m, nil
		}
	}
	args = append(args, edgePath)

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

func (m *calendarTUIModel) beginAddMode(atRoot bool) {
	m.addMode = true
	m.addTitle = ""
	m.addTitleCursor = 0
	m.addKind = m.defaultKind
	m.addField = calendarAddFieldTitle
	m.addAtRoot = atRoot
	if atRoot {
		m.message = "root task title を入力"
		return
	}
	m.message = "子 task title を入力"
}

func (m *calendarTUIModel) createTaskFromAddMode(title string) error {
	input := shelf.AddTaskInput{
		Title:  title,
		Kind:   m.addKind,
		Status: m.defaultStatus,
	}
	switch m.mode {
	case calendarModeCalendar, calendarModeReview, calendarModeNow:
		day := m.focusedDay()
		if day == nil {
			return fmt.Errorf("選択中の日付がありません")
		}
		input.DueOn = day.Date
	case calendarModeBoard:
		if len(m.boardColumns) > 0 && m.boardColumnIdx >= 0 && m.boardColumnIdx < len(m.boardColumns) {
			input.Status = m.boardColumns[m.boardColumnIdx].Status
		}
	}
	if !m.addAtRoot {
		task, ok := m.selectedTask()
		if !ok {
			return fmt.Errorf("親 task が選択されていません")
		}
		input.Parent = task.ID
	}
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
	if err := m.reload(); err != nil {
		return err
	}
	if !calendarStatusIncluded(m.statuses, created.Status) {
		m.replaceTaskInVisibleState(created)
		m.insertTaskOnFocusedDay(created)
		m.rebuildModeState()
		m.selectTaskByID(created.ID)
		m.message = fmt.Sprintf("Created %s on %s (current filter excludes it; visible until reload)", created.Title, created.DueOn)
		if strings.TrimSpace(created.DueOn) == "" {
			m.message = fmt.Sprintf("Created %s (current filter excludes it; visible until reload)", created.Title)
		}
		return nil
	}
	m.selectTaskByID(created.ID)
	if strings.TrimSpace(created.DueOn) != "" {
		m.message = fmt.Sprintf("Created %s on %s", created.Title, created.DueOn)
		return nil
	}
	m.message = fmt.Sprintf("Created %s", created.Title)
	return nil
}

func (m *calendarTUIModel) createTaskOnFocusedDay(title string) error {
	m.addAtRoot = true
	return m.createTaskFromAddMode(title)
}

func (m *calendarTUIModel) applySelectedTaskKind(kind shelf.Kind) error {
	taskIDs := m.activeTaskIDs()
	if len(taskIDs) == 0 {
		return fmt.Errorf("選択中の task がありません")
	}
	updatedCount := 0
	if err := withWriteLock(m.rootDir, func() error {
		if err := prepareUndoSnapshot(m.rootDir, "calendar-kind"); err != nil {
			return err
		}
		for _, taskID := range taskIDs {
			if _, err := shelf.SetTask(m.rootDir, taskID, shelf.SetTaskInput{
				Kind: &kind,
			}); err != nil {
				return err
			}
			updatedCount++
		}
		return nil
	}); err != nil {
		return err
	}
	if err := m.reload(); err != nil {
		return err
	}
	m.clearMarkedSelection()
	m.selectTaskByID(taskIDs[0])
	if updatedCount == 1 {
		m.message = fmt.Sprintf("Updated kind to %s", kind)
		return nil
	}
	m.message = fmt.Sprintf("Updated kind to %s for %d tasks", kind, updatedCount)
	return nil
}

func (m *calendarTUIModel) applySelectedTaskTags(tags []string) error {
	task, ok := m.selectedTask()
	if !ok {
		return fmt.Errorf("選択中の task がありません")
	}
	normalized := shelf.NormalizeTags(tags)
	if err := withWriteLock(m.rootDir, func() error {
		if err := prepareUndoSnapshot(m.rootDir, "calendar-tags"); err != nil {
			return err
		}
		_, err := shelf.SetTask(m.rootDir, task.ID, shelf.SetTaskInput{
			Tags: &normalized,
		})
		return err
	}); err != nil {
		return err
	}
	if err := m.reload(); err != nil {
		return err
	}
	m.selectTaskByID(task.ID)
	if len(normalized) == 0 {
		m.message = "Cleared tags"
		return nil
	}
	m.message = fmt.Sprintf("Updated tags: %s", strings.Join(normalized, ", "))
	return nil
}

func (m calendarTUIModel) currentTagBulkDelta() shelf.SetTaskInput {
	addTags := make([]string, 0, len(m.tagBulkStates))
	removeTags := make([]string, 0, len(m.tagBulkStates))
	for _, tag := range m.tagChoices {
		switch m.tagBulkState(tag) {
		case calendarTagBulkStateAdd:
			addTags = append(addTags, tag)
		case calendarTagBulkStateRemove:
			removeTags = append(removeTags, tag)
		}
	}
	return shelf.SetTaskInput{
		AddTags:    addTags,
		RemoveTags: removeTags,
	}
}

func (m *calendarTUIModel) applySelectedTaskTagDelta(input shelf.SetTaskInput) error {
	addTags := shelf.NormalizeTags(input.AddTags)
	removeTags := shelf.NormalizeTags(input.RemoveTags)
	if len(addTags) == 0 && len(removeTags) == 0 {
		m.message = "No tag changes selected"
		return nil
	}
	taskIDs := m.activeTaskIDs()
	if len(taskIDs) == 0 {
		return fmt.Errorf("選択中の task がありません")
	}
	if err := withWriteLock(m.rootDir, func() error {
		if err := prepareUndoSnapshot(m.rootDir, "calendar-tags"); err != nil {
			return err
		}
		for _, taskID := range taskIDs {
			if _, err := shelf.SetTask(m.rootDir, taskID, shelf.SetTaskInput{
				AddTags:    addTags,
				RemoveTags: removeTags,
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if err := m.reload(); err != nil {
		return err
	}
	m.clearMarkedSelection()
	m.selectTaskByID(taskIDs[0])
	m.message = formatCalendarTagDeltaMessage(len(taskIDs), addTags, removeTags)
	return nil
}

func formatCalendarTagDeltaMessage(taskCount int, addTags []string, removeTags []string) string {
	parts := make([]string, 0, 2)
	if len(addTags) > 0 {
		parts = append(parts, "+"+strings.Join(addTags, ","))
	}
	if len(removeTags) > 0 {
		parts = append(parts, "-"+strings.Join(removeTags, ","))
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("Updated tags for %d tasks (%s)", taskCount, strings.Join(parts, "; "))
}

func (m calendarTUIModel) selectedTitleCopyText() (string, int, error) {
	taskIDs := m.activeTaskIDs()
	titles := make([]string, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		task, ok := m.taskByID[taskID]
		if !ok {
			continue
		}
		titles = append(titles, task.Title)
	}
	if len(titles) == 0 {
		return "", 0, fmt.Errorf("コピー対象の title がありません")
	}
	return strings.Join(titles, m.copySeparator), len(titles), nil
}

func (m calendarTUIModel) selectedPathCopyText() (string, int, error) {
	taskIDs := m.activeTaskIDs()
	paths := make([]string, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		if _, ok := m.taskByID[taskID]; !ok {
			continue
		}
		paths = append(paths, filepath.Join(shelf.TasksDir(m.rootDir), taskID+".md"))
	}
	if len(paths) == 0 {
		return "", 0, fmt.Errorf("コピー対象の path がありません")
	}
	return strings.Join(paths, m.copySeparator), len(paths), nil
}

func (m calendarTUIModel) selectedBodyCopyText() (string, int, error) {
	taskIDs := m.activeTaskIDs()
	bodies := make([]string, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		task, ok := m.taskByID[taskID]
		if !ok || strings.TrimSpace(task.Body) == "" {
			continue
		}
		bodies = append(bodies, task.Body)
	}
	if len(bodies) == 0 {
		return "", 0, fmt.Errorf("コピー対象の本文がありません")
	}
	return strings.Join(bodies, m.copySeparator), len(bodies), nil
}

func (m calendarTUIModel) selectedSubtreeCopyText() (string, int, error) {
	rootIDs := m.copySubtreeRootIDs()
	if len(rootIDs) == 0 {
		return "", 0, fmt.Errorf("選択中の task がありません")
	}
	lines := make([]string, 0, len(rootIDs))
	count := 0
	for i, rootID := range rootIDs {
		subtreeText, subtreeCount := m.renderTaskSubtreeText(rootID, shelf.CopySubtreeStyleIndented)
		if strings.TrimSpace(subtreeText) == "" {
			continue
		}
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, subtreeText)
		count += subtreeCount
	}
	if count == 0 {
		return "", 0, fmt.Errorf("コピー対象の subtree がありません")
	}
	return strings.Join(lines, "\n"), count, nil
}

func (m *calendarTUIModel) copyClipboardPayload(text string, count int, singular string, plural string) error {
	if err := copyTextToClipboard(text); err != nil {
		return err
	}
	if count == 1 {
		m.message = "Copied " + singular
		return nil
	}
	m.message = fmt.Sprintf("Copied %d %s", count, plural)
	return nil
}

func (m *calendarTUIModel) copySelectedTaskTitles() error {
	text, count, err := m.selectedTitleCopyText()
	if err != nil {
		return err
	}
	return m.copyClipboardPayload(text, count, "title", "titles")
}

func (m *calendarTUIModel) copySelectedTaskPaths() error {
	text, count, err := m.selectedPathCopyText()
	if err != nil {
		return err
	}
	return m.copyClipboardPayload(text, count, "path", "paths")
}

func (m *calendarTUIModel) copySelectedTaskBodies() error {
	text, count, err := m.selectedBodyCopyText()
	if err != nil {
		return err
	}
	return m.copyClipboardPayload(text, count, "body", "bodies")
}

func (m *calendarTUIModel) copySelectedTaskSubtrees() error {
	text, count, err := m.selectedSubtreeCopyText()
	if err != nil {
		return err
	}
	return m.copyClipboardPayload(text, count, "subtree", "subtrees")
}

func (m *calendarTUIModel) beginLinkMode(action calendarLinkAction) {
	if _, ok := m.selectedTask(); !ok {
		m.message = "選択中の task がありません"
		return
	}
	m.linkMode = true
	m.linkAction = action
	m.linkIndex = 0
	m.linkQuery = ""
	m.linkQueryCursor = 0
	m.linkQueryMode = false
	m.linkTypeIndex = 0
	m.linkCollapsedTree = map[string]struct{}{}
	if action == calendarLinkActionAdd {
		m.message = "link 先を選択"
		return
	}
	if len(m.currentLinkCandidates()) == 0 {
		m.linkMode = false
		m.message = "削除できる outbound link がありません"
		return
	}
	m.message = "削除する outbound link を選択"
}

func (m calendarTUIModel) updateLinkMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.linkQueryMode && msg.Type == tea.KeyRunes && string(msg.Runes) == "/" {
		m.linkQueryMode = true
		m.linkQueryCursor = len([]rune(m.linkQuery))
		m.message = "query を入力"
		return m, nil
	}
	if m.linkQueryMode {
		switch msg.String() {
		case "esc", "ctrl+[":
			m.linkQueryMode = false
			m.message = "query 入力を終了"
			return m, nil
		case "enter":
			m.linkQueryMode = false
			m.message = "query を適用"
			return m, nil
		case "backspace":
			m.linkQuery, m.linkQueryCursor = interactive.DeleteRuneBeforeCursor(m.linkQuery, m.linkQueryCursor)
			m.linkIndex = 0
			return m, nil
		case "left":
			m.linkQueryCursor = interactive.MoveTextCursorLeft(m.linkQuery, m.linkQueryCursor)
			return m, nil
		case "right":
			m.linkQueryCursor = interactive.MoveTextCursorRight(m.linkQuery, m.linkQueryCursor)
			return m, nil
		default:
			if msg.Type == tea.KeyRunes {
				for _, r := range msg.Runes {
					m.linkQuery, m.linkQueryCursor = interactive.InsertRuneAtCursor(m.linkQuery, m.linkQueryCursor, r)
				}
				m.linkIndex = 0
				return m, nil
			}
		}
		return m, nil
	}
	switch msg.String() {
	case "esc", "q", "ctrl+[":
		m.linkMode = false
		m.linkQuery = ""
		m.linkQueryCursor = 0
		m.linkQueryMode = false
		m.message = "link 操作をキャンセルしました"
		return m, nil
	case "up", "k":
		if m.linkIndex > 0 {
			m.linkIndex--
		}
		return m, nil
	case "down", "j":
		if m.linkIndex < len(m.currentLinkCandidates())-1 {
			m.linkIndex++
		}
		return m, nil
	case "tab":
		if m.linkAction == calendarLinkActionAdd && len(m.linkTypes()) > 0 {
			m.linkTypeIndex = (m.linkTypeIndex + 1) % len(m.linkTypes())
		}
		return m, nil
	case "shift+tab":
		if m.linkAction == calendarLinkActionAdd && len(m.linkTypes()) > 0 {
			m.linkTypeIndex--
			if m.linkTypeIndex < 0 {
				m.linkTypeIndex = len(m.linkTypes()) - 1
			}
		}
		return m, nil
	case "left", "h":
		m.collapseCurrentLinkCandidate()
		return m, nil
	case "right", "l":
		m.expandCurrentLinkCandidate()
		return m, nil
	case "enter":
		candidates := m.currentLinkCandidates()
		if len(candidates) == 0 {
			m.linkMode = false
			m.message = "候補がありません"
			return m, nil
		}
		if m.linkIndex >= len(candidates) {
			m.linkIndex = len(candidates) - 1
		}
		candidate := candidates[m.linkIndex]
		if err := m.applyLinkCandidate(candidate); err != nil {
			m.linkMode = false
			m.message = err.Error()
			return m, nil
		}
		m.linkMode = false
		m.linkQuery = ""
		m.linkQueryCursor = 0
		m.linkQueryMode = false
		return m, nil
	}
	return m, nil
}

func (m *calendarTUIModel) collapseCurrentLinkCandidate() {
	candidates := m.currentLinkCandidates()
	if len(candidates) == 0 || m.linkIndex < 0 || m.linkIndex >= len(candidates) {
		return
	}
	candidate := candidates[m.linkIndex]
	if candidate.HasChildren && !candidate.Collapsed {
		m.linkCollapsedTree[candidate.TaskID] = struct{}{}
		m.selectLinkCandidate(candidate.TaskID)
		return
	}
	if strings.TrimSpace(candidate.Parent) != "" {
		m.selectLinkCandidate(candidate.Parent)
	}
}

func (m *calendarTUIModel) expandCurrentLinkCandidate() {
	candidates := m.currentLinkCandidates()
	if len(candidates) == 0 || m.linkIndex < 0 || m.linkIndex >= len(candidates) {
		return
	}
	candidate := candidates[m.linkIndex]
	if candidate.HasChildren && candidate.Collapsed {
		delete(m.linkCollapsedTree, candidate.TaskID)
		m.selectLinkCandidate(candidate.TaskID)
	}
}

func (m *calendarTUIModel) selectLinkCandidate(taskID string) {
	candidates := m.currentLinkCandidates()
	for i, candidate := range candidates {
		if candidate.TaskID == taskID {
			m.linkIndex = i
			return
		}
	}
	if len(candidates) == 0 {
		m.linkIndex = 0
		return
	}
	if m.linkIndex >= len(candidates) {
		m.linkIndex = len(candidates) - 1
	}
}

func (m calendarTUIModel) linkTypes() []shelf.LinkType {
	return append([]shelf.LinkType{}, m.linkChoices...)
}

func (m calendarTUIModel) linkTypeLabel() string {
	if m.linkAction != calendarLinkActionAdd {
		return ""
	}
	types := m.linkTypes()
	if len(types) == 0 {
		return ""
	}
	if m.linkTypeIndex < 0 || m.linkTypeIndex >= len(types) {
		return string(types[0])
	}
	return string(types[m.linkTypeIndex])
}

func (m calendarTUIModel) currentLinkCandidates() []calendarLinkCandidate {
	task, ok := m.selectedTask()
	if !ok {
		return nil
	}
	query := strings.ToLower(strings.TrimSpace(m.linkQuery))
	ordered := m.linkCandidatesInTreeOrder(task.ID)
	if m.linkAction == calendarLinkActionRemove {
		edgeStore := shelf.NewEdgeStore(m.rootDir)
		outbound, err := edgeStore.ListOutbound(task.ID)
		if err != nil {
			return nil
		}
		typesByTaskID := make(map[string]shelf.LinkType, len(outbound))
		for _, edge := range outbound {
			typesByTaskID[edge.To] = edge.Type
		}
		filtered := make([]calendarLinkCandidate, 0, len(outbound))
		for _, candidate := range ordered {
			linkType, exists := typesByTaskID[candidate.TaskID]
			if !exists {
				continue
			}
			candidate.Type = linkType
			if query != "" && !strings.Contains(strings.ToLower(candidate.searchText()), query) {
				continue
			}
			filtered = append(filtered, candidate)
		}
		return filtered
	}

	filtered := make([]calendarLinkCandidate, 0, len(ordered))
	for _, candidate := range ordered {
		if query != "" && !strings.Contains(strings.ToLower(candidate.searchText()), query) {
			continue
		}
		filtered = append(filtered, candidate)
	}
	return filtered
}

func (c calendarLinkCandidate) searchText() string {
	label := strings.TrimSpace(c.Label)
	if label == "" {
		label = c.Title
	}
	if c.Type != "" {
		return fmt.Sprintf("%s %s %s %s %s", c.TaskID, c.Title, c.Path, label, c.Type)
	}
	return fmt.Sprintf("%s %s %s %s", c.TaskID, c.Title, c.Path, label)
}

func (m calendarTUIModel) linkCandidatesInTreeOrder(excludeTaskID string) []calendarLinkCandidate {
	byParent := make(map[string][]shelf.Task)
	for _, task := range m.allTasks {
		if task.ID == excludeTaskID {
			continue
		}
		byParent[task.Parent] = append(byParent[task.Parent], task)
	}
	for parent := range byParent {
		sort.Slice(byParent[parent], func(i, j int) bool {
			return byParent[parent][i].ID < byParent[parent][j].ID
		})
	}
	return flattenLinkCandidates(byParent[""], byParent, "", true, m.showID, m.taskByID, m.linkCollapsedTree, "")
}

func flattenLinkCandidates(tasks []shelf.Task, byParent map[string][]shelf.Task, prefix string, isRoot bool, showID bool, byID map[string]shelf.Task, collapsed map[string]struct{}, parentID string) []calendarLinkCandidate {
	candidates := make([]calendarLinkCandidate, 0)
	for i, task := range tasks {
		isLast := i == len(tasks)-1
		branch := "├─ "
		nextPrefix := prefix + "│  "
		if isLast {
			branch = "└─ "
			nextPrefix = prefix + "   "
		}
		if isRoot {
			branch = ""
		}
		label := task.Title
		if showID {
			label = fmt.Sprintf("[%s] %s", shelf.ShortID(task.ID), label)
		}
		hasChildren := len(byParent[task.ID]) > 0
		collapsedNode := false
		if hasChildren {
			_, collapsedNode = collapsed[task.ID]
		}
		marker := ""
		if hasChildren {
			if collapsedNode {
				marker = "[+] "
			} else {
				marker = "[-] "
			}
		}
		label = prefix + branch + marker + label
		candidates = append(candidates, calendarLinkCandidate{
			TaskID:      task.ID,
			Title:       task.Title,
			Label:       label,
			Path:        buildTaskPath(task, byID),
			Parent:      parentID,
			HasChildren: hasChildren,
			Collapsed:   collapsedNode,
		})
		if !collapsedNode {
			candidates = append(candidates, flattenLinkCandidates(byParent[task.ID], byParent, nextPrefix, false, showID, byID, collapsed, task.ID)...)
		}
	}
	return candidates
}

func (m *calendarTUIModel) applyLinkCandidate(candidate calendarLinkCandidate) error {
	task, ok := m.selectedTask()
	if !ok {
		return fmt.Errorf("選択中の task がありません")
	}
	if m.linkAction == calendarLinkActionRemove {
		if err := withWriteLock(m.rootDir, func() error {
			if err := prepareUndoSnapshot(m.rootDir, "calendar-unlink"); err != nil {
				return err
			}
			removed, err := shelf.UnlinkTasks(m.rootDir, task.ID, candidate.TaskID, candidate.Type)
			if err != nil {
				return err
			}
			if !removed {
				return fmt.Errorf("link not found: %s --%s--> %s", task.ID, candidate.Type, candidate.TaskID)
			}
			return nil
		}); err != nil {
			return err
		}
		if err := m.reload(); err != nil {
			return err
		}
		m.selectTaskByID(task.ID)
		m.message = fmt.Sprintf("Removed %s --%s--> %s", task.Title, candidate.Type, candidate.Title)
		return nil
	}

	linkTypes := m.linkTypes()
	if m.linkTypeIndex < 0 || m.linkTypeIndex >= len(linkTypes) {
		m.linkTypeIndex = 0
	}
	linkType := linkTypes[m.linkTypeIndex]
	if err := withWriteLock(m.rootDir, func() error {
		if err := prepareUndoSnapshot(m.rootDir, "calendar-link"); err != nil {
			return err
		}
		return shelf.LinkTasks(m.rootDir, task.ID, candidate.TaskID, linkType)
	}); err != nil {
		return err
	}
	if err := m.reload(); err != nil {
		return err
	}
	m.selectTaskByID(task.ID)
	m.message = fmt.Sprintf("Linked %s --%s--> %s", task.Title, linkType, candidate.Title)
	return nil
}

func (m calendarTUIModel) beginMoveSelection() (tea.Model, tea.Cmd) {
	if m.mode != calendarModeTree {
		task, ok := m.selectedTask()
		if !ok {
			m.message = "move 対象の task がありません"
			return m, nil
		}
		m.switchMode(calendarModeTree)
		m.selectTaskByID(task.ID)
	}
	if m.rangeMarkMode {
		m.rangeMarkMode = false
		m.rangeAnchorID = ""
		m.rangeBaseIDs = map[string]struct{}{}
	}
	taskIDs := m.activeTaskIDs()
	if len(taskIDs) == 0 {
		m.message = "move 対象の task がありません"
		return m, nil
	}
	m.moveMode = true
	m.moveSourceIDs = append([]string{}, taskIDs...)
	m.message = fmt.Sprintf("move target を選択して Enter (%d task)", len(taskIDs))
	return m, nil
}

func (m calendarTUIModel) toggleArchivedState() (tea.Model, tea.Cmd) {
	taskIDs := m.activeTaskIDs()
	if len(taskIDs) == 0 {
		m.message = "no task selected"
		return m, nil
	}
	now := time.Now().Format(time.RFC3339)
	updatedTasks := make([]shelf.Task, 0, len(taskIDs))
	action := "archived"
	err := withWriteLock(m.rootDir, func() error {
		if err := prepareUndoSnapshot(m.rootDir, "calendar-archive"); err != nil {
			return err
		}
		for _, taskID := range taskIDs {
			task := m.taskByID[taskID]
			nextArchivedAt := now
			if strings.TrimSpace(task.ArchivedAt) != "" {
				nextArchivedAt = ""
				action = "unarchived"
			}
			updatedTask, err := shelf.SetTask(m.rootDir, taskID, shelf.SetTaskInput{ArchivedAt: &nextArchivedAt})
			if err != nil {
				return err
			}
			updatedTasks = append(updatedTasks, updatedTask)
		}
		return nil
	})
	if err != nil {
		m.message = err.Error()
		return m, nil
	}
	if err := m.reload(); err != nil {
		m.message = err.Error()
		return m, nil
	}
	m.clearMarkedSelection()
	if len(updatedTasks) == 1 {
		m.message = fmt.Sprintf("%s %s", action, updatedTasks[0].Title)
		return m, nil
	}
	m.message = fmt.Sprintf("%s %d task(s)", action, len(updatedTasks))
	return m, nil
}

func (m calendarTUIModel) applyMoveSelection() (tea.Model, tea.Cmd) {
	if m.mode != calendarModeTree {
		m.moveMode = false
		m.moveSourceIDs = nil
		m.message = "move は Tree mode で使ってください"
		return m, nil
	}
	sources := append([]string{}, m.moveSourceIDs...)
	if len(sources) == 0 {
		sources = m.activeTaskIDs()
	}
	if len(sources) == 0 {
		m.moveMode = false
		m.message = "move 対象の task がありません"
		return m, nil
	}
	targetParentID := ""
	targetLabel := "root"
	if m.treeRowIndex >= 0 {
		target, ok := m.selectedTask()
		if !ok {
			m.message = "move 先の task がありません"
			return m, nil
		}
		targetParentID = target.ID
		targetLabel = target.Title
	}
	for _, sourceID := range sources {
		if targetParentID != "" && sourceID == targetParentID {
			m.message = "自分自身の下には移動できません"
			return m, nil
		}
		if targetParentID != "" && isDescendantTask(m.taskByID, targetParentID, sourceID) {
			m.message = "子孫 task の下には移動できません"
			return m, nil
		}
	}

	updated := make([]shelf.Task, 0, len(sources))
	err := withWriteLock(m.rootDir, func() error {
		if err := prepareUndoSnapshot(m.rootDir, "tree-move"); err != nil {
			return err
		}
		for _, sourceID := range sources {
			task, err := shelf.SetTask(m.rootDir, sourceID, shelf.SetTaskInput{Parent: &targetParentID})
			if err != nil {
				return err
			}
			updated = append(updated, task)
		}
		return nil
	})
	if err != nil {
		m.message = err.Error()
		return m, nil
	}
	if err := m.reload(); err != nil {
		m.message = err.Error()
		return m, nil
	}
	m.moveMode = false
	m.moveSourceIDs = nil
	m.clearMarkedSelection()
	if targetParentID != "" {
		m.selectTaskByID(targetParentID)
	}
	m.message = fmt.Sprintf("moved %d task(s) under %s", len(updated), targetLabel)
	return m, nil
}

func isDescendantTask(tasks map[string]shelf.Task, candidateID string, ancestorID string) bool {
	current := strings.TrimSpace(candidateID)
	for current != "" {
		if current == ancestorID {
			return true
		}
		task, ok := tasks[current]
		if !ok {
			return false
		}
		current = strings.TrimSpace(task.Parent)
	}
	return false
}

func (m *calendarTUIModel) applySnoozeOption(option snoozePreset) error {
	taskIDs := m.activeTaskIDs()
	if len(taskIDs) == 0 {
		return fmt.Errorf("選択中の task がありません")
	}
	nextDueByTaskID := make(map[string]string, len(taskIDs))
	dueTargets := map[string]struct{}{}

	if err := withWriteLock(m.rootDir, func() error {
		if err := prepareUndoSnapshot(m.rootDir, "calendar-snooze"); err != nil {
			return err
		}
		for _, taskID := range taskIDs {
			task, ok := m.taskByID[taskID]
			if !ok {
				continue
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
			if _, err := shelf.SetTask(m.rootDir, task.ID, shelf.SetTaskInput{DueOn: &nextDue}); err != nil {
				return err
			}
			nextDueByTaskID[task.ID] = nextDue
			dueTargets[nextDue] = struct{}{}
		}
		return nil
	}); err != nil {
		return err
	}
	if err := m.reload(); err != nil {
		return err
	}
	m.clearMarkedSelection()
	m.selectTaskByID(taskIDs[0])
	if len(taskIDs) == 1 {
		task := m.taskByID[taskIDs[0]]
		m.message = fmt.Sprintf("Snoozed %s to %s", task.Title, nextDueByTaskID[taskIDs[0]])
		return nil
	}
	if len(dueTargets) == 1 {
		for nextDue := range dueTargets {
			m.message = fmt.Sprintf("Snoozed %d tasks to %s", len(nextDueByTaskID), nextDue)
			return nil
		}
	}
	m.message = fmt.Sprintf("Snoozed %d tasks using %s", len(nextDueByTaskID), option.Label)
	return nil
}

func (m calendarTUIModel) applyStatusChange(nextStatus shelf.Status) (tea.Model, tea.Cmd) {
	taskIDs := m.activeTaskIDs()
	if len(taskIDs) == 0 {
		m.message = "no task selected"
		return m, nil
	}
	titles := make([]string, 0, len(taskIDs))
	updatedTasks := make([]shelf.Task, 0, len(taskIDs))
	err := withWriteLock(m.rootDir, func() error {
		if err := prepareUndoSnapshot(m.rootDir, "calendar-status"); err != nil {
			return err
		}
		for _, taskID := range taskIDs {
			task := m.taskByID[taskID]
			updatedTask, err := shelf.SetTask(m.rootDir, taskID, shelf.SetTaskInput{Status: &nextStatus})
			if err != nil {
				return err
			}
			titles = append(titles, task.Title)
			updatedTasks = append(updatedTasks, updatedTask)
		}
		return nil
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
		m.clearMarkedSelection()
		m.selectTaskByID(updatedTasks[0].ID)
		m.message = formatBatchStatusMessage(titles, nextStatus, false)
		return m, nil
	}
	for _, updatedTask := range updatedTasks {
		m.replaceTaskInVisibleState(updatedTask)
	}
	m.rebuildModeState()
	m.clearMarkedSelection()
	m.selectTaskByID(updatedTasks[0].ID)
	m.message = formatBatchStatusMessage(titles, nextStatus, true)
	return m, nil
}

func formatBatchStatusMessage(titles []string, nextStatus shelf.Status, filteredOut bool) string {
	if len(titles) == 0 {
		return ""
	}
	if len(titles) == 1 {
		if filteredOut {
			return fmt.Sprintf("%s -> %s (current filter excludes it; visible until reload)", titles[0], nextStatus)
		}
		return fmt.Sprintf("%s -> %s", titles[0], nextStatus)
	}
	if filteredOut {
		return fmt.Sprintf("%d tasks -> %s (current filter excludes them; visible until reload)", len(titles), nextStatus)
	}
	return fmt.Sprintf("%d tasks -> %s", len(titles), nextStatus)
}

func calendarStatusIncluded(statuses []shelf.Status, target shelf.Status) bool {
	if len(statuses) == 0 {
		return true
	}
	for _, status := range statuses {
		if status == target {
			return true
		}
	}
	return false
}

func deriveActiveStatuses(choices []shelf.Status, filter shelf.TaskFilter) []shelf.Status {
	active := make([]shelf.Status, 0, len(choices))
	if len(filter.Statuses) > 0 {
		for _, status := range filter.Statuses {
			if slices.Contains(filter.NotStatuses, status) {
				continue
			}
			active = append(active, status)
		}
		return active
	}
	for _, status := range choices {
		if slices.Contains(filter.NotStatuses, status) {
			continue
		}
		active = append(active, status)
	}
	return active
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
	m.selectTaskByIDWithOptions(taskID, false)
}

func (m *calendarTUIModel) selectTaskByIDPreservingSection(taskID string) {
	m.selectTaskByIDWithOptions(taskID, true)
}

func (m *calendarTUIModel) selectTaskByIDWithOptions(taskID string, preserveCurrentSection bool) {
	m.selectedTaskID = taskID
	if m.mode == calendarModeTree {
		for i, row := range m.treeRows {
			if row.Task.ID == taskID {
				m.treeRowIndex = i
				m.syncFocusedDateToTask(taskID)
				m.syncRangeSelection()
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
					m.syncFocusedDateToTask(taskID)
					m.syncRangeSelection()
					return
				}
			}
		}
		m.clampBoardSelection()
		return
	}
	findSectionTask := findCalendarSectionTask
	if m.mode == calendarModeReview || m.mode == calendarModeNow {
		if preserveCurrentSection {
			if section := m.currentSection(); section != nil {
				if secIdx, rowIdx, ok := findCalendarSectionTaskInSection(m.sections, section.ID, taskID); ok {
					m.sectionIndex = secIdx
					m.sectionRows[m.sections[secIdx].ID] = rowIdx
					m.syncFocusedDateToTask(taskID)
					return
				}
			}
		}
		findSectionTask = findPreferredCalendarSectionTask
	}
	if secIdx, rowIdx, ok := findSectionTask(m.sections, taskID); ok {
		m.sectionIndex = secIdx
		m.sectionRows[m.sections[secIdx].ID] = rowIdx
		m.syncFocusedDateToTask(taskID)
		return
	}
	m.clampSectionSelection()
}

func (m *calendarTUIModel) syncFocusedDateToTask(taskID string) {
	if m.mode == calendarModeCalendar {
		return
	}
	due := strings.TrimSpace(m.effectiveDue[taskID])
	if due == "" {
		due = strings.TrimSpace(m.taskByID[taskID].DueOn)
	}
	if due == "" {
		return
	}
	target, err := time.ParseInLocation("2006-01-02", due, time.Now().Location())
	if err != nil {
		return
	}
	newStart, newIndex := planCalendarWindow(m.startDate, m.daysCount, target)
	m.startDate = newStart
	m.days = buildCalendarDays(applyEffectiveDue(m.visibleTasks, m.effectiveDue), m.startDate, m.daysCount)
	m.dayIndex = newIndex
	m.syncFocusedDayRow()
	m.rebuildSections()
}

func (m *calendarTUIModel) syncFocusedDayRow() {
	focused := m.focusedDay()
	if focused == nil {
		m.sectionRows[calendarSectionFocusedDay] = 0
		return
	}
	for i, task := range focused.Tasks {
		if task.ID == m.selectedTaskID {
			m.sectionRows[calendarSectionFocusedDay] = i
			return
		}
	}
	m.sectionRows[calendarSectionFocusedDay] = 0
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
		titleStyle.Render("Cockpit"),
		accentStyle.Render("[" + strings.ToUpper(string(m.mode)) + "]"),
		metaStyle.Render("Date " + focused.Format("Mon 2006-01-02")),
		metaStyle.Render(fmt.Sprintf("Range %s..%s", m.days[0].Date, m.days[len(m.days)-1].Date)),
		metaStyle.Render("Filter " + formatCalendarStatusFilter(m.statuses)),
		metaStyle.Render("?:help"),
	}
	if m.markedCount() > 0 {
		parts = append(parts, accentStyle.Render(fmt.Sprintf("Marked %d", m.markedCount())))
	}
	if m.moveMode {
		parts = append(parts, accentStyle.Render("Move Target"))
	}
	if labels := m.transientModeLabels(); len(labels) > 0 {
		parts = append(parts, renderTransientModeBadges(labels))
		parts = append(parts, metaStyle.Render("Ctrl+[: normal"))
	}
	return trimLine(strings.Join(parts, "  "), max(48, m.width-2))
}

func renderTransientModeBadges(labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	parts := []string{lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true).Render("Mode")}
	for _, label := range labels {
		parts = append(parts, transientModeBadgeStyle(label).Render(label))
	}
	return strings.Join(parts, " ")
}

func transientModeBadgeStyle(label string) lipgloss.Style {
	background := lipgloss.Color("238")
	foreground := lipgloss.Color("255")
	switch label {
	case "help":
		background = lipgloss.Color("61")
	case "range":
		background = lipgloss.Color("98")
	case "move":
		background = lipgloss.Color("166")
	case "add":
		background = lipgloss.Color("29")
	case "snooze":
		background = lipgloss.Color("94")
	case "filter":
		background = lipgloss.Color("62")
	}
	return lipgloss.NewStyle().
		Foreground(foreground).
		Background(background).
		Bold(true).
		Padding(0, 1)
}

func (m calendarTUIModel) transientModeLabels() []string {
	labels := make([]string, 0, 7)
	if m.showHelp {
		labels = append(labels, "help")
	}
	if m.rangeMarkMode {
		labels = append(labels, "range")
	}
	if m.moveMode {
		labels = append(labels, "move")
	}
	if m.addMode {
		labels = append(labels, "add")
	}
	if m.snoozeMode {
		labels = append(labels, "snooze")
	}
	if m.linkMode {
		labels = append(labels, "links")
	}
	if m.filterMode {
		labels = append(labels, "filter")
	}
	if m.kindMode {
		labels = append(labels, "kind")
	}
	if m.tagMode {
		labels = append(labels, "tags")
	}
	if m.textPromptMode {
		labels = append(labels, "input")
	}
	return labels
}

func calendarPanelStyle(totalWidth int, totalHeight int, borderColor lipgloss.Color) lipgloss.Style {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)
	innerWidth := max(1, totalWidth-style.GetHorizontalFrameSize())
	style = style.Width(innerWidth)
	if totalHeight > 0 {
		innerHeight := max(1, totalHeight-style.GetVerticalFrameSize())
		style = style.Height(innerHeight)
	}
	return style
}

func calendarPanelContentSize(totalWidth int, totalHeight int) (int, int) {
	frameStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)
	innerWidth := max(1, totalWidth-frameStyle.GetHorizontalFrameSize())
	innerHeight := 0
	if totalHeight > 0 {
		innerHeight = max(1, totalHeight-frameStyle.GetVerticalFrameSize())
	}
	return innerWidth, innerHeight
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

func splitSidebarHeights(totalHeight int) (int, int) {
	if totalHeight <= 0 {
		return 0, 0
	}
	usable := max(0, totalHeight-1)
	top := usable / 2
	bottom := usable - top
	return top, bottom
}

func splitSecondarySidebarHeights(totalHeight int) (int, int, int) {
	if totalHeight <= 0 {
		return 0, 0, 0
	}
	usable := max(0, totalHeight-2)
	calendar := usable * 40 / 98
	selectedDay := usable * 28 / 98
	inspector := usable - calendar - selectedDay
	return calendar, selectedDay, inspector
}

func calendarMainCellHeight(totalHeight int) int {
	if totalHeight <= 0 {
		return 4
	}
	usable := totalHeight - 10
	if usable <= 0 {
		return 4
	}
	return min(8, max(4, usable/6))
}

func renderCockpitHelpOverlay(mode calendarMode, width int, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	lines := []string{
		titleStyle.Render("Help"),
		mutedStyle.Render(fmt.Sprintf("mode=%s", mode)),
		mutedStyle.Render(calendarModeHelpSummary(mode)),
		"Tab: pane (non-calendar only)  C/T/B/R/N: mode  ?: close",
		"h/l: day move or tree collapse/expand  j/k: rows or weeks  n/p: Selected Day task switch",
		"PgUp/PgDn or Ctrl+U/D: scroll body  Home/End: top/bottom",
		"sidebar: Calendar / Selected Day / Inspector with two-way selection sync",
		"v: mark  u: clear marks  V: range mark  m: move in tree (root included)",
		"o/i/b/d/c: status  a: add child  A: add root  e: edit  y/Y/P/O: copy  M: advanced copy  K: kind  #: tags  f: filter  L/U: link/unlink  z: snooze  r: reload",
		"Enter: details  Ctrl+[: leave popup/input  q: close help or quit  Esc: close/cancel transient state",
	}
	return renderPopupBox(lines, width, height, lipgloss.Color("141"), -1)
}

func calendarModeHelpSummary(mode calendarMode) string {
	switch mode {
	case calendarModeCalendar:
		return "calendar: t today  h/l day  j/k week  [/] month  a child  A root"
	case calendarModeTree:
		return "tree: h/l collapse-expand  j/k rows  m move  v/V mark range"
	case calendarModeBoard:
		return "board: h/l columns  j/k rows  a child  A root  v/V mark range"
	case calendarModeReview:
		return "review: inbox / overdue / blocked / ready  j/k rows  h/l tabs"
	case calendarModeNow:
		return "now: today-focused worklist  j/k rows  n/p Selected Day sync"
	default:
		return "Cockpit: mode-specific task workspace"
	}
}

func renderCalendarLegend() string {
	item := func(color lipgloss.Color, label string) string {
		return lipgloss.NewStyle().Foreground(color).Render("■ " + label)
	}
	return strings.Join([]string{item(lipgloss.Color("203"), "blocked"), item(lipgloss.Color("220"), "in_progress"), item(lipgloss.Color("81"), "open"), item(lipgloss.Color("78"), "done"), item(lipgloss.Color("245"), "cancelled")}, "  ")
}

func renderCalendarModeTabs(mode calendarMode, width int) string {
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
	return trimLine(lipgloss.JoinHorizontal(lipgloss.Left, parts...), width)
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
	renderedWidth := 0
	start, end := 0, len(sections)
	if selected >= len(sections) {
		selected = len(sections) - 1
	}
	if selected < 0 {
		selected = 0
	}
	for {
		parts = parts[:0]
		renderedWidth = 0
		for i := start; i < end; i++ {
			label := fmt.Sprintf("%s %d", sections[i].Title, len(sections[i].Items))
			style := idleStyle
			if i == selected {
				style = activeStyle
			}
			tab := style.Render(label)
			nextWidth := renderedWidth + ansi.StringWidth(tab)
			if len(parts) > 0 {
				nextWidth++
			}
			if nextWidth > max(12, width) && len(parts) > 0 {
				break
			}
			if len(parts) > 0 {
				renderedWidth++
			}
			renderedWidth += ansi.StringWidth(tab)
			parts = append(parts, tab)
		}
		if renderedWidth <= max(12, width) || selected <= start || end-start <= 1 {
			break
		}
		start++
	}
	return trimLine(lipgloss.JoinHorizontal(lipgloss.Left, parts...), max(12, width))
}

func renderCalendarMainPane(m calendarTUIModel, month calendarMonthView, width int, height int, active bool) string {
	borderColor := lipgloss.Color("240")
	if active {
		borderColor = lipgloss.Color("45")
	}
	boxStyle := calendarPanelStyle(width, height, borderColor)
	contentWidth, contentHeight := calendarPanelContentSize(width, height)
	contentWidth = max(20, contentWidth)
	contentHeight = max(1, contentHeight)

	if m.mode == calendarModeTree {
		header := renderCockpitContextStrip(m, contentWidth)
		treeHeight := max(1, contentHeight-lipgloss.Height(header)-1)
		parts := []string{header, renderCockpitTreePane(m.treeRows, m.treeRowIndex, m.markedTaskIDs, m.moveMode, contentWidth, treeHeight)}
		return boxStyle.Render(strings.Join(parts, "\n\n"))
	}
	if m.mode == calendarModeBoard {
		header := renderCockpitContextStrip(m, contentWidth)
		boardHeight := max(1, contentHeight-lipgloss.Height(header)-1)
		parts := []string{header, renderCockpitBoardPane(m.boardColumns, m.boardColumnIdx, m.boardRowIndex, m.markedTaskIDs, m.showID, contentWidth, boardHeight)}
		return boxStyle.Render(strings.Join(parts, "\n\n"))
	}
	parts := make([]string, 0, 4)
	occupiedHeight := 0
	if m.mode != calendarModeCalendar {
		header := renderCockpitContextStrip(m, contentWidth)
		parts = append(parts, header)
		occupiedHeight += lipgloss.Height(header)
	}
	if m.mode != calendarModeCalendar && m.mode != calendarModeNow {
		tabs := renderCalendarSectionTabs(m.sections, m.sectionIndex, contentWidth)
		parts = append(parts, tabs)
		occupiedHeight += lipgloss.Height(tabs)
	}
	if m.mode == calendarModeCalendar {
		parts = append(parts, renderCalendarLegend())
		parts = append(parts, renderCalendarMonth(month, m.focusedDayLabel(), max(56, contentWidth), false, calendarMainCellHeight(height)))
		return boxStyle.Render(strings.Join(parts, "\n\n"))
	}
	listHeight := max(1, contentHeight-occupiedHeight-max(0, len(parts)-1))
	if m.mode == calendarModeNow {
		parts = append(parts, renderCalendarTriptychSections(m.sections, m.sectionIndex, m.sectionRows, m.showID, max(36, contentWidth), listHeight))
		return boxStyle.Render(strings.Join(parts, "\n\n"))
	}
	parts = append(parts, renderCalendarActiveSection(m.currentSection(), m.sectionRows, m.showID, contentWidth, listHeight))
	return boxStyle.Render(strings.Join(parts, "\n\n"))
}

func renderCockpitContextStrip(m calendarTUIModel, width int) string {
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	parts := []string{
		"Date " + m.focusedDayLabel(),
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

func renderCalendarSidebarPane(m calendarTUIModel, task shelf.Task, ok bool, width int, height int) string {
	topHeight, bottomHeight := splitSidebarHeights(height)
	focusedPane := renderSelectedDayPane(m, width, topHeight)
	inspector := renderCalendarInspectorPane(task, ok, m.showID, m.showTaskBody, m.taskByID, m.readiness, m.outboundCount, m.inboundCount, m.blockingLink, width, bottomHeight)
	return lipgloss.JoinVertical(lipgloss.Left, focusedPane, "", inspector)
}

func renderCalendarSecondarySidebarPane(m calendarTUIModel, task shelf.Task, ok bool, width int, height int, active bool) string {
	calendarHeight, selectedDayHeight, inspectorHeight := splitSecondarySidebarHeights(height)
	calendarPane := renderCalendarMiniSidebar(m, width, calendarHeight, active)
	selectedDayPane := renderSelectedDayPane(m, width, selectedDayHeight)
	inspector := renderCalendarInspectorPane(task, ok, m.showID, m.showTaskBody, m.taskByID, m.readiness, m.outboundCount, m.inboundCount, m.blockingLink, width, inspectorHeight)
	return lipgloss.JoinVertical(lipgloss.Left, calendarPane, "", selectedDayPane, "", inspector)
}

func renderCalendarMiniSidebar(m calendarTUIModel, width int, height int, active bool) string {
	focused, err := m.focusedDate()
	if err != nil {
		focused = time.Now().Local()
	}
	month := buildCalendarMonthView(m.days, focused)
	borderColor := lipgloss.Color("240")
	if active {
		borderColor = lipgloss.Color("45")
	}
	boxStyle := calendarPanelStyle(width, height, borderColor)
	contentWidth, contentHeight := calendarPanelContentSize(width, height)
	contentWidth = max(20, contentWidth)
	contentHeight = max(1, contentHeight)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	header := []string{
		titleStyle.Render("Calendar"),
		metaStyle.Render("selection synced"),
	}
	body := strings.Split(renderCalendarMonth(month, m.focusedDayLabel(), max(28, contentWidth), true, 2), "\n")
	return boxStyle.Render(renderFixedBlockWithHeader(header, body, contentWidth, contentHeight, -1))
}

func renderSelectedDayPane(m calendarTUIModel, width int, height int) string {
	boxStyle := calendarPanelStyle(width, height, lipgloss.Color("240"))
	contentWidth, contentHeight := calendarPanelContentSize(width, height)
	contentWidth = max(20, contentWidth)
	contentHeight = max(1, contentHeight)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	section := m.focusedDaySection()
	count := 0
	if section != nil {
		count = len(section.Items)
	}
	lines := []string{
		titleStyle.Render("Selected Day: " + m.focusedDayLabel()),
		metaStyle.Render("n/p: task switch"),
		metaStyle.Render(fmt.Sprintf("Tasks: %d", count)),
	}
	anchor := -1
	if section == nil || len(section.Items) == 0 {
		return boxStyle.Render(renderFixedBlockWithHeader(lines, []string{metaStyle.Render("(none)")}, contentWidth, contentHeight, anchor))
	}
	body := make([]string, 0, len(section.Items)*2)
	row := m.sectionRows[section.ID]
	if row < 0 {
		row = 0
	}
	if row >= len(section.Items) {
		row = len(section.Items) - 1
	}
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	for i, item := range section.Items {
		label := item.Task.Title
		if m.showID {
			label = fmt.Sprintf("[%s] %s", shelf.ShortID(item.Task.ID), label)
		}
		if i == row {
			anchor = len(body)
			body = append(body, selectedStyle.Render("> "+trimLine(label, max(18, contentWidth))))
		} else {
			body = append(body, "  "+trimLine(label, max(18, contentWidth)))
		}
		original := m.taskByID[item.Task.ID]
		inherited := strings.TrimSpace(original.DueOn) == "" && strings.TrimSpace(item.Task.DueOn) != ""
		body = append(body, "  "+trimLine(formatTaskMetaLine(item.Task.Kind, item.Task.Status, item.Task.DueOn, inherited), max(18, contentWidth-2)))
	}
	return boxStyle.Render(renderFixedBlockWithHeader(lines, body, contentWidth, contentHeight, anchor))
}

func renderCockpitTreePane(rows []cockpitTreeRow, selected int, marked map[string]struct{}, moveMode bool, width int, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	header := []string{titleStyle.Render("Tree")}
	anchor := -1
	if moveMode {
		header = append(header, mutedStyle.Render("move target を選んで Enter"))
		rootLabel := trimLine("(root)", max(20, width))
		if selected == -1 {
			anchor = 0
			return renderFixedBlockWithHeader(header, []string{selectedStyle.Render("> " + rootLabel)}, width, height, anchor)
		} else {
			header = append(header, "  "+rootLabel)
		}
	} else {
		header = append(header, mutedStyle.Render("h: collapse / parent  l: expand"))
	}
	lines := make([]string, 0, len(rows)*2)
	if len(rows) == 0 {
		lines = append(lines, mutedStyle.Render("(none)"))
		return renderFixedBlockWithHeader(header, lines, width, height, anchor)
	}
	for i, row := range rows {
		marker := " "
		if _, ok := marked[row.Task.ID]; ok {
			marker = "*"
		}
		label := trimLine(marker+" "+row.Label, max(20, width))
		meta := uiMuted(row.Meta)
		if strings.TrimSpace(row.DueOn) != "" {
			meta += "  due=" + uiDue(row.DueOn)
			if row.DueInherited {
				meta += uiMuted(" (inherited)")
			}
		}
		meta = trimLine(meta, max(18, width-2))
		if i == selected {
			anchor = len(lines)
			lines = append(lines, selectedStyle.Render("> "+label))
			lines = append(lines, "  "+meta)
		} else {
			lines = append(lines, "  "+label)
			lines = append(lines, "  "+meta)
		}
	}
	return renderFixedBlockWithHeader(header, lines, width, height, anchor)
}

func renderCockpitBoardPane(columns []boardColumn, selectedColumn int, rowIndex map[int]int, marked map[string]struct{}, showID bool, width int, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	if len(columns) == 0 {
		return renderFixedBlockWithHeader([]string{titleStyle.Render("Board")}, []string{mutedStyle.Render("(none)")}, width, height, -1)
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
			if _, ok := marked[task.ID]; ok {
				label = "* " + label
			}
			meta := trimLine(renderBoardTaskMeta(task), max(14, columnWidth-2))
			prefix := "  "
			if colIdx == selectedColumn && taskIdx == currentRow {
				prefix = "> "
				lines = append(lines, selectedStyle.Render(trimLine(prefix+label, max(14, columnWidth-2))))
				lines = append(lines, "  "+meta)
			} else {
				lines = append(lines, prefix+trimLine(label, max(14, columnWidth-2)))
				lines = append(lines, "  "+meta)
			}
		}
		style := lipgloss.NewStyle().Width(columnWidth)
		rendered = append(rendered, style.Render(strings.Join(lines, "\n")))
	}
	anchor := 0
	if selectedColumn >= 0 && selectedColumn < len(columns) {
		anchor = rowIndex[selectedColumn] * 2
	}
	return renderFixedBlockWithHeader([]string{titleStyle.Render("Board")}, strings.Split(joinFixedColumns(rendered, " │ "), "\n"), width, height, anchor)
}

func renderCalendarGridPane(month calendarMonthView, focusedDate string, width int, active bool) string {
	borderColor := lipgloss.Color("240")
	if active {
		borderColor = lipgloss.Color("45")
	}
	containerStyle := calendarPanelStyle(width, 0, borderColor)
	content := renderCalendarMonth(month, focusedDate, max(42, width-4), true, 2)
	return containerStyle.Render(content)
}

func renderCalendarActiveSection(section *calendarSection, sectionRows map[calendarSectionID]int, showID bool, width int, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)

	if section == nil {
		return renderFixedBlock([]string{mutedStyle.Render("No section")}, width, height, -1)
	}
	header := []string{titleStyle.Render(fmt.Sprintf("%s %d", section.Title, len(section.Items)))}
	anchor := -1
	if len(section.Items) == 0 {
		return renderFixedBlockWithHeader(header, []string{mutedStyle.Render("(none)")}, width, height, anchor)
	}
	lines := make([]string, 0, len(section.Items)*3)
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
			anchor = len(lines)
			rendered = selectedStyle.Render("> " + trimLine(label, max(18, width)))
		}
		lines = append(lines, rendered)
		lines = append(lines, "  "+renderCalendarItemMeta(item, max(18, width-2)))
		if i == row && strings.TrimSpace(item.Reason) != "" {
			lines = append(lines, mutedStyle.Render("  "+trimLine(item.Reason, max(18, width-2))))
		}
	}
	return renderFixedBlockWithHeader(header, lines, width, height, anchor)
}

func renderCalendarTriptychSections(sections []calendarSection, selected int, sectionRows map[calendarSectionID]int, showID bool, width int, height int) string {
	if len(sections) == 0 {
		return ""
	}
	gap := 3
	columnWidth := max(22, (width-gap*2)/3)
	rendered := make([]string, 0, 3)
	for i, section := range sections {
		rendered = append(rendered, lipgloss.NewStyle().Width(columnWidth).Render(
			renderCalendarSectionColumn(&section, i == selected, sectionRows, showID, columnWidth, height),
		))
	}
	return joinFixedColumns(rendered, " │ ")
}

func renderCalendarSectionColumn(section *calendarSection, active bool, sectionRows map[calendarSectionID]int, showID bool, width int, height int) string {
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
	headLines := []string{titleStyle.Render(header)}
	anchor := -1
	if len(section.Items) == 0 {
		return renderFixedBlockWithHeader(headLines, []string{mutedStyle.Render("(none)")}, width, height, anchor)
	}
	lines := make([]string, 0, len(section.Items)*3)
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
			anchor = len(lines)
			lines = append(lines, selectedStyle.Render("> "+trimLine(label, max(18, width))))
		} else {
			lines = append(lines, "  "+trimLine(label, max(18, width)))
		}
		lines = append(lines, "  "+renderCalendarItemMeta(item, max(18, width-2)))
		if i == row && active && strings.TrimSpace(item.Reason) != "" {
			lines = append(lines, mutedStyle.Render("  "+trimLine(item.Reason, max(18, width-2))))
		}
	}
	return renderFixedBlockWithHeader(headLines, lines, width, height, anchor)
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
	meta := fmt.Sprintf("%s/%s", uiKind(task.Kind), uiStatus(task.Status))
	if strings.TrimSpace(task.DueOn) != "" {
		meta += "  due=" + uiDue(task.DueOn)
	}
	return meta
}

func renderCalendarMonth(month calendarMonthView, focusedDate string, width int, compact bool, cellHeight int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	dayHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244"))
	cellWidth := max(8, (width-2)/7)
	if compact {
		cellWidth = min(12, cellWidth)
	}
	headers := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	headerCells := make([]string, 0, len(headers))
	for _, header := range headers {
		headerCells = append(headerCells, dayHeaderStyle.Width(cellWidth).Align(lipgloss.Center).Render(header))
	}
	title := month.Label
	if focusedDate != "" {
		title = fmt.Sprintf("%s - %s", month.Label, strings.ReplaceAll(focusedDate, "-", "/"))
	}
	rows := []string{titleStyle.Width(width).Render(title), lipgloss.JoinHorizontal(lipgloss.Top, headerCells...)}
	for _, week := range month.Weeks {
		cells := make([]string, 0, len(week))
		for _, cell := range week {
			cells = append(cells, renderCalendarCell(cell, focusedDate, cellWidth, compact, cellHeight))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	return strings.Join(rows, "\n")
}

func renderCalendarCell(cell calendarMonthCell, focusedDate string, cellWidth int, compact bool, cellHeight int) string {
	key := cell.Date.Format("2006-01-02")
	today := time.Now().Format("2006-01-02")
	contentWidth := max(6, cellWidth)
	height := 2
	if !compact {
		height = max(4, cellHeight)
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
	for i := 0; i < max(0, height-2); i++ {
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

	lines := []string{titleStyle.Render("Cockpit")}
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
			lines = append(lines, "      "+renderCalendarItemMeta(item, max(16, width-8)))
			if strings.TrimSpace(item.Reason) != "" {
				lines = append(lines, mutedStyle.Render("      "+trimLine(item.Reason, max(16, width-8))))
			}
		}
	}
	return boxStyle.Render(strings.Join(lines, "\n"))
}

func renderCalendarItemMeta(item calendarSectionItem, width int) string {
	parts := []string{formatTaskMetaLine(item.Task.Kind, item.Task.Status, item.Task.DueOn, false)}
	parts = append(parts, "parent="+uiMuted(item.ParentTitle))
	return trimLine(strings.Join(parts, "  "), width)
}

func renderCalendarInspectorPane(task shelf.Task, ok bool, showID bool, showTaskBody bool, taskByID map[string]shelf.Task, readiness map[string]shelf.TaskReadiness, outboundCount map[string]int, inboundCount map[string]int, blockingLinkType shelf.LinkType, width int, height int) string {
	boxStyle := calendarPanelStyle(width, height, lipgloss.Color("240"))
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
	dueLabel := uiDue(task.DueOn)
	if original, ok := taskByID[task.ID]; ok && strings.TrimSpace(original.DueOn) == "" && strings.TrimSpace(task.DueOn) != "" {
		dueLabel = dueLabel + " (inherited)"
	}
	lines = append(lines,
		titleStyle.Render(label),
		fmt.Sprintf("kind=%s  status=%s", uiKind(task.Kind), uiStatus(task.Status)),
		fmt.Sprintf("due=%s  repeat=%s", dueLabel, formatCalendarRepeat(task.RepeatEvery)),
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
				lines = append(lines, fmt.Sprintf("%s=%s", blockingLinkType, strings.Join(labels, ", ")))
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

func renderCalendarSnoozePicker(targetLabel string, selected int, width int, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	header := []string{
		titleStyle.Render("Snooze Presets"),
		helpStyle.Render("Target: " + targetLabel),
		helpStyle.Render("j/k: 移動  Enter: 決定  Esc/q: 戻る"),
	}
	lines := make([]string, 0, len(calendarSnoozeOptions()))
	anchor := -1
	for i, option := range calendarSnoozeOptions() {
		line := "  " + option.Label
		if i == selected {
			anchor = len(lines)
			line = selectedStyle.Render("> " + option.Label)
		}
		lines = append(lines, line)
	}
	return renderPopupBoxWithHeader(header, lines, width, height, lipgloss.Color("141"), anchor)
}

func renderCalendarLinkPicker(action calendarLinkAction, linkType string, query string, queryMode bool, selectedTask string, candidates []calendarLinkCandidate, selected int, width int, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	title := "Add Link"
	help := "/: query  j/k: move  h/l: collapse-expand  Left/Right: cursor  Tab/Shift+Tab: type  Enter: apply  Esc/q: close"
	if action == calendarLinkActionRemove {
		title = "Remove Link"
		help = "/: query  j/k: move  h/l: collapse-expand  Left/Right: cursor  Enter: remove  Esc/q: close"
	}
	header := []string{titleStyle.Render(title), helpStyle.Render(help), helpStyle.Render("Selected: " + selectedTask)}
	if action == calendarLinkActionAdd {
		header = append(header, fmt.Sprintf("type=%s  query=%s", linkType, displayPickerQuery(query)))
	} else {
		header = append(header, fmt.Sprintf("query=%s", displayPickerQuery(query)))
	}
	if queryMode {
		header = append(header, helpStyle.Render("query input mode"))
	}
	if len(candidates) == 0 {
		return renderPopupBoxWithHeader(header, []string{helpStyle.Render("(no candidates)")}, width, height, lipgloss.Color("212"), -1)
	}
	lines := make([]string, 0, len(candidates))
	anchor := -1
	for i, candidate := range candidates {
		label := candidate.Label
		if strings.TrimSpace(label) == "" {
			label = candidate.TaskID
		}
		if candidate.Type != "" {
			label = fmt.Sprintf("%s  (%s)", label, candidate.Type)
		}
		line := "  " + label
		if i == selected {
			anchor = len(lines)
			line = selectedStyle.Render("> " + label)
		}
		lines = append(lines, line)
	}
	return renderPopupBoxWithHeader(header, lines, width, height, lipgloss.Color("212"), anchor)
}

func pickerWindow(selected int, total int, visible int) (int, int) {
	if total <= visible {
		return 0, total
	}
	if selected < 0 {
		selected = 0
	}
	if selected >= total {
		selected = total - 1
	}
	start := selected - visible/2
	if start < 0 {
		start = 0
	}
	end := start + visible
	if end > total {
		end = total
		start = end - visible
	}
	return start, end
}

func displayPickerQuery(query string) string {
	if strings.TrimSpace(query) == "" {
		return "(none)"
	}
	return query
}

func renderCalendarFilterPicker(m calendarTUIModel, width int, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	sectionTitles := []string{"Include Status", "Exclude Status", "Include Kind", "Exclude Kind"}
	header := []string{
		titleStyle.Render("Filter"),
		mutedStyle.Render("h/l or Tab: section  j/k: move  Space: toggle  Enter: apply  Esc/q: cancel"),
	}
	lines := make([]string, 0, 32)
	anchor := -1
	for section := calendarFilterIncludeStatuses; section <= calendarFilterExcludeKinds; section++ {
		heading := sectionTitles[int(section)]
		if section == m.filterSection {
			heading = selectedStyle.Render("> " + heading)
		} else {
			heading = titleStyle.Render("  " + heading)
		}
		lines = append(lines, heading)
		options := m.filterOptionsForSection(section)
		for i, option := range options {
			marker := "[ ]"
			if m.filterSectionEnabled(section, option) {
				marker = "[x]"
			}
			line := fmt.Sprintf("  %s %s", marker, option)
			if section == m.filterSection && i == m.filterIndex {
				anchor = len(lines)
				line = selectedStyle.Render("> " + line)
			}
			lines = append(lines, line)
		}
	}
	return renderPopupBoxWithHeader(header, lines, width, height, lipgloss.Color("62"), anchor)
}

func (m calendarTUIModel) filterOptionsForSection(section calendarFilterSection) []string {
	switch section {
	case calendarFilterIncludeStatuses, calendarFilterExcludeStatuses:
		options := make([]string, 0, len(m.statusChoices))
		for _, status := range m.statusChoices {
			options = append(options, string(status))
		}
		return options
	default:
		options := make([]string, 0, len(m.kindChoices))
		for _, kind := range m.kindChoices {
			options = append(options, string(kind))
		}
		return options
	}
}

func (m calendarTUIModel) filterSectionEnabled(section calendarFilterSection, option string) bool {
	switch section {
	case calendarFilterIncludeStatuses:
		return slices.Contains(m.filter.Statuses, shelf.Status(option))
	case calendarFilterExcludeStatuses:
		return slices.Contains(m.filter.NotStatuses, shelf.Status(option))
	case calendarFilterIncludeKinds:
		return slices.Contains(m.filter.Kinds, shelf.Kind(option))
	default:
		return slices.Contains(m.filter.NotKinds, shelf.Kind(option))
	}
}

func renderCalendarKindPicker(selectedTask string, kinds []shelf.Kind, selected int, width int, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	header := []string{titleStyle.Render("Kind"), helpStyle.Render("Selected: " + selectedTask), helpStyle.Render("j/k: 移動  Enter: 決定  Esc/q: 戻る")}
	lines := make([]string, 0, len(kinds))
	anchor := -1
	for i, kind := range kinds {
		line := "  " + string(kind)
		if i == selected {
			anchor = len(lines)
			line = selectedStyle.Render("> " + string(kind))
		}
		lines = append(lines, line)
	}
	return renderPopupBoxWithHeader(header, lines, width, height, lipgloss.Color("81"), anchor)
}

func renderCalendarTagPicker(selectedTask string, tags []string, selectedTags []string, bulkMode bool, bulkStates map[string]calendarTagBulkState, selected int, inputMode bool, inputValue string, width int, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	targetLabel := "Selected: " + selectedTask
	helpLabel := "j/k: move  Space: toggle  Left/Right: cursor  Enter: Done/Add  Ctrl+S: save  Esc/q: close"
	if bulkMode {
		targetLabel = "Target: " + selectedTask
		helpLabel = "j/k: move  Space: cycle [+]/[-]/[ ]  Left/Right: cursor  Enter: Done/Add  Ctrl+S: save  Esc/q: close"
	}
	header := []string{
		titleStyle.Render("Tags"),
		helpStyle.Render(targetLabel),
		helpStyle.Render(helpLabel),
	}
	if inputMode {
		header = append(header, helpStyle.Render("Add new tag: "+inputValue))
	}
	options := []string{"Done", "+ Add new tag"}
	for _, tag := range tags {
		marker := "[ ]"
		if bulkMode {
			if bulkStates != nil {
				marker = markerForCalendarTagBulkState(bulkStates[tag])
			}
		} else if containsTag(selectedTags, tag) {
			marker = "[x]"
		}
		options = append(options, fmt.Sprintf("%s %s", marker, tag))
	}
	lines := make([]string, 0, len(options))
	anchor := -1
	for i, option := range options {
		line := "  " + option
		if i == selected {
			anchor = len(lines)
			line = selectedStyle.Render("> " + option)
		}
		lines = append(lines, line)
	}
	return renderPopupBoxWithHeader(header, lines, width, height, lipgloss.Color("220"), anchor)
}

func renderCalendarTextPrompt(title string, value string) string {
	boxStyle := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("141")).Padding(1, 2).Width(72)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	lines := []string{
		titleStyle.Render(title),
		helpStyle.Render("入力して Enter で確定  Left/Right で移動  Esc/q でキャンセル"),
		"Value: " + value,
	}
	return boxStyle.Render(strings.Join(lines, "\n"))
}

func renderCalendarAddComposer(date string, defaultKind shelf.Kind, defaultStatus shelf.Status, title string, field calendarAddField, targetLabel string, atRoot bool, width int, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	targetMode := "child"
	if atRoot {
		targetMode = "root"
	}
	titleLine := "Title: " + title
	kindLine := "Kind: " + string(defaultKind)
	if field == calendarAddFieldTitle {
		titleLine = activeStyle.Render("> " + titleLine)
	} else {
		titleLine = "  " + titleLine
	}
	if field == calendarAddFieldKind {
		kindLine = activeStyle.Render("> " + kindLine)
	} else {
		kindLine = "  " + kindLine
	}
	header := []string{
		titleStyle.Render("Add Task"),
		helpStyle.Render("Tab/Shift+Tab: switch  Left/Right: cursor  Enter: create  j/k: kind  Esc: cancel"),
		fmt.Sprintf("due=%s  status=%s  target=%s", date, defaultStatus, targetMode),
		fmt.Sprintf("parent=%s", trimLine(targetLabel, 42)),
	}
	lines := []string{titleLine, kindLine}
	anchor := 4
	if field == calendarAddFieldKind {
		anchor = 1
	} else {
		anchor = 0
	}
	return renderPopupBoxWithHeader(header, lines, width, height, lipgloss.Color("81"), anchor)
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
	if width <= 0 {
		return ""
	}
	value = trimLine(value, width)
	padding := width - ansi.StringWidth(value)
	if padding < 0 {
		padding = 0
	}
	return value + strings.Repeat(" ", padding)
}

func trimLine(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(value) <= width {
		return value
	}
	if width == 1 {
		return ansi.Truncate(value, width, "")
	}
	return ansi.Truncate(value, width, "…")
}

func renderFixedBlock(lines []string, width int, height int, anchor int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	if len(lines) == 0 {
		lines = []string{""}
	}
	start, end := pickerWindow(anchor, len(lines), height)
	if anchor < 0 || len(lines) <= height {
		start = 0
		end = min(len(lines), height)
	}
	rendered := make([]string, 0, height)
	for _, line := range lines[start:end] {
		rendered = append(rendered, padOrTrim(line, width))
	}
	for len(rendered) < height {
		rendered = append(rendered, strings.Repeat(" ", width))
	}
	return strings.Join(rendered, "\n")
}

func renderPopupBox(lines []string, width int, height int, borderColor lipgloss.Color, anchor int) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(borderColor).
		Padding(1, 2)
	innerWidth := max(1, width-boxStyle.GetHorizontalFrameSize())
	innerHeight := max(1, height-boxStyle.GetVerticalFrameSize())
	boxStyle = boxStyle.Width(innerWidth).Height(innerHeight)
	return boxStyle.Render(renderFixedBlock(lines, innerWidth, innerHeight, anchor))
}

func renderPopupBoxWithHeader(header []string, body []string, width int, height int, borderColor lipgloss.Color, anchor int) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(borderColor).
		Padding(1, 2)
	innerWidth := max(1, width-boxStyle.GetHorizontalFrameSize())
	innerHeight := max(1, height-boxStyle.GetVerticalFrameSize())
	boxStyle = boxStyle.Width(innerWidth).Height(innerHeight)
	return boxStyle.Render(renderFixedBlockWithHeader(header, body, innerWidth, innerHeight, anchor))
}

func renderFixedBlockWithHeader(header []string, body []string, width int, height int, anchor int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	rendered := make([]string, 0, height)
	for _, line := range header {
		if len(rendered) == height {
			break
		}
		rendered = append(rendered, padOrTrim(line, width))
	}
	bodyHeight := height - len(rendered)
	if bodyHeight > 0 {
		bodyBlock := renderFixedBlock(body, width, bodyHeight, anchor)
		rendered = append(rendered, strings.Split(bodyBlock, "\n")...)
	}
	for len(rendered) < height {
		rendered = append(rendered, strings.Repeat(" ", width))
	}
	return strings.Join(rendered[:height], "\n")
}

func formatTaskMetaLine(kind shelf.Kind, status shelf.Status, dueOn string, dueInherited bool) string {
	parts := []string{fmt.Sprintf("%s/%s", uiKind(kind), uiStatus(status))}
	if strings.TrimSpace(dueOn) != "" {
		due := "due=" + uiDue(dueOn)
		if dueInherited {
			due += uiMuted(" (inherited)")
		}
		parts = append(parts, due)
	}
	return strings.Join(parts, "  ")
}

func taskDueDetails(task shelf.Task, effectiveDue map[string]string, taskByID map[string]shelf.Task) (string, bool) {
	dueOn := strings.TrimSpace(task.DueOn)
	inherited := false
	if effective := strings.TrimSpace(effectiveDue[task.ID]); effective != "" {
		dueOn = effective
		inherited = strings.TrimSpace(taskByID[task.ID].DueOn) == ""
	}
	return dueOn, inherited
}
