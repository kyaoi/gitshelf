package cli

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
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
	paths, err := managedGitPaths(rootDir)
	if err != nil {
		return err
	}
	changedPaths, err := gitChangedPaths(rootDir, paths)
	if err != nil {
		return err
	}
	if len(changedPaths) == 0 {
		return nil
	}
	addArgs := []string{"add", "-A", "--"}
	addArgs = append(addArgs, changedPaths...)
	if _, err := runGitCommand(rootDir, addArgs...); err != nil {
		return err
	}
	args := []string{"commit", "--only", "-m", settings.CommitMessage, "--"}
	args = append(args, changedPaths...)
	if _, err := runGitCommand(rootDir, args...); err != nil {
		return err
	}
	if settings.Action == postExitGitCommitPush {
		if _, err := runGitCommand(rootDir, "push"); err != nil {
			return err
		}
	}
	return nil
}

func gitChangedPaths(rootDir string, paths []string) ([]string, error) {
	args := []string{"status", "--porcelain=v1", "-z", "--untracked-files=all", "--"}
	args = append(args, paths...)
	output, err := runGitCommand(rootDir, args...)
	if err != nil {
		return nil, err
	}
	entries := strings.Split(output, "\x00")
	changed := make([]string, 0, len(entries))
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if entry == "" || len(entry) < 4 {
			continue
		}
		status := entry[:2]
		path := entry[3:]
		if path == "" {
			continue
		}
		if !slices.Contains(changed, path) {
			changed = append(changed, path)
		}
		if status[0] == 'R' || status[0] == 'C' {
			if i+1 >= len(entries) {
				continue
			}
			i++
			previousPath := entries[i]
			if previousPath == "" || slices.Contains(changed, previousPath) {
				continue
			}
			changed = append(changed, previousPath)
		}
	}
	return changed, nil
}

func managedGitPaths(rootDir string) ([]string, error) {
	paths := []string{filepath.ToSlash(filepath.Join(shelf.ShelfDirName, "config.toml"))}
	cfg, err := shelf.LoadConfig(rootDir)
	if err != nil {
		return nil, err
	}
	paths, err = appendManagedStoragePaths(rootDir, paths, cfg.StorageRoot)
	if err != nil {
		return nil, err
	}
	headConfig, err := gitHeadShelfConfig(rootDir)
	if err == nil {
		paths, err = appendManagedStoragePaths(rootDir, paths, headConfig.StorageRoot)
		if err != nil {
			return nil, err
		}
	}
	return paths, nil
}

func appendManagedStoragePaths(rootDir string, paths []string, storageRoot string) ([]string, error) {
	storageDir, err := shelf.ResolveStorageRootDir(rootDir, storageRoot)
	if err != nil {
		return nil, err
	}
	for _, abs := range []string{
		filepath.Join(storageDir, "tasks"),
		filepath.Join(storageDir, "edges"),
	} {
		rel, err := filepath.Rel(rootDir, abs)
		if err != nil {
			return nil, err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." || slices.Contains(paths, rel) {
			continue
		}
		paths = append(paths, rel)
	}
	return paths, nil
}

func gitHeadShelfConfig(rootDir string) (shelf.Config, error) {
	configPath := filepath.ToSlash(filepath.Join(shelf.ShelfDirName, "config.toml"))
	data, err := runGitCommand(rootDir, "show", "HEAD:"+configPath)
	if err != nil {
		return shelf.Config{}, err
	}
	return shelf.ParseConfigTOML([]byte(data))
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
