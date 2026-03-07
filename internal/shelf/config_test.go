package shelf

import (
	"strings"
	"testing"
)

func TestDefaultConfigIsValid(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should be valid: %v", err)
	}
	if err := cfg.ValidateKind("inbox"); err != nil {
		t.Fatalf("default config should include inbox kind: %v", err)
	}
	expected := []Status{"open", "in_progress", "blocked", "done", "cancelled"}
	if len(cfg.Statuses) != len(expected) {
		t.Fatalf("unexpected default statuses: %+v", cfg.Statuses)
	}
	for i, status := range expected {
		if cfg.Statuses[i] != status {
			t.Fatalf("unexpected default statuses: %+v", cfg.Statuses)
		}
	}
	if cfg.DefaultStatus != "open" {
		t.Fatalf("unexpected default status: %s", cfg.DefaultStatus)
	}
	if cfg.Commands.Calendar.DefaultRangeUnit != "days" {
		t.Fatalf("unexpected calendar default use: %s", cfg.Commands.Calendar.DefaultRangeUnit)
	}
	if cfg.Commands.Calendar.DefaultDays != 7 || cfg.Commands.Calendar.DefaultMonths != 6 || cfg.Commands.Calendar.DefaultYears != 2 {
		t.Fatalf("unexpected calendar defaults: %+v", cfg)
	}
}

func TestConfigRoundTrip(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Tags = []string{"backend", "urgent"}
	data := FormatConfigTOML(cfg)

	parsed, err := ParseConfigTOML(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if len(parsed.Kinds) != len(cfg.Kinds) || len(parsed.Statuses) != len(cfg.Statuses) || len(parsed.LinkTypes) != len(cfg.LinkTypes) {
		t.Fatalf("parsed config mismatch: %+v", parsed)
	}
	if len(parsed.Tags) != 2 || parsed.Tags[0] != "backend" || parsed.Tags[1] != "urgent" {
		t.Fatalf("parsed tags mismatch: %+v", parsed.Tags)
	}
	if parsed.DefaultKind != cfg.DefaultKind || parsed.DefaultStatus != cfg.DefaultStatus {
		t.Fatalf("parsed defaults mismatch: %+v", parsed)
	}
	if parsed.Commands.Calendar.DefaultRangeUnit != cfg.Commands.Calendar.DefaultRangeUnit ||
		parsed.Commands.Calendar.DefaultDays != cfg.Commands.Calendar.DefaultDays ||
		parsed.Commands.Calendar.DefaultMonths != cfg.Commands.Calendar.DefaultMonths ||
		parsed.Commands.Calendar.DefaultYears != cfg.Commands.Calendar.DefaultYears {
		t.Fatalf("parsed calendar defaults mismatch: %+v", parsed)
	}
}

func TestConfigValidationRejectsUnknownKindStatusLinkType(t *testing.T) {
	cfg := DefaultConfig()

	if err := cfg.ValidateKind("unknown"); err == nil {
		t.Fatal("expected unknown kind error")
	}
	if err := cfg.ValidateStatus("unknown"); err == nil {
		t.Fatal("expected unknown status error")
	}
	if err := cfg.ValidateLinkType("unknown"); err == nil {
		t.Fatal("expected unknown link type error")
	}
}

func TestConfigValidationRejectsUnsupportedLinkTypeInConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LinkTypes = append(cfg.LinkTypes, "derived_from")

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestParseConfigTOMLLegacyStateKeys(t *testing.T) {
	raw := `
kinds = ["todo"]
states = ["open", "done"]
link_types = ["depends_on", "related"]
default_kind = "todo"
default_state = "open"
`
	cfg, err := ParseConfigTOML([]byte(raw))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(cfg.Statuses) != 2 {
		t.Fatalf("unexpected statuses: %+v", cfg.Statuses)
	}
	if cfg.DefaultStatus != "open" {
		t.Fatalf("unexpected default status: %s", cfg.DefaultStatus)
	}
	data := string(FormatConfigTOML(cfg))
	if !strings.Contains(data, "statuses =") || !strings.Contains(data, "default_status =") || !strings.Contains(data, "[commands.calendar]") {
		t.Fatalf("formatted config should use status keys: %s", data)
	}
}

func TestConfigValidationRejectsInvalidCalendarDefaultDays(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Commands.Calendar.DefaultDays = 0
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "commands.calendar.default_days") {
		t.Fatalf("expected invalid calendar default days error, got: %v", err)
	}
}

func TestConfigValidationRejectsInvalidCalendarDefaultUse(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Commands.Calendar.DefaultRangeUnit = "weeks"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "commands.calendar.default_range_unit") {
		t.Fatalf("expected invalid calendar default use error, got: %v", err)
	}
}

func TestConfigAppendMissingTags(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Tags = []string{"backend"}
	changed := cfg.AppendMissingTags([]string{"backend", "urgent", " urgent "})
	if !changed {
		t.Fatal("expected config tags to change")
	}
	if len(cfg.Tags) != 2 || cfg.Tags[1] != "urgent" {
		t.Fatalf("unexpected tags: %+v", cfg.Tags)
	}
	if cfg.AppendMissingTags([]string{"backend"}) {
		t.Fatal("expected no change for existing tags")
	}
}
