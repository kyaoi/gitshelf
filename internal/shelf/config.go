package shelf

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Kinds         []Kind
	Statuses      []Status
	Tags          []string
	LinkTypes     LinkTypesConfig
	DefaultKind   Kind
	DefaultStatus Status
	Commands      CommandsConfig
}

type LinkTypesConfig struct {
	Names    []LinkType
	Blocking LinkType
}

type CommandsConfig struct {
	Calendar CalendarCommandConfig
	Cockpit  CockpitCommandConfig
}

type CalendarCommandConfig struct {
	DefaultRangeUnit string
	DefaultDays      int
	DefaultMonths    int
	DefaultYears     int
}

type CockpitCommandConfig struct {
	CopySeparator     string
	PostExitGitAction string
	CommitMessage     string
}

type configFile struct {
	Kinds         []string       `toml:"kinds"`
	Statuses      []string       `toml:"statuses"`
	Tags          []string       `toml:"tags"`
	LinkTypes     toml.Primitive `toml:"link_types"`
	DefaultKind   string         `toml:"default_kind"`
	DefaultStatus string         `toml:"default_status"`
	Commands      configCommands `toml:"commands"`
}

type configLinkTypes struct {
	Names    []string `toml:"names"`
	Blocking string   `toml:"blocking"`
}

type configCommands struct {
	Calendar configCalendarCommand `toml:"calendar"`
	Cockpit  configCockpitCommand  `toml:"cockpit"`
}

type configCalendarCommand struct {
	DefaultRangeUnit string `toml:"default_range_unit"`
	DefaultDays      int    `toml:"default_days"`
	DefaultMonths    int    `toml:"default_months"`
	DefaultYears     int    `toml:"default_years"`
}

type configCockpitCommand struct {
	CopySeparator     string `toml:"copy_separator"`
	PostExitGitAction string `toml:"post_exit_git_action"`
	CommitMessage     string `toml:"commit_message"`
}

func DefaultConfig() Config {
	return Config{
		Kinds:    []Kind{"todo", "idea", "memo", "inbox"},
		Statuses: []Status{"open", "in_progress", "blocked", "done", "cancelled"},
		Tags:     []string{},
		LinkTypes: LinkTypesConfig{
			Names:    []LinkType{"depends_on", "related"},
			Blocking: "depends_on",
		},
		DefaultKind:   Kind("todo"),
		DefaultStatus: Status("open"),
		Commands: CommandsConfig{
			Calendar: CalendarCommandConfig{
				DefaultRangeUnit: "days",
				DefaultDays:      7,
				DefaultMonths:    6,
				DefaultYears:     2,
			},
			Cockpit: CockpitCommandConfig{
				CopySeparator:     "\n",
				PostExitGitAction: "none",
				CommitMessage:     "chore: update shelf data",
			},
		},
	}
}

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
	md, err := toml.Decode(string(data), &f)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse config TOML: %w", err)
	}

	defaults := DefaultConfig()
	cfg := Config{
		Kinds:         make([]Kind, len(f.Kinds)),
		Statuses:      make([]Status, len(f.Statuses)),
		Tags:          make([]string, len(f.Tags)),
		LinkTypes:     defaults.LinkTypes,
		DefaultKind:   Kind(strings.TrimSpace(f.DefaultKind)),
		DefaultStatus: Status(strings.TrimSpace(f.DefaultStatus)),
		Commands: CommandsConfig{
			Calendar: CalendarCommandConfig{
				DefaultRangeUnit: strings.TrimSpace(f.Commands.Calendar.DefaultRangeUnit),
				DefaultDays:      f.Commands.Calendar.DefaultDays,
				DefaultMonths:    f.Commands.Calendar.DefaultMonths,
				DefaultYears:     f.Commands.Calendar.DefaultYears,
			},
			Cockpit: CockpitCommandConfig{
				CopySeparator:     f.Commands.Cockpit.CopySeparator,
				PostExitGitAction: strings.TrimSpace(f.Commands.Cockpit.PostExitGitAction),
				CommitMessage:     strings.TrimSpace(f.Commands.Cockpit.CommitMessage),
			},
		},
	}
	if cfg.Commands.Calendar.DefaultRangeUnit == "" {
		cfg.Commands.Calendar.DefaultRangeUnit = defaults.Commands.Calendar.DefaultRangeUnit
	}
	if cfg.Commands.Calendar.DefaultDays == 0 {
		cfg.Commands.Calendar.DefaultDays = defaults.Commands.Calendar.DefaultDays
	}
	if cfg.Commands.Calendar.DefaultMonths == 0 {
		cfg.Commands.Calendar.DefaultMonths = defaults.Commands.Calendar.DefaultMonths
	}
	if cfg.Commands.Calendar.DefaultYears == 0 {
		cfg.Commands.Calendar.DefaultYears = defaults.Commands.Calendar.DefaultYears
	}
	if cfg.Commands.Cockpit.CopySeparator == "" {
		cfg.Commands.Cockpit.CopySeparator = defaults.Commands.Cockpit.CopySeparator
	}
	if cfg.Commands.Cockpit.PostExitGitAction == "" {
		cfg.Commands.Cockpit.PostExitGitAction = defaults.Commands.Cockpit.PostExitGitAction
	}
	if cfg.Commands.Cockpit.CommitMessage == "" {
		cfg.Commands.Cockpit.CommitMessage = defaults.Commands.Cockpit.CommitMessage
	}
	for i, kind := range f.Kinds {
		cfg.Kinds[i] = Kind(strings.TrimSpace(kind))
	}
	for i, status := range f.Statuses {
		cfg.Statuses[i] = Status(strings.TrimSpace(status))
	}
	for i, tag := range f.Tags {
		cfg.Tags[i] = strings.TrimSpace(tag)
	}
	linkTypes, err := parseConfigLinkTypes(md, f.LinkTypes, defaults.LinkTypes)
	if err != nil {
		return Config{}, err
	}
	cfg.LinkTypes = linkTypes
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func parseConfigLinkTypes(md toml.MetaData, primitive toml.Primitive, defaults LinkTypesConfig) (LinkTypesConfig, error) {
	if !md.IsDefined("link_types") {
		return defaults, nil
	}

	var legacy []string
	if err := md.PrimitiveDecode(primitive, &legacy); err == nil && len(legacy) > 0 {
		names := make([]LinkType, 0, len(legacy))
		for _, name := range legacy {
			names = append(names, LinkType(strings.TrimSpace(name)))
		}
		blocking := defaults.Blocking
		if !slices.Contains(names, blocking) && len(names) > 0 {
			blocking = names[0]
		}
		return LinkTypesConfig{
			Names:    names,
			Blocking: blocking,
		}, nil
	}

	var table configLinkTypes
	if err := md.PrimitiveDecode(primitive, &table); err != nil {
		return LinkTypesConfig{}, fmt.Errorf("failed to parse link_types: %w", err)
	}
	names := make([]LinkType, 0, len(table.Names))
	for _, name := range table.Names {
		names = append(names, LinkType(strings.TrimSpace(name)))
	}
	blocking := LinkType(strings.TrimSpace(table.Blocking))
	if blocking == "" {
		blocking = defaults.Blocking
	}
	return LinkTypesConfig{
		Names:    names,
		Blocking: blocking,
	}, nil
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
	buf.WriteString(fmt.Sprintf("default_kind = %q\n", cfg.DefaultKind))
	buf.WriteString(fmt.Sprintf("default_status = %q\n\n", cfg.DefaultStatus))

	buf.WriteString("[link_types]\n")
	buf.WriteString("names = [")
	for i, linkType := range cfg.LinkTypes.Names {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%q", linkType))
	}
	buf.WriteString("]\n")
	buf.WriteString(fmt.Sprintf("blocking = %q\n", cfg.LinkTypes.Blocking))
	buf.WriteString("\n[commands.calendar]\n")
	buf.WriteString(fmt.Sprintf("default_range_unit = %q\n", cfg.Commands.Calendar.DefaultRangeUnit))
	buf.WriteString(fmt.Sprintf("default_days = %d\n", cfg.Commands.Calendar.DefaultDays))
	buf.WriteString(fmt.Sprintf("default_months = %d\n", cfg.Commands.Calendar.DefaultMonths))
	buf.WriteString(fmt.Sprintf("default_years = %d\n", cfg.Commands.Calendar.DefaultYears))
	buf.WriteString("\n[commands.cockpit]\n")
	buf.WriteString(fmt.Sprintf("copy_separator = %q\n", cfg.Commands.Cockpit.CopySeparator))
	buf.WriteString(fmt.Sprintf("post_exit_git_action = %q\n", cfg.Commands.Cockpit.PostExitGitAction))
	buf.WriteString(fmt.Sprintf("commit_message = %q\n", cfg.Commands.Cockpit.CommitMessage))

	return buf.Bytes()
}

