package shelf

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Kinds               []Kind
	Statuses            []Status
	Tags                []string
	LinkTypes           []LinkType
	DefaultKind         Kind
	DefaultStatus       Status
	CalendarDefaultDays int
	Views               map[string]TaskView
	OutputPresets       map[string]OutputPreset
}

type TaskView struct {
	Kinds       []Kind
	Statuses    []Status
	Tags        []string
	NotKinds    []Kind
	NotStatuses []Status
	NotTags     []string
	ReadyOnly   bool
	DepsBlocked bool
	DueBefore   string
	DueAfter    string
	Overdue     bool
	NoDue       bool
	Parent      string
	Search      string
	Limit       int
}

type OutputPreset struct {
	Command string `toml:"command"`
	Format  string `toml:"format"`
	View    string `toml:"view"`
	Limit   int    `toml:"limit"`
}

type configFile struct {
	Kinds               []string `toml:"kinds"`
	Statuses            []string `toml:"statuses"`
	Tags                []string `toml:"tags"`
	LegacyStates        []string `toml:"states"`
	LinkTypes           []string `toml:"link_types"`
	DefaultKind         string   `toml:"default_kind"`
	DefaultStatus       string   `toml:"default_status"`
	CalendarDefaultDays int      `toml:"calendar_default_days"`
	LegacyDefaultState  string   `toml:"default_state"`
	Views               map[string]configView
	OutputPresets       map[string]configOutputPreset `toml:"output_presets"`
}

type configView struct {
	Kinds       []string `toml:"kinds"`
	Statuses    []string `toml:"statuses"`
	Tags        []string `toml:"tags"`
	NotKinds    []string `toml:"not_kinds"`
	NotStatuses []string `toml:"not_statuses"`
	NotTags     []string `toml:"not_tags"`
	ReadyOnly   bool     `toml:"ready"`
	DepsBlocked bool     `toml:"blocked_by_deps"`
	DueBefore   string   `toml:"due_before"`
	DueAfter    string   `toml:"due_after"`
	Overdue     bool     `toml:"overdue"`
	NoDue       bool     `toml:"no_due"`
	Parent      string   `toml:"parent"`
	Search      string   `toml:"search"`
	Limit       int      `toml:"limit"`
}

type configOutputPreset struct {
	Command string `toml:"command"`
	Format  string `toml:"format"`
	View    string `toml:"view"`
	Limit   int    `toml:"limit"`
}

func DefaultConfig() Config {
	return Config{
		Kinds:               []Kind{"todo", "idea", "memo", "inbox"},
		Statuses:            []Status{"open", "in_progress", "blocked", "done", "cancelled"},
		Tags:                []string{},
		LinkTypes:           []LinkType{"depends_on", "related"},
		DefaultKind:         Kind("todo"),
		DefaultStatus:       Status("open"),
		CalendarDefaultDays: 7,
		Views:               map[string]TaskView{},
		OutputPresets:       map[string]OutputPreset{},
	}
}

var supportedLinkTypes = []LinkType{"depends_on", "related"}

func LoadConfig(rootDir string) (Config, error) {
	path := ConfigPath(rootDir)
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config %s: %w", path, err)
	}
	cfg, err := ParseConfigTOML(data)
	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}

func SaveConfig(rootDir string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	path := ConfigPath(rootDir)
	return atomicWriteFile(path, FormatConfigTOML(cfg), 0o644)
}

