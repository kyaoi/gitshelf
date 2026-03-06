package cli

import (
	"fmt"
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
}

type calendarMonthView struct {
	Label string
	Weeks [][]calendarMonthCell
}

type calendarTUIModel struct {
	days     []calendarDay
	showID   bool
	dayIndex int
	width    int
	height   int
}

func runCalendarTUI(days []calendarDay, showID bool) error {
	model := newCalendarTUIModel(days, showID)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

func newCalendarTUIModel(days []calendarDay, showID bool) calendarTUIModel {
	return calendarTUIModel{
		days:     append([]calendarDay{}, days...),
		showID:   showID,
		dayIndex: 0,
	}
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
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "left", "h":
			if m.dayIndex > 0 {
				m.dayIndex--
			}
			return m, nil
		case "right", "l":
			if m.dayIndex < len(m.days)-1 {
				m.dayIndex++
			}
			return m, nil
		case "up", "k":
			m.dayIndex = max(0, m.dayIndex-7)
			return m, nil
		case "down", "j":
			if len(m.days) == 0 {
				return m, nil
			}
			m.dayIndex = min(len(m.days)-1, m.dayIndex+7)
			return m, nil
		case "g":
			m.dayIndex = 0
			return m, nil
		case "G":
			if len(m.days) > 0 {
				m.dayIndex = len(m.days) - 1
			}
			return m, nil
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
		helpStyle.Render("h/l: 日移動  j/k: 週移動  g/G: 先頭/末尾  q: quit"),
		subStyle.Render(fmt.Sprintf("Focused: %s", focusedDate.Format("Mon 2006-01-02"))),
	}

	parts := []string{
		strings.Join(header, "\n"),
		renderCalendarMonth(month, focused.Date, m.width),
		renderCalendarDayDetails(focused, m.showID),
	}

	return strings.Join(parts, "\n\n") + "\n"
}

func buildCalendarMonthView(days []calendarDay, focusDate time.Time) calendarMonthView {
	taskCounts := map[string]int{}
	inRange := map[string]struct{}{}
	for _, day := range days {
		taskCounts[day.Date] = len(day.Tasks)
		inRange[day.Date] = struct{}{}
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

	contentWidth := max(6, cellWidth-2)
	style := lipgloss.NewStyle().
		Width(contentWidth).
		Height(3).
		Padding(0, 1)

	switch {
	case key == focusedDate:
		style = style.Background(lipgloss.Color("25")).Foreground(lipgloss.Color("255")).Bold(true)
	case cell.InRange && cell.TaskCount > 0:
		style = style.Background(lipgloss.Color("237")).Foreground(lipgloss.Color("230"))
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

func renderCalendarDayDetails(day calendarDay, showID bool) string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	lines := []string{
		titleStyle.Render(fmt.Sprintf("%s (%d due)", day.Date, len(day.Tasks))),
	}
	if len(day.Tasks) == 0 {
		lines = append(lines, mutedStyle.Render("(no due tasks)"))
		return boxStyle.Render(strings.Join(lines, "\n"))
	}

	for _, task := range day.Tasks {
		label := task.Title
		if showID {
			label = fmt.Sprintf("[%s] %s", shelf.ShortID(task.ID), label)
		}
		lines = append(lines, fmt.Sprintf("• %s (%s/%s)", label, task.Kind, task.Status))
		body := strings.TrimSpace(task.Body)
		if body == "" {
			continue
		}
		firstLine := strings.Split(body, "\n")[0]
		lines = append(lines, mutedStyle.Render("  "+trimLine(firstLine, 96)))
	}

	return boxStyle.Render(strings.Join(lines, "\n"))
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
