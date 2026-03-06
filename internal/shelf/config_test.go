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
}

func TestConfigRoundTrip(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Tags = []string{"backend", "urgent"}
	cfg.Views["active"] = TaskView{
		Tags:        []string{"backend"},
		NotStatuses: []Status{"done", "cancelled"},
	}
	cfg.OutputPresets["focus"] = OutputPreset{
		Command: "ls",
		Format:  "detail",
		View:    "active",
		Limit:   10,
	}
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
	if _, ok := parsed.Views["active"]; !ok {
		t.Fatalf("parsed views mismatch: %+v", parsed.Views)
	}
	if len(parsed.Views["active"].Tags) != 1 || parsed.Views["active"].Tags[0] != "backend" {
		t.Fatalf("parsed view tags mismatch: %+v", parsed.Views["active"].Tags)
	}
	if _, ok := parsed.OutputPresets["focus"]; !ok {
		t.Fatalf("parsed output presets mismatch: %+v", parsed.OutputPresets)
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

func TestParseConfigTOMLViewsValidation(t *testing.T) {
	raw := `
kinds = ["todo", "memo"]
statuses = ["open", "done"]
link_types = ["depends_on", "related"]
default_kind = "todo"
default_status = "open"

[views."active"]
not_statuses = ["done"]
limit = 20
`
	cfg, err := ParseConfigTOML([]byte(raw))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	view, ok := cfg.Views["active"]
	if !ok {
		t.Fatalf("expected active view: %+v", cfg.Views)
	}
	if len(view.NotStatuses) != 1 || view.NotStatuses[0] != "done" {
		t.Fatalf("unexpected not statuses: %+v", view.NotStatuses)
	}
	if view.Limit != 20 {
		t.Fatalf("unexpected view limit: %d", view.Limit)
	}
}

func TestConfigValidationRejectsUnknownStatusInView(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Views["bad"] = TaskView{
		Statuses: []Status{"unknown"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for unknown view status")
	}
}

func TestConfigValidationRejectsUnknownTagInView(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Tags = []string{"backend"}
	cfg.Views["bad"] = TaskView{
		Tags: []string{"unknown"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for unknown view tag")
	}
}

func TestConfigValidationOutputPreset(t *testing.T) {
	cfg := DefaultConfig()
	cfg.OutputPresets["bad"] = OutputPreset{
		Command: "next",
		Format:  "detail",
	}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "next does not support format") {
		t.Fatalf("expected invalid output preset format error, got: %v", err)
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
