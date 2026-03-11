package cli

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

type calendarCopyPresetFocus int

const (
	calendarCopyPresetFocusPresets calendarCopyPresetFocus = iota
	calendarCopyPresetFocusName
	calendarCopyPresetFocusScope
	calendarCopyPresetFocusSubtreeStyle
	calendarCopyPresetFocusTemplate
	calendarCopyPresetFocusJoinWith
)

func defaultCustomCopyPreset() shelf.CopyPreset {
	return shelf.CopyPreset{
		Name:         "subtree-path",
		Scope:        shelf.CopyPresetScopeSubtree,
		SubtreeStyle: shelf.CopySubtreeStyleIndented,
		Template:     "{{path}}\n{{subtree}}",
		JoinWith:     "\n\n",
	}
}

func encodeCopyPresetEscapes(value string) string {
	var b strings.Builder
	for _, r := range value {
		switch r {
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\\':
			b.WriteString(`\\`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func decodeCopyPresetEscapes(value string) string {
	var b strings.Builder
	runes := []rune(value)
	for i := 0; i < len(runes); i++ {
		if runes[i] != '\\' || i == len(runes)-1 {
			b.WriteRune(runes[i])
			continue
		}
		i++
		switch runes[i] {
		case 'n':
			b.WriteRune('\n')
		case 't':
			b.WriteRune('\t')
		case '\\':
			b.WriteRune('\\')
		default:
			b.WriteRune('\\')
			b.WriteRune(runes[i])
		}
	}
	return b.String()
}

func shellQuoteANSI(value string) string {
	var b strings.Builder
	b.WriteString("$'")
	for _, r := range value {
		switch r {
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteString("'")
	return b.String()
}

func (m *calendarTUIModel) resetCopyPresetDraft() {
	preset := defaultCustomCopyPreset()
	m.copyPresetName = preset.Name
	m.copyPresetNameCursor = len([]rune(m.copyPresetName))
	m.copyPresetScope = preset.Scope
	m.copyPresetSubtreeStyle = preset.EffectiveSubtreeStyle()
	m.copyPresetTemplate = encodeCopyPresetEscapes(preset.Template)
	m.copyPresetTemplateCursor = len([]rune(m.copyPresetTemplate))
	m.copyPresetJoinWith = encodeCopyPresetEscapes(preset.JoinWith)
	m.copyPresetJoinWithCursor = len([]rune(m.copyPresetJoinWith))
}

func (m *calendarTUIModel) beginCopyPresetMode() {
	m.copyPresetMode = true
	m.copyPresetFocus = calendarCopyPresetFocusPresets
	m.copyPresetIndex = 0
	m.resetCopyPresetDraft()
	m.message = "advanced copy"
}

func (m *calendarTUIModel) activeCopyPreset() shelf.CopyPreset {
	if m.copyPresetIndex <= 0 || m.copyPresetIndex > len(m.copyPresets) {
		return m.customCopyPreset()
	}
	return m.copyPresets[m.copyPresetIndex-1]
}

func (m calendarTUIModel) customCopyPreset() shelf.CopyPreset {
	return shelf.CopyPreset{
		Name:         strings.TrimSpace(m.copyPresetName),
		Scope:        m.copyPresetScope,
		SubtreeStyle: m.copyPresetSubtreeStyle,
		Template:     decodeCopyPresetEscapes(m.copyPresetTemplate),
		JoinWith:     decodeCopyPresetEscapes(m.copyPresetJoinWith),
	}
}

func (m *calendarTUIModel) ensureCustomCopyPresetSelected() {
	if m.copyPresetIndex <= 0 || m.copyPresetIndex > len(m.copyPresets) {
		return
	}
	selected := m.copyPresets[m.copyPresetIndex-1]
	m.copyPresetName = selected.Name
	m.copyPresetNameCursor = len([]rune(m.copyPresetName))
	m.copyPresetScope = selected.Scope
	m.copyPresetSubtreeStyle = selected.EffectiveSubtreeStyle()
	m.copyPresetTemplate = encodeCopyPresetEscapes(selected.Template)
	m.copyPresetTemplateCursor = len([]rune(m.copyPresetTemplate))
	m.copyPresetJoinWith = encodeCopyPresetEscapes(selected.JoinWith)
	m.copyPresetJoinWithCursor = len([]rune(m.copyPresetJoinWith))
	m.copyPresetIndex = 0
}

func (m *calendarTUIModel) cycleCopyPresetFocus(delta int) {
	const count = int(calendarCopyPresetFocusJoinWith) + 1
	next := (int(m.copyPresetFocus) + delta + count) % count
	m.copyPresetFocus = calendarCopyPresetFocus(next)
}

func (m *calendarTUIModel) moveCopyPresetSelection(delta int) {
	maxIndex := len(m.copyPresets)
	next := m.copyPresetIndex + delta
	if next < 0 {
		next = 0
	}
	if next > maxIndex {
		next = maxIndex
	}
	m.copyPresetIndex = next
}

func (m calendarTUIModel) updateCopyPresetMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "ctrl+[":
		m.copyPresetMode = false
		m.message = "advanced copy をキャンセルしました"
		return m, nil
	case "enter":
		if err := m.copySelectedWithPreset(m.activeCopyPreset()); err != nil {
			m.message = err.Error()
		}
		return m, nil
	case "ctrl+s":
		if err := m.saveActiveCopyPreset(); err != nil {
			m.message = err.Error()
		}
		return m, nil
	case "tab":
		m.cycleCopyPresetFocus(1)
		return m, nil
	case "shift+tab":
		m.cycleCopyPresetFocus(-1)
		return m, nil
	}

	switch m.copyPresetFocus {
	case calendarCopyPresetFocusPresets:
		switch msg.String() {
		case "up", "k":
			m.moveCopyPresetSelection(-1)
			return m, nil
		case "down", "j":
			m.moveCopyPresetSelection(1)
			return m, nil
		}
	case calendarCopyPresetFocusScope:
		switch msg.String() {
		case "left", "h", "up", "k", "right", "l", "down", "j", " ":
			m.ensureCustomCopyPresetSelected()
			if m.copyPresetScope == shelf.CopyPresetScopeSubtree {
				m.copyPresetScope = shelf.CopyPresetScopeTask
			} else {
				m.copyPresetScope = shelf.CopyPresetScopeSubtree
			}
			return m, nil
		}
	case calendarCopyPresetFocusSubtreeStyle:
		switch msg.String() {
		case "left", "h", "up", "k", "right", "l", "down", "j", " ":
			m.ensureCustomCopyPresetSelected()
			if m.copyPresetSubtreeStyle == shelf.CopySubtreeStyleTree {
				m.copyPresetSubtreeStyle = shelf.CopySubtreeStyleIndented
			} else {
				m.copyPresetSubtreeStyle = shelf.CopySubtreeStyleTree
			}
			return m, nil
		}
	case calendarCopyPresetFocusName:
		return m.updateCopyPresetTextField(msg, &m.copyPresetName, &m.copyPresetNameCursor)
	case calendarCopyPresetFocusTemplate:
		return m.updateCopyPresetTextField(msg, &m.copyPresetTemplate, &m.copyPresetTemplateCursor)
	case calendarCopyPresetFocusJoinWith:
		return m.updateCopyPresetTextField(msg, &m.copyPresetJoinWith, &m.copyPresetJoinWithCursor)
	}
	return m, nil
}

func (m *calendarTUIModel) updateCopyPresetTextField(msg tea.KeyMsg, value *string, cursor *int) (tea.Model, tea.Cmd) {
	m.ensureCustomCopyPresetSelected()
	switch msg.String() {
	case "backspace":
		*value, *cursor = interactive.DeleteRuneBeforeCursor(*value, *cursor)
	case "left":
		*cursor = interactive.MoveTextCursorLeft(*value, *cursor)
	case "right":
		*cursor = interactive.MoveTextCursorRight(*value, *cursor)
	default:
		if msg.Type == tea.KeyRunes {
			for _, r := range msg.Runes {
				*value, *cursor = interactive.InsertRuneAtCursor(*value, *cursor, r)
			}
		}
	}
	return m, nil
}

func (m *calendarTUIModel) copySelectedWithPreset(preset shelf.CopyPreset) error {
	text, count, err := m.renderCopyPresetPayload(preset)
	if err != nil {
		return err
	}
	return m.copyClipboardPayload(text, count, "item", "items")
}

func (m *calendarTUIModel) saveActiveCopyPreset() error {
	preset := m.activeCopyPreset()
	if err := shelf.ValidateCopyPreset(preset); err != nil {
		return err
	}
	cfg, err := shelf.LoadConfig(m.rootDir)
	if err != nil {
		return err
	}
	updated, err := cfg.UpsertCopyPreset(preset)
	if err != nil {
		return err
	}
	if err := shelf.SaveConfig(m.rootDir, cfg); err != nil {
		return err
	}
	m.copyPresets = append([]shelf.CopyPreset{}, cfg.Commands.Cockpit.CopyPresets...)
	for i, candidate := range m.copyPresets {
		if candidate.Name == preset.Name {
			m.copyPresetIndex = i + 1
			break
		}
	}
	if updated {
		m.message = fmt.Sprintf("copy preset を更新しました: %s", preset.Name)
	} else {
		m.message = fmt.Sprintf("copy preset を保存しました: %s", preset.Name)
	}
	return nil
}

func (m calendarTUIModel) renderCopyPresetPayload(preset shelf.CopyPreset) (string, int, error) {
	if err := shelf.ValidateCopyPreset(preset); err != nil {
		return "", 0, err
	}
	taskIDs := m.activeTaskIDs()
	if preset.Scope == shelf.CopyPresetScopeSubtree {
		taskIDs = m.copySubtreeRootIDs()
	}
	if len(taskIDs) == 0 {
		return "", 0, fmt.Errorf("選択中の task がありません")
	}
	items := make([]string, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		task, ok := m.taskByID[taskID]
		if !ok {
			continue
		}
		subtreeText, _ := m.renderTaskSubtreeText(taskID, preset.EffectiveSubtreeStyle())
		rendered := preset.Template
		replacements := map[string]string{
			"{{title}}":   task.Title,
			"{{path}}":    filepath.Join(shelf.TasksDir(m.rootDir), taskID+".md"),
			"{{body}}":    task.Body,
			"{{subtree}}": subtreeText,
		}
		for _, placeholder := range shelf.SupportedCopyTemplatePlaceholders() {
			rendered = strings.ReplaceAll(rendered, placeholder, replacements[placeholder])
		}
		if strings.TrimSpace(rendered) == "" {
			continue
		}
		items = append(items, rendered)
	}
	if len(items) == 0 {
		return "", 0, fmt.Errorf("コピー対象がありません")
	}
	return strings.Join(items, preset.EffectiveJoinWith(m.copySeparator)), len(items), nil
}

func (m calendarTUIModel) renderTaskSubtreeText(rootID string, style shelf.CopySubtreeStyle) (string, int) {
	if strings.TrimSpace(rootID) == "" {
		return "", 0
	}
	byParent := make(map[string][]shelf.Task)
	for _, task := range m.allTasks {
		byParent[task.Parent] = append(byParent[task.Parent], task)
	}
	for parentID := range byParent {
		children := byParent[parentID]
		sortTasksForCopy(children)
		byParent[parentID] = children
	}

	lines := []string{}
	count := 0
	var appendIndented func(taskID string, depth int)
	appendIndented = func(taskID string, depth int) {
		task, ok := m.taskByID[taskID]
		if !ok {
			return
		}
		lines = append(lines, strings.Repeat("  ", depth)+task.Title)
		count++
		for _, child := range byParent[taskID] {
			appendIndented(child.ID, depth+1)
		}
	}
	var appendTree func(taskID string, prefix string, isLast bool, isRoot bool)
	appendTree = func(taskID string, prefix string, isLast bool, isRoot bool) {
		task, ok := m.taskByID[taskID]
		if !ok {
			return
		}
		label := task.Title
		if !isRoot {
			connector := "|- "
			if isLast {
				connector = "`- "
			}
			label = prefix + connector + task.Title
		}
		lines = append(lines, label)
		count++
		children := byParent[taskID]
		nextPrefix := prefix
		if !isRoot {
			if isLast {
				nextPrefix += "   "
			} else {
				nextPrefix += "|  "
			}
		}
		for i, child := range children {
			appendTree(child.ID, nextPrefix, i == len(children)-1, false)
		}
	}
	if style == shelf.CopySubtreeStyleTree {
		appendTree(rootID, "", true, true)
	} else {
		appendIndented(rootID, 0)
	}
	return strings.Join(lines, "\n"), count
}

func sortTasksForCopy(tasks []shelf.Task) {
	slices.SortFunc(tasks, func(left shelf.Task, right shelf.Task) int {
		if left.Title != right.Title {
			if left.Title < right.Title {
				return -1
			}
			return 1
		}
		if left.ID < right.ID {
			return -1
		}
		if left.ID > right.ID {
			return 1
		}
		return 0
	})
}

func (m calendarTUIModel) copyPresetSaveCommand(preset shelf.CopyPreset) string {
	command := []string{
		"shelf",
		"config",
		"copy-preset",
		"set",
		"--name", shellQuoteANSI(preset.Name),
		"--scope", shellQuoteANSI(string(preset.Scope)),
		"--subtree-style", shellQuoteANSI(string(preset.EffectiveSubtreeStyle())),
		"--template", shellQuoteANSI(preset.Template),
	}
	if preset.JoinWith != "" {
		command = append(command, "--join-with", shellQuoteANSI(preset.JoinWith))
	}
	return strings.Join(command, " ")
}

func (m calendarTUIModel) copyPresetPreviewLines(width int, maxLines int) []string {
	text, _, err := m.renderCopyPresetPayload(m.activeCopyPreset())
	if err != nil {
		return []string{trimLine("(preview unavailable: "+err.Error()+")", width)}
	}
	raw := strings.Split(text, "\n")
	if len(raw) > maxLines {
		raw = raw[:maxLines]
		raw = append(raw, "…")
	}
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		lines = append(lines, trimLine(line, width))
	}
	return lines
}

func (m calendarTUIModel) quickCopyLegendLines(width int) []string {
	type copyAction struct {
		key   string
		label string
		count func() (int, error)
	}
	actions := []copyAction{
		{key: "y", label: "titles", count: func() (int, error) { _, count, err := m.selectedTitleCopyText(); return count, err }},
		{key: "Y", label: "subtrees", count: func() (int, error) { _, count, err := m.selectedSubtreeCopyText(); return count, err }},
		{key: "P", label: "paths", count: func() (int, error) { _, count, err := m.selectedPathCopyText(); return count, err }},
		{key: "O", label: "bodies", count: func() (int, error) { _, count, err := m.selectedBodyCopyText(); return count, err }},
		{key: "M", label: "advanced preset preview", count: func() (int, error) {
			_, count, err := m.renderCopyPresetPayload(m.activeCopyPreset())
			return count, err
		}},
	}
	lines := make([]string, 0, len(actions))
	for _, action := range actions {
		count, err := action.count()
		suffix := ""
		if err != nil {
			suffix = " (unavailable)"
		} else {
			suffix = fmt.Sprintf(" (%d item", count)
			if count != 1 {
				suffix += "s"
			}
			suffix += ")"
		}
		lines = append(lines, trimLine(fmt.Sprintf("  %s  %s%s", action.key, action.label, suffix), width))
	}
	return lines
}

func renderCalendarCopyPresetPopup(m calendarTUIModel, width int, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("81"))

	header := []string{
		titleStyle.Render("Advanced Copy"),
		helpStyle.Render(popupControls("j/k: move", "Tab: focus", "Enter: copy", "Ctrl+S: save", "Esc/q: close")),
		helpStyle.Render("Target: " + m.bulkActionPopupLabel()),
		helpStyle.Render("template uses {{title}} {{path}} {{body}} {{subtree}}; enter \\n in fields for newlines"),
	}

	active := m.activeCopyPreset()
	templateValue := encodeCopyPresetEscapes(active.Template)
	joinValue := encodeCopyPresetEscapes(active.JoinWith)
	nameValue := active.Name
	scopeValue := string(active.Scope)
	subtreeStyleValue := string(active.EffectiveSubtreeStyle())
	if m.copyPresetIndex == 0 {
		if m.copyPresetFocus == calendarCopyPresetFocusName {
			nameValue = interactive.RenderTextCursor(m.copyPresetName, m.copyPresetNameCursor)
		}
		if m.copyPresetFocus == calendarCopyPresetFocusTemplate {
			templateValue = interactive.RenderTextCursor(m.copyPresetTemplate, m.copyPresetTemplateCursor)
		}
		if m.copyPresetFocus == calendarCopyPresetFocusJoinWith {
			joinValue = interactive.RenderTextCursor(m.copyPresetJoinWith, m.copyPresetJoinWithCursor)
		}
		if m.copyPresetFocus == calendarCopyPresetFocusSubtreeStyle {
			subtreeStyleValue = string(m.copyPresetSubtreeStyle)
		}
	}

	lines := []string{
		labelStyle.Render("Presets"),
	}
	presetLabels := []string{"Custom"}
	for _, preset := range m.copyPresets {
		presetLabels = append(presetLabels, preset.Name)
	}
	for i, label := range presetLabels {
		line := "  " + label
		if m.copyPresetFocus == calendarCopyPresetFocusPresets && i == m.copyPresetIndex {
			line = selectedStyle.Render("> " + label)
		} else if i == m.copyPresetIndex {
			line = "• " + label
		}
		lines = append(lines, trimLine(line, max(24, width-6)))
	}

	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Custom Preset"))
	lines = append(lines, renderCopyPresetField("Name", nameValue, m.copyPresetFocus == calendarCopyPresetFocusName))
	scopeLine := scopeValue
	if scopeLine == "" {
		scopeLine = string(shelf.CopyPresetScopeSubtree)
	}
	lines = append(lines, renderCopyPresetField("Scope", scopeLine, m.copyPresetFocus == calendarCopyPresetFocusScope))
	lines = append(lines, renderCopyPresetField("Subtree Style", subtreeStyleValue, m.copyPresetFocus == calendarCopyPresetFocusSubtreeStyle))
	lines = append(lines, renderCopyPresetField("Template", templateValue, m.copyPresetFocus == calendarCopyPresetFocusTemplate))
	lines = append(lines, renderCopyPresetField("Join With", joinValue, m.copyPresetFocus == calendarCopyPresetFocusJoinWith))

	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Quick Copy"))
	lines = append(lines, m.quickCopyLegendLines(max(24, width-8))...)

	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Preview"))
	lines = append(lines, m.copyPresetPreviewLines(max(24, width-8), 6)...)

	lines = append(lines, "")
	lines = append(lines, labelStyle.Render("Save Command"))
	command := m.copyPresetSaveCommand(active)
	if err := shelf.ValidateCopyPreset(active); err != nil {
		command = "(invalid preset: " + err.Error() + ")"
	}
	lines = append(lines, helpStyle.Render(trimLine(command, max(24, width-8))))

	return renderPopupBoxWithHeader(header, lines, width, height, lipgloss.Color("141"), 1+m.copyPresetIndex)
}

func renderCopyPresetField(label string, value string, focused bool) string {
	line := fmt.Sprintf("  %s: %s", label, value)
	if focused {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true).Render("> " + strings.TrimPrefix(line, "  "))
	}
	return line
}
