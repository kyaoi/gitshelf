package cli

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

type postExitGitAction string

const (
	postExitGitNone       postExitGitAction = "none"
	postExitGitCommit     postExitGitAction = "commit"
	postExitGitCommitPush postExitGitAction = "commit_push"
)

type postExitGitSettings struct {
	Action        postExitGitAction
	CommitMessage string
}

func resolvePostExitGitSettings(ctx *commandContext, cfg shelf.Config) (postExitGitSettings, error) {
	action := strings.TrimSpace(ctx.gitOnExit)
	if action == "" {
		action = cfg.Commands.Cockpit.PostExitGitAction
	}
	switch postExitGitAction(action) {
	case postExitGitNone, postExitGitCommit, postExitGitCommitPush:
	default:
		return postExitGitSettings{}, fmt.Errorf("--git-on-exit must be one of none/commit/commit_push")
	}
	message := strings.TrimSpace(ctx.gitMessage)
	if message == "" {
		message = cfg.Commands.Cockpit.CommitMessage
	}
	if message == "" {
		return postExitGitSettings{}, fmt.Errorf("git commit message is empty")
	}
	return postExitGitSettings{
		Action:        postExitGitAction(action),
		CommitMessage: message,
	}, nil
}

func runPostExitGitAction(rootDir string, settings postExitGitSettings) error {
	if settings.Action == postExitGitNone {
		return nil
	}
	changed, err := gitPathHasChanges(rootDir, ".shelf")
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	if _, err := runGitCommand(rootDir, "add", ".shelf"); err != nil {
		return err
	}
	if _, err := runGitCommand(rootDir, "commit", "--only", "-m", settings.CommitMessage, "--", ".shelf"); err != nil {
		return err
	}
	if settings.Action == postExitGitCommitPush {
		if _, err := runGitCommand(rootDir, "push"); err != nil {
			return err
		}
	}
	return nil
}

func gitPathHasChanges(rootDir, path string) (bool, error) {
	output, err := runGitCommand(rootDir, "status", "--porcelain", "--", path)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) != "", nil
}

func runGitCommand(rootDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = rootDir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), message)
	}
	return stdout.String(), nil
}