func (c Config) Validate() error {
	if len(c.Kinds) == 0 {
		return fmt.Errorf("config kinds is empty")
	}
	if len(c.Statuses) == 0 {
		return fmt.Errorf("config statuses is empty")
	}
	if len(c.LinkTypes.Names) == 0 {
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
	if err := validateLinkTypes(c.LinkTypes); err != nil {
		return err
	}
	if err := c.ValidateKind(c.DefaultKind); err != nil {
		return fmt.Errorf("default_kind: %w", err)
	}
	if err := c.ValidateStatus(c.DefaultStatus); err != nil {
		return fmt.Errorf("default_status: %w", err)
	}
	switch c.Commands.Calendar.DefaultRangeUnit {
	case "days", "months", "years":
	default:
		return fmt.Errorf("commands.calendar.default_range_unit must be one of days/months/years")
	}
	if c.Commands.Calendar.DefaultDays <= 0 {
		return fmt.Errorf("commands.calendar.default_days must be > 0")
	}
	if c.Commands.Calendar.DefaultMonths <= 0 {
		return fmt.Errorf("commands.calendar.default_months must be > 0")
	}
	if c.Commands.Calendar.DefaultYears <= 0 {
		return fmt.Errorf("commands.calendar.default_years must be > 0")
	}
	switch c.Commands.Cockpit.PostExitGitAction {
	case "none", "commit", "commit_push":
	default:
		return fmt.Errorf("commands.cockpit.post_exit_git_action must be one of none/commit/commit_push")
	}
	if strings.TrimSpace(c.Commands.Cockpit.CommitMessage) == "" {
		return fmt.Errorf("commands.cockpit.commit_message must not be empty")
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
	if slices.Contains(c.LinkTypes.Names, linkType) {
		return nil
	}
	return fmt.Errorf("unknown link type: %s", linkType)
}

func (c Config) BlockingLinkType() LinkType {
	return c.LinkTypes.Blocking
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

func validateLinkTypes(cfg LinkTypesConfig) error {
	seen := map[LinkType]struct{}{}
	for _, value := range cfg.Names {
		if value == "" {
			return fmt.Errorf("link_types must not include empty value")
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("link_types contains duplicate value: %s", value)
		}
		seen[value] = struct{}{}
	}
	if strings.TrimSpace(string(cfg.Blocking)) == "" {
		return fmt.Errorf("link_types.blocking must not be empty")
	}
	if _, ok := seen[cfg.Blocking]; !ok {
		return fmt.Errorf("link_types.blocking must be included in link_types.names: %s", cfg.Blocking)
	}
	return nil
}
