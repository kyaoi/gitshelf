package shelf

import (
	"testing"
)

func TestDefaultConfigIsValid(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should be valid: %v", err)
	}
}

func TestConfigRoundTrip(t *testing.T) {
	cfg := DefaultConfig()
	data := FormatConfigTOML(cfg)

	parsed, err := ParseConfigTOML(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if len(parsed.Kinds) != len(cfg.Kinds) || len(parsed.States) != len(cfg.States) || len(parsed.LinkTypes) != len(cfg.LinkTypes) {
		t.Fatalf("parsed config mismatch: %+v", parsed)
	}
	if parsed.DefaultKind != cfg.DefaultKind || parsed.DefaultState != cfg.DefaultState {
		t.Fatalf("parsed defaults mismatch: %+v", parsed)
	}
}

func TestConfigValidationRejectsUnknownKindStateLinkType(t *testing.T) {
	cfg := DefaultConfig()

	if err := cfg.ValidateKind("unknown"); err == nil {
		t.Fatal("expected unknown kind error")
	}
	if err := cfg.ValidateState("unknown"); err == nil {
		t.Fatal("expected unknown state error")
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
