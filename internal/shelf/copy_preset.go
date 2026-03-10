package shelf

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type CopyPresetScope string

const (
	CopyPresetScopeTask    CopyPresetScope = "task"
	CopyPresetScopeSubtree CopyPresetScope = "subtree"
)

type CopyPreset struct {
	Name     string
	Scope    CopyPresetScope
	Template string
	JoinWith string
}

var (
	copyTemplatePlaceholderPattern = regexp.MustCompile(`\{\{[^}]+\}\}`)
	copyTemplatePlaceholders       = []string{"{{title}}", "{{path}}", "{{body}}", "{{subtree}}"}
)

func SupportedCopyTemplatePlaceholders() []string {
	return append([]string{}, copyTemplatePlaceholders...)
}

func ValidateCopyPreset(preset CopyPreset) error {
	if strings.TrimSpace(preset.Name) == "" {
		return fmt.Errorf("copy preset name is required")
	}
	switch CopyPresetScope(strings.TrimSpace(string(preset.Scope))) {
	case CopyPresetScopeTask, CopyPresetScopeSubtree:
	default:
		return fmt.Errorf("copy preset %q scope must be one of task/subtree", preset.Name)
	}
	if err := ValidateCopyTemplate(preset.Template); err != nil {
		return fmt.Errorf("copy preset %q: %w", preset.Name, err)
	}
	return nil
}

func ValidateCopyTemplate(template string) error {
	if strings.TrimSpace(template) == "" {
		return fmt.Errorf("copy template must not be empty")
	}
	for _, placeholder := range copyTemplatePlaceholderPattern.FindAllString(template, -1) {
		if slices.Contains(copyTemplatePlaceholders, placeholder) {
			continue
		}
		return fmt.Errorf("unsupported copy template placeholder: %s", placeholder)
	}
	return nil
}

func (preset CopyPreset) EffectiveJoinWith(defaultSeparator string) string {
	if preset.JoinWith != "" {
		return preset.JoinWith
	}
	return defaultSeparator
}

func (c *Config) UpsertCopyPreset(preset CopyPreset) (bool, error) {
	if err := ValidateCopyPreset(preset); err != nil {
		return false, err
	}
	for i := range c.Commands.Cockpit.CopyPresets {
		if c.Commands.Cockpit.CopyPresets[i].Name != preset.Name {
			continue
		}
		c.Commands.Cockpit.CopyPresets[i] = preset
		return true, nil
	}
	c.Commands.Cockpit.CopyPresets = append(c.Commands.Cockpit.CopyPresets, preset)
	return false, nil
}
