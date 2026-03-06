package cli

import (
	"strings"
	"testing"
)

func TestShowWithoutIDFailsOnNonTTY(t *testing.T) {
	root := t.TempDir()

	cmd := NewRootCommand("test")
	cmd.SetArgs([]string{"init", "--root", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cmd = NewRootCommand("test")
	cmd.SetArgs([]string{"show", "--root", root})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "非TTYでは対話入力できません。<id> を指定してください") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEditWithoutIDFailsOnNonTTY(t *testing.T) {
	root := t.TempDir()

	cmd := NewRootCommand("test")
	cmd.SetArgs([]string{"init", "--root", root})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	cmd = NewRootCommand("test")
	cmd.SetArgs([]string{"edit", "--root", root})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "<id> を指定してください") {
		t.Fatalf("unexpected error: %v", err)
	}
}
