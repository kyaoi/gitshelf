package versioninfo

import (
	"runtime/debug"
	"testing"
)

func TestResolvePrefersExplicitVersion(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "v1.3"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abcdef1234567890"},
		},
	}
	if got := resolve("v9.9", info); got != "v9.9" {
		t.Fatalf("expected explicit version, got %q", got)
	}
}

func TestResolveUsesModuleVersionWhenAvailable(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "v1.3"},
	}
	if got := resolve("", info); got != "v1.3" {
		t.Fatalf("expected module version, got %q", got)
	}
}

func TestResolveUsesShortRevisionForDevelBuilds(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: develVersion},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abcdef1234567890"},
		},
	}
	if got := resolve("", info); got != "dev+abcdef1" {
		t.Fatalf("expected short revision, got %q", got)
	}
}

func TestResolveMarksDirtyDevelBuilds(t *testing.T) {
	info := &debug.BuildInfo{
		Main: debug.Module{Version: develVersion},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abcdef1234567890"},
			{Key: "vcs.modified", Value: "true"},
		},
	}
	if got := resolve("", info); got != "dev+abcdef1-dirty" {
		t.Fatalf("expected dirty revision, got %q", got)
	}
}

func TestResolveFallsBackToDev(t *testing.T) {
	if got := resolve("", nil); got != "dev" {
		t.Fatalf("expected dev fallback, got %q", got)
	}
}