func ParseConfigTOML(data []byte) (Config, error) {
	var f configFile
	if _, err := toml.Decode(string(data), &f); err != nil {
		return Config{}, fmt.Errorf("failed to parse config TOML: %w", err)
	}

	statuses := f.Statuses
	if len(statuses) == 0 {
		statuses = f.LegacyStates
	}
	defaultStatus := strings.TrimSpace(f.DefaultStatus)
	if defaultStatus == "" {
		defaultStatus = strings.TrimSpace(f.LegacyDefaultState)
	}

	cfg := Config{
		Kinds:               make([]Kind, len(f.Kinds)),
		Statuses:            make([]Status, len(statuses)),
		Tags:                make([]string, len(f.Tags)),
		LinkTypes:           make([]LinkType, len(f.LinkTypes)),
		DefaultKind:         Kind(strings.TrimSpace(f.DefaultKind)),
		DefaultStatus:       Status(defaultStatus),
		CalendarDefaultDays: f.CalendarDefaultDays,
		Views:               map[string]TaskView{},
		OutputPresets:       map[string]OutputPreset{},
	}
	if cfg.CalendarDefaultDays == 0 {
		cfg.CalendarDefaultDays = DefaultConfig().CalendarDefaultDays
	}
	for i, kind := range f.Kinds {
		cfg.Kinds[i] = Kind(strings.TrimSpace(kind))
	}
	for i, status := range statuses {
		cfg.Statuses[i] = Status(strings.TrimSpace(status))
	}
	for i, tag := range f.Tags {
		cfg.Tags[i] = strings.TrimSpace(tag)
	}
	for i, linkType := range f.LinkTypes {
		cfg.LinkTypes[i] = LinkType(strings.TrimSpace(linkType))
	}
	for name, rawView := range f.Views {
		view := TaskView{
			Kinds:       make([]Kind, len(rawView.Kinds)),
			Statuses:    make([]Status, len(rawView.Statuses)),
			Tags:        make([]string, len(rawView.Tags)),
			NotKinds:    make([]Kind, len(rawView.NotKinds)),
			NotStatuses: make([]Status, len(rawView.NotStatuses)),
			NotTags:     make([]string, len(rawView.NotTags)),
			ReadyOnly:   rawView.ReadyOnly,
			DepsBlocked: rawView.DepsBlocked,
			DueBefore:   strings.TrimSpace(rawView.DueBefore),
			DueAfter:    strings.TrimSpace(rawView.DueAfter),
			Overdue:     rawView.Overdue,
			NoDue:       rawView.NoDue,
			Parent:      strings.TrimSpace(rawView.Parent),
			Search:      strings.TrimSpace(rawView.Search),
			Limit:       rawView.Limit,
		}
		for i, kind := range rawView.Kinds {
			view.Kinds[i] = Kind(strings.TrimSpace(kind))
		}
		for i, status := range rawView.Statuses {
			view.Statuses[i] = Status(strings.TrimSpace(status))
		}
		for i, tag := range rawView.Tags {
			view.Tags[i] = strings.TrimSpace(tag)
		}
		for i, kind := range rawView.NotKinds {
			view.NotKinds[i] = Kind(strings.TrimSpace(kind))
		}
		for i, status := range rawView.NotStatuses {
			view.NotStatuses[i] = Status(strings.TrimSpace(status))
		}
		for i, tag := range rawView.NotTags {
			view.NotTags[i] = strings.TrimSpace(tag)
		}
		cfg.Views[strings.TrimSpace(name)] = view
	}
	for name, rawPreset := range f.OutputPresets {
		cfg.OutputPresets[strings.TrimSpace(name)] = OutputPreset{
			Command: strings.TrimSpace(rawPreset.Command),
			Format:  strings.TrimSpace(rawPreset.Format),
			View:    strings.TrimSpace(rawPreset.View),
			Limit:   rawPreset.Limit,
		}
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func FormatConfigTOML(cfg Config) []byte {
	var buf bytes.Buffer
	buf.WriteString("# gitshelf config\n")
	buf.WriteString("kinds = [")
	for i, kind := range cfg.Kinds {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%q", kind))
	}
	buf.WriteString("]\n")

	buf.WriteString("statuses = [")
	for i, status := range cfg.Statuses {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%q", status))
	}
	buf.WriteString("]\n")

	buf.WriteString("tags = [")
	for i, tag := range cfg.Tags {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%q", tag))
	}
	buf.WriteString("]\n")

	buf.WriteString("link_types = [")
	for i, linkType := range cfg.LinkTypes {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%q", linkType))
	}
	buf.WriteString("]\n\n")
	buf.WriteString(fmt.Sprintf("default_kind = %q\n", cfg.DefaultKind))
	buf.WriteString(fmt.Sprintf("default_status = %q\n", cfg.DefaultStatus))
	buf.WriteString(fmt.Sprintf("calendar_default_days = %d\n", cfg.CalendarDefaultDays))

	if len(cfg.Views) > 0 {
		viewNames := make([]string, 0, len(cfg.Views))
		for name := range cfg.Views {
			viewNames = append(viewNames, name)
		}
		sort.Strings(viewNames)
		for _, name := range viewNames {
			view := cfg.Views[name]
			buf.WriteString("\n")
			buf.WriteString(fmt.Sprintf("[views.%q]\n", name))
			writeKinds := func(key string, values []Kind) {
				if len(values) == 0 {
					return
				}
				buf.WriteString(key + " = [")
				for i, value := range values {
					if i > 0 {
						buf.WriteString(", ")
					}
					buf.WriteString(fmt.Sprintf("%q", value))
				}
				buf.WriteString("]\n")
			}
			writeStatuses := func(key string, values []Status) {
				if len(values) == 0 {
					return
				}
				buf.WriteString(key + " = [")
				for i, value := range values {
					if i > 0 {
						buf.WriteString(", ")
					}
					buf.WriteString(fmt.Sprintf("%q", value))
				}
				buf.WriteString("]\n")
			}
			writeKinds("kinds", view.Kinds)
			writeStatuses("statuses", view.Statuses)
			writeTags := func(key string, values []string) {
				if len(values) == 0 {
					return
				}
				buf.WriteString(key + " = [")
				for i, value := range values {
					if i > 0 {
						buf.WriteString(", ")
					}
					buf.WriteString(fmt.Sprintf("%q", value))
				}
				buf.WriteString("]\n")
			}
			writeTags("tags", view.Tags)
			writeKinds("not_kinds", view.NotKinds)
			writeStatuses("not_statuses", view.NotStatuses)
			writeTags("not_tags", view.NotTags)
			if view.ReadyOnly {
				buf.WriteString("ready = true\n")
			}
			if view.DepsBlocked {
				buf.WriteString("blocked_by_deps = true\n")
			}
			if view.DueBefore != "" {
				buf.WriteString(fmt.Sprintf("due_before = %q\n", view.DueBefore))
			}
			if view.DueAfter != "" {
				buf.WriteString(fmt.Sprintf("due_after = %q\n", view.DueAfter))
			}
			if view.Overdue {
				buf.WriteString("overdue = true\n")
			}
			if view.NoDue {
				buf.WriteString("no_due = true\n")
			}
			if view.Parent != "" {
				buf.WriteString(fmt.Sprintf("parent = %q\n", view.Parent))
			}
			if view.Search != "" {
				buf.WriteString(fmt.Sprintf("search = %q\n", view.Search))
			}
			if view.Limit > 0 {
				buf.WriteString(fmt.Sprintf("limit = %d\n", view.Limit))
			}
		}
	}
	if len(cfg.OutputPresets) > 0 {
		names := make([]string, 0, len(cfg.OutputPresets))
		for name := range cfg.OutputPresets {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			preset := cfg.OutputPresets[name]
			buf.WriteString("\n")
			buf.WriteString(fmt.Sprintf("[output_presets.%q]\n", name))
			buf.WriteString(fmt.Sprintf("command = %q\n", preset.Command))
			if preset.Format != "" {
				buf.WriteString(fmt.Sprintf("format = %q\n", preset.Format))
			}
			if preset.View != "" {
				buf.WriteString(fmt.Sprintf("view = %q\n", preset.View))
			}
			if preset.Limit > 0 {
				buf.WriteString(fmt.Sprintf("limit = %d\n", preset.Limit))
			}
		}
	}
	return buf.Bytes()
}

func (c Config) Validate() error {
	if len(c.Kinds) == 0 {
		return fmt.Errorf("config kinds is empty")
	}
	if len(c.Statuses) == 0 {
		return fmt.Errorf("config statuses is empty")
	}
	if len(c.LinkTypes) == 0 {
		return fmt.Errorf("config link_types is empty")
	}
	if err := validateUniqueKinds(c.Kinds); err != nil {
		return err
	}
	if err := validateUniqueStatuses(c.Statuses); err != nil {
		return err
	}
	if err := validateUniqueTags(c.Tags); err != nil {
		return err
	}
	if err := validateUniqueLinkTypes(c.LinkTypes); err != nil {
		return err
	}
	if err := c.ValidateKind(c.DefaultKind); err != nil {
		return fmt.Errorf("default_kind: %w", err)
	}
	if err := c.ValidateStatus(c.DefaultStatus); err != nil {
		return fmt.Errorf("default_status: %w", err)
	}
	if c.CalendarDefaultDays <= 0 {
		return fmt.Errorf("calendar_default_days must be > 0")
	}
	for name, view := range c.Views {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("view name is required")
		}
		for _, kind := range view.Kinds {
			if err := c.ValidateKind(kind); err != nil {
				return fmt.Errorf("views.%s.kinds: %w", name, err)
			}
		}
		for _, kind := range view.NotKinds {
			if err := c.ValidateKind(kind); err != nil {
				return fmt.Errorf("views.%s.not_kinds: %w", name, err)
			}
		}
		for _, status := range view.Statuses {
			if err := c.ValidateStatus(status); err != nil {
				return fmt.Errorf("views.%s.statuses: %w", name, err)
			}
		}
		for _, tag := range view.Tags {
			if err := c.ValidateTag(tag); err != nil {
				return fmt.Errorf("views.%s.tags: %w", name, err)
			}
		}
		for _, status := range view.NotStatuses {
			if err := c.ValidateStatus(status); err != nil {
				return fmt.Errorf("views.%s.not_statuses: %w", name, err)
			}
		}
		for _, tag := range view.NotTags {
			if err := c.ValidateTag(tag); err != nil {
				return fmt.Errorf("views.%s.not_tags: %w", name, err)
			}
		}
		if view.ReadyOnly && view.DepsBlocked {
			return fmt.Errorf("views.%s: ready and blocked_by_deps cannot be true together", name)
		}
		if view.NoDue && (view.DueBefore != "" || view.DueAfter != "" || view.Overdue) {
			return fmt.Errorf("views.%s: no_due cannot be combined with due_before/due_after/overdue", name)
		}
		if view.DueBefore != "" {
			if _, err := NormalizeDueOn(view.DueBefore); err != nil {
				return fmt.Errorf("views.%s.due_before: %w", name, err)
			}
		}
		if view.DueAfter != "" {
			if _, err := NormalizeDueOn(view.DueAfter); err != nil {
				return fmt.Errorf("views.%s.due_after: %w", name, err)
			}
		}
		if view.Limit < 0 {
			return fmt.Errorf("views.%s.limit must be >= 0", name)
		}
	}
	for name, preset := range c.OutputPresets {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("output preset name is required")
		}
		command := strings.TrimSpace(preset.Command)
		switch command {
		case "ls", "tree", "next", "agenda", "today":
		default:
			return fmt.Errorf("output_presets.%s.command: unsupported command %q", name, preset.Command)
		}
		if strings.TrimSpace(preset.Format) != "" {
			switch command {
			case "ls":
				if preset.Format != "compact" && preset.Format != "detail" && preset.Format != "kanban" {
					return fmt.Errorf("output_presets.%s.format: invalid format for ls", name)
				}
			case "tree":
				if preset.Format != "compact" && preset.Format != "detail" {
					return fmt.Errorf("output_presets.%s.format: invalid format for tree", name)
				}
			case "agenda", "today":
				if preset.Format != "compact" && preset.Format != "detail" {
					return fmt.Errorf("output_presets.%s.format: invalid format for %s", name, command)
				}
			case "next":
				return fmt.Errorf("output_presets.%s.format: next does not support format", name)
			}
		}
		if preset.View != "" {
			if _, ok := c.Views[preset.View]; !ok {
				if preset.View != "active" && preset.View != "ready" && preset.View != "blocked" && preset.View != "overdue" {
					return fmt.Errorf("output_presets.%s.view: unknown view %q", name, preset.View)
				}
			}
		}
		if preset.Limit < 0 {
			return fmt.Errorf("output_presets.%s.limit must be >= 0", name)
		}
	}
	return nil
}

