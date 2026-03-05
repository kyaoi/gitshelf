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
	Kinds        []Kind
	States       []State
	LinkTypes    []LinkType
	DefaultKind  Kind
	DefaultState State
}

type configFile struct {
	Kinds        []string `toml:"kinds"`
	States       []string `toml:"states"`
	LinkTypes    []string `toml:"link_types"`
	DefaultKind  string   `toml:"default_kind"`
	DefaultState string   `toml:"default_state"`
}

func DefaultConfig() Config {
	return Config{
		Kinds:        []Kind{"todo", "idea", "memo"},
		States:       []State{"open", "done"},
		LinkTypes:    []LinkType{"derived_from", "depends_on", "related"},
		DefaultKind:  Kind("todo"),
		DefaultState: State("open"),
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
	if _, err := toml.Decode(string(data), &f); err != nil {
		return Config{}, fmt.Errorf("failed to parse config TOML: %w", err)
	}

	cfg := Config{
		Kinds:        make([]Kind, len(f.Kinds)),
		States:       make([]State, len(f.States)),
		LinkTypes:    make([]LinkType, len(f.LinkTypes)),
		DefaultKind:  Kind(strings.TrimSpace(f.DefaultKind)),
		DefaultState: State(strings.TrimSpace(f.DefaultState)),
	}
	for i, kind := range f.Kinds {
		cfg.Kinds[i] = Kind(strings.TrimSpace(kind))
	}
	for i, state := range f.States {
		cfg.States[i] = State(strings.TrimSpace(state))
	}
	for i, linkType := range f.LinkTypes {
		cfg.LinkTypes[i] = LinkType(strings.TrimSpace(linkType))
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

	buf.WriteString("states = [")
	for i, state := range cfg.States {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%q", state))
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
	buf.WriteString(fmt.Sprintf("default_state = %q\n", cfg.DefaultState))
	return buf.Bytes()
}

func (c Config) Validate() error {
	if len(c.Kinds) == 0 {
		return fmt.Errorf("config kinds is empty")
	}
	if len(c.States) == 0 {
		return fmt.Errorf("config states is empty")
	}
	if len(c.LinkTypes) == 0 {
		return fmt.Errorf("config link_types is empty")
	}
	if err := validateUniqueKinds(c.Kinds); err != nil {
		return err
	}
	if err := validateUniqueStates(c.States); err != nil {
		return err
	}
	if err := validateUniqueLinkTypes(c.LinkTypes); err != nil {
		return err
	}
	if err := c.ValidateKind(c.DefaultKind); err != nil {
		return fmt.Errorf("default_kind: %w", err)
	}
	if err := c.ValidateState(c.DefaultState); err != nil {
		return fmt.Errorf("default_state: %w", err)
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

func (c Config) ValidateState(state State) error {
	if strings.TrimSpace(string(state)) == "" {
		return fmt.Errorf("state is required")
	}
	if slices.Contains(c.States, state) {
		return nil
	}
	return fmt.Errorf("unknown state: %s", state)
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

func validateUniqueStates(values []State) error {
	seen := map[State]struct{}{}
	for _, value := range values {
		if value == "" {
			return fmt.Errorf("states must not include empty value")
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("states contains duplicate value: %s", value)
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
		if _, ok := seen[value]; ok {
			return fmt.Errorf("link_types contains duplicate value: %s", value)
		}
		seen[value] = struct{}{}
	}
	return nil
}
