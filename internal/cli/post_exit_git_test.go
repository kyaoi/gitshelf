package cli

import (
	"strings"
	"testing"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestRunPostExitGitActionCommit(t *testing.T) {
	root := t.TempDir()
	if _, err := runGitCommand(root, "init"); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	if _, err := runGitCommand(root, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config user.name failed: %v", err)
	}
	if _, err := runGitCommand(root, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config user.email failed: %v", err)
	}
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Task", Kind: "todo", Status: "open"}); err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	if err := runPostExitGitAction(root, postExitGitSettings{
		Action:        postExitGitCommit,
		CommitMessage: "test: commit shelf data",
	}); err != nil {
		t.Fatalf("runPostExitGitAction failed: %v", err)
	}

	message, err := runGitCommand(root, "log", "-1", "--pretty=%s")
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}
	if strings.TrimSpace(message) != "test: commit shelf data" {
		t.Fatalf("unexpected commit message: %q", message)
	}
}

func TestRunPostExitGitActionSkipsWhenNoShelfChanges(t *testing.T) {
	root := t.TempDir()
	if _, err := runGitCommand(root, "init"); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	if _, err := runGitCommand(root, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config user.name failed: %v", err)
	}
	if _, err := runGitCommand(root, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config user.email failed: %v", err)
	}
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Task", Kind: "todo", Status: "open"}); err != nil {
		t.Fatalf("add task failed: %v", err)
	}
	settings := postExitGitSettings{
		Action:        postExitGitCommit,
		CommitMessage: "test: commit shelf data",
	}
	if err := runPostExitGitAction(root, settings); err != nil {
		t.Fatalf("first runPostExitGitAction failed: %v", err)
	}
	if err := runPostExitGitAction(root, settings); err != nil {
		t.Fatalf("second runPostExitGitAction failed: %v", err)
	}
	count, err := runGitCommand(root, "rev-list", "--count", "HEAD")
	if err != nil {
		t.Fatalf("git rev-list failed: %v", err)
	}
	if strings.TrimSpace(count) != "1" {
		t.Fatalf("expected single commit, got %q", count)
	}
}

func TestRunPostExitGitActionCommitsConfiguredStoragePaths(t *testing.T) {
	root := t.TempDir()
	if _, err := runGitCommand(root, "init"); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	if _, err := runGitCommand(root, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config user.name failed: %v", err)
	}
	if _, err := runGitCommand(root, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config user.email failed: %v", err)
	}
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	cfg, err := shelf.LoadConfig(root)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	cfg.StorageRoot = "."
	if err := shelf.SaveConfig(root, cfg); err != nil {
		t.Fatalf("save config failed: %v", err)
	}
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("reinitialize failed: %v", err)
	}
	if _, err := shelf.AddTask(root, shelf.AddTaskInput{Title: "Task", Kind: "todo", Status: "open"}); err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	if err := runPostExitGitAction(root, postExitGitSettings{
		Action:        postExitGitCommit,
		CommitMessage: "test: commit configured storage",
	}); err != nil {
		t.Fatalf("runPostExitGitAction failed: %v", err)
	}

	files, err := runGitCommand(root, "show", "--pretty=", "--name-only", "HEAD")
	if err != nil {
		t.Fatalf("git show failed: %v", err)
	}
	for _, want := range []string{".shelf/config.toml", "tasks"} {
		if !strings.Contains(files, want) {
			t.Fatalf("expected committed paths to include %q, got %q", want, files)
		}
	}
}