func (c Config) ValidateKind(kind Kind) error {
	if strings.TrimSpace(string(kind)) == "" {
		return fmt.Errorf("kind is required")
	}
	if slices.Contains(c.Kinds, kind) {
		return nil
	}
	return fmt.Errorf("unknown kind: %s", kind)
}

func (c Config) ValidateStatus(status Status) error {
	if strings.TrimSpace(string(status)) == "" {
		return fmt.Errorf("status is required")
	}
	if slices.Contains(c.Statuses, status) {
		return nil
	}
	return fmt.Errorf("unknown status: %s", status)
}

func (c Config) ValidateTag(tag string) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return fmt.Errorf("tag is required")
	}
	if slices.Contains(c.Tags, tag) {
		return nil
	}
	return fmt.Errorf("unknown tag: %s", tag)
}

func (c Config) ValidateLinkType(linkType LinkType) error {
	if strings.TrimSpace(string(linkType)) == "" {
		return fmt.Errorf("link type is required")
	}
	if slices.Contains(c.LinkTypes, linkType) {
		return nil
	}
	return fmt.Errorf("unknown link type: %s", linkType)
}

func (c *Config) AppendMissingTags(tags []string) bool {
	candidate := NormalizeTags(tags)
	if len(candidate) == 0 {
		return false
	}
	changed := false
	for _, tag := range candidate {
		if slices.Contains(c.Tags, tag) {
			continue
		}
		c.Tags = append(c.Tags, tag)
		changed = true
	}
	return changed
}

func validateUniqueKinds(values []Kind) error {
	seen := map[Kind]struct{}{}
	for _, value := range values {
		if value == "" {
			return fmt.Errorf("kinds must not include empty value")
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("kinds contains duplicate value: %s", value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateUniqueStatuses(values []Status) error {
	seen := map[Status]struct{}{}
	for _, value := range values {
		if value == "" {
			return fmt.Errorf("statuses must not include empty value")
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("statuses contains duplicate value: %s", value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateUniqueTags(values []string) error {
	seen := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("tags must not include empty value")
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("tags contains duplicate value: %s", value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateUniqueLinkTypes(values []LinkType) error {
	seen := map[LinkType]struct{}{}
	for _, value := range values {
		if value == "" {
			return fmt.Errorf("link_types must not include empty value")
		}
		if !slices.Contains(supportedLinkTypes, value) {
			return fmt.Errorf("unsupported link type: %s (allowed: depends_on, related)", value)
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("link_types contains duplicate value: %s", value)
		}
		seen[value] = struct{}{}
	}
	return nil
}
