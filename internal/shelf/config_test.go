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
}

func TestConfigRoundTrip(t *testing.T) {
	cfg := DefaultConfig()
	data := FormatConfigTOML(cfg)

	parsed, err := ParseConfigTOML(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if len(parsed.Kinds) != len(cfg.Kinds) || len(parsed.Statuses) != len(cfg.Statuses) || len(parsed.LinkTypes) != len(cfg.LinkTypes) {
		t.Fatalf("parsed config mismatch: %+v", parsed)
	}
	if parsed.DefaultKind != cfg.DefaultKind || parsed.DefaultStatus != cfg.DefaultStatus {
		t.Fatalf("parsed defaults mismatch: %+v", parsed)
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
	if !strings.Contains(data, "statuses =") || !strings.Contains(data, "default_status =") {
		t.Fatalf("formatted config should use status keys: %s", data)
	}
}
