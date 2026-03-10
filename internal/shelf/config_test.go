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
	if cfg.Commands.Cockpit.CopySeparator != "\n" {
		t.Fatalf("unexpected cockpit copy separator: %q", cfg.Commands.Cockpit.CopySeparator)
	}
	if cfg.Commands.Cockpit.PostExitGitAction != "none" || cfg.Commands.Cockpit.CommitMessage == "" {
		t.Fatalf("unexpected cockpit git defaults: %+v", cfg.Commands.Cockpit)
	}
	if cfg.StorageRoot != ".shelf" {
		t.Fatalf("unexpected storage root: %q", cfg.StorageRoot)
	}
	if len(cfg.LinkTypes.Names) != 2 || cfg.LinkTypes.Names[0] != "depends_on" || cfg.LinkTypes.Blocking != "depends_on" {
		t.Fatalf("unexpected default link types: %+v", cfg.LinkTypes)
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

	if len(parsed.Kinds) != len(cfg.Kinds) || len(parsed.Statuses) != len(cfg.Statuses) || len(parsed.LinkTypes.Names) != len(cfg.LinkTypes.Names) {
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
	if parsed.Commands.Cockpit.CopySeparator != cfg.Commands.Cockpit.CopySeparator {
		t.Fatalf("parsed cockpit defaults mismatch: %+v", parsed)
	}
	if parsed.Commands.Cockpit.PostExitGitAction != cfg.Commands.Cockpit.PostExitGitAction || parsed.Commands.Cockpit.CommitMessage != cfg.Commands.Cockpit.CommitMessage {
		t.Fatalf("parsed cockpit git settings mismatch: %+v", parsed.Commands.Cockpit)
	}
	if parsed.StorageRoot != cfg.StorageRoot {
		t.Fatalf("parsed storage root mismatch: %q", parsed.StorageRoot)
	}
	if parsed.LinkTypes.Blocking != cfg.LinkTypes.Blocking {
		t.Fatalf("parsed blocking link type mismatch: %+v", parsed.LinkTypes)
	}
}

func TestParseConfigSupportsLegacyLinkTypesArray(t *testing.T) {
	data := []byte(`kinds = ["todo"]
statuses = ["open", "done"]
tags = []
link_types = ["develop_first", "related"]
default_kind = "todo"
default_status = "open"
`)
	cfg, err := ParseConfigTOML(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(cfg.LinkTypes.Names) != 2 || cfg.LinkTypes.Names[0] != "develop_first" {
		t.Fatalf("unexpected parsed link types: %+v", cfg.LinkTypes)
	}
	if cfg.LinkTypes.Blocking != "develop_first" {
		t.Fatalf("expected first legacy link type to become blocking, got %+v", cfg.LinkTypes)
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

func TestConfigValidationRejectsBlockingLinkTypeOutsideNames(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LinkTypes.Blocking = "derived_from"

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestConfigValidationRejectsEmptyStorageRoot(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StorageRoot = "  "
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "storage_root") {
		t.Fatalf("expected storage_root validation error, got: %v", err)
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

func TestConfigValidationRejectsInvalidCockpitGitAction(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Commands.Cockpit.PostExitGitAction = "push_only"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "commands.cockpit.post_exit_git_action") {
		t.Fatalf("expected invalid cockpit git action error, got: %v", err)
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

func TestResolveStorageRootDirRejectsPathOutsideRoot(t *testing.T) {
	root := t.TempDir()
	if _, err := ResolveStorageRootDir(root, "../outside"); err == nil {
		t.Fatal("expected error")
	}
}
