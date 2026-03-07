package cli

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

type boardColumn struct {
	Status shelf.Status
	Tasks  []shelf.Task
}

type boardModel struct {
	rootDir     string
	showID      bool
	statuses    []shelf.Status
	columns     []boardColumn
	columnIndex int
	rowIndex    map[int]int
	width       int
	height      int
	message     string
}

func newBoardCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "board",
		Short: "Open a kanban-style status board",
		Example: "  shelf board\n" +
			"  shelf board --show-id",
		RunE: func(_ *cobra.Command, _ []string) error {
			if !interactive.IsTTY() {
				return errors.New("board はTTYが必要です")
			}
			model, err := newBoardModel(ctx.rootDir, ctx.showID)
			if err != nil {
				return err
			}
			program := tea.NewProgram(model, tea.WithAltScreen())
			_, err = program.Run()
			return err
		},
	}
	return cmd
}

func newBoardModel(rootDir string, showID bool) (boardModel, error) {
	cfg, err := shelf.LoadConfig(rootDir)
	if err != nil {
		return boardModel{}, err
	}
	model := boardModel{
		rootDir:     rootDir,
		showID:      showID,
		statuses:    append([]shelf.Status{}, cfg.Statuses...),
		columnIndex: 0,
		rowIndex:    map[int]int{},
	}
	if err := model.reload(); err != nil {
		return boardModel{}, err
	}
	return model, nil
}

func (m boardModel) Init() tea.Cmd {
	return nil
}

func (m boardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "left", "h":
			if m.columnIndex > 0 {
				m.columnIndex--
			}
			m.clampRow()
			return m, nil
		case "right", "l":
			if m.columnIndex < len(m.columns)-1 {
				m.columnIndex++
			}
			m.clampRow()
			return m, nil
		case "up", "k":
			if m.rowIndex[m.columnIndex] > 0 {
				m.rowIndex[m.columnIndex]--
			}
			return m, nil
		case "down", "j":
			if m.rowIndex[m.columnIndex] < len(m.columns[m.columnIndex].Tasks)-1 {
				m.rowIndex[m.columnIndex]++
			}
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
		case "s":
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

func (m boardModel) View() string {
	if len(m.columns) == 0 {
		return "No statuses configured.\n"
	}
	columnWidth := 24
	if m.width > 0 {
		columnWidth = max(20, (m.width-(len(m.columns)-1)*2)/len(m.columns))
	}
	activeBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("45")).
		Padding(0, 1).
		Width(columnWidth)
	idleBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(columnWidth)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	views := make([]string, 0, len(m.columns))
	for colIdx, column := range m.columns {
		lines := []string{titleStyle.Render(string(column.Status))}
		if len(column.Tasks) == 0 {
			lines = append(lines, mutedStyle.Render("(none)"))
		}
		for rowIdx, task := range column.Tasks {
			label := task.Title
			if m.showID {
				label = fmt.Sprintf("[%s] %s", shelf.ShortID(task.ID), label)
			}
			if rowIdx == m.rowIndex[colIdx] && colIdx == m.columnIndex {
				label = selectedStyle.Render("> " + label)
			} else {
				label = "  " + label
			}
			lines = append(lines, label)
		}
		style := idleBorder
		if colIdx == m.columnIndex {
			style = activeBorder
		}
		views = append(views, style.Render(strings.Join(lines, "\n")))
	}

	footer := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("h/l: column  j/k: move  o/i/s/b/d/c: status  r: reload  q: quit"),
	}
	if task, ok := m.selectedTask(); ok {
		footer = append(footer, lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(task.Title))
		body := strings.TrimSpace(task.Body)
		if body == "" {
			body = "(empty body)"
		}
		lines := strings.Split(body, "\n")
		if len(lines) > 4 {
			lines = lines[:4]
			lines = append(lines, "...")
		}
		footer = append(footer, strings.Join(lines, "\n"))
	}
	if strings.TrimSpace(m.message) != "" {
		footer = append(footer, lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Render(m.message))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, views...),
		"",
		strings.Join(footer, "\n"),
	) + "\n"
}

func (m *boardModel) reload() error {
	tasks, err := shelf.ListTasks(m.rootDir, shelf.TaskFilter{Limit: 0})
	if err != nil {
		return err
	}
	m.columns = buildBoardColumns(m.statuses, tasks)
	for i := range m.columns {
		if _, ok := m.rowIndex[i]; !ok {
			m.rowIndex[i] = 0
		}
	}
	m.clampRow()
	return nil
}

func buildBoardColumns(statuses []shelf.Status, tasks []shelf.Task) []boardColumn {
	grouped := map[shelf.Status][]shelf.Task{}
	for _, task := range tasks {
		grouped[task.Status] = append(grouped[task.Status], task)
	}
	columns := make([]boardColumn, 0, len(statuses))
	for _, status := range statuses {
		columns = append(columns, boardColumn{
			Status: status,
			Tasks:  grouped[status],
		})
	}
	return columns
}

func (m *boardModel) clampRow() {
	if len(m.columns) == 0 {
		m.columnIndex = 0
		return
	}
	if m.columnIndex < 0 {
		m.columnIndex = 0
	}
	if m.columnIndex >= len(m.columns) {
		m.columnIndex = len(m.columns) - 1
	}
	maxRow := len(m.columns[m.columnIndex].Tasks) - 1
	if maxRow < 0 {
		m.rowIndex[m.columnIndex] = 0
		return
	}
	if m.rowIndex[m.columnIndex] > maxRow {
		m.rowIndex[m.columnIndex] = maxRow
	}
	if m.rowIndex[m.columnIndex] < 0 {
		m.rowIndex[m.columnIndex] = 0
	}
}

func (m *boardModel) selectedTask() (shelf.Task, bool) {
	if len(m.columns) == 0 {
		return shelf.Task{}, false
	}
	column := m.columns[m.columnIndex]
	if len(column.Tasks) == 0 {
		return shelf.Task{}, false
	}
	row := m.rowIndex[m.columnIndex]
	if row < 0 || row >= len(column.Tasks) {
		return shelf.Task{}, false
	}
	return column.Tasks[row], true
}

func (m boardModel) applyStatusChange(nextStatus shelf.Status) (tea.Model, tea.Cmd) {
	task, ok := m.selectedTask()
	if !ok {
		m.message = "no task selected"
		return m, nil
	}
	err := withWriteLock(m.rootDir, func() error {
		if err := prepareUndoSnapshot(m.rootDir, "board-status"); err != nil {
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
