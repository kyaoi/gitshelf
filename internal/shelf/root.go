package shelf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyaoi/gitshelf/internal/paths"
)

const ShelfDirName = ".shelf"

var ErrShelfNotFound = errors.New(".shelf directory not found")

func ResolveShelfRoot(rootOverride, cwd string) (string, error) {
	if cwd == "" {
		return "", errors.New("cwd is required")
	}

	if rootOverride != "" {
		rootAbs, err := filepath.Abs(rootOverride)
		if err != nil {
			return "", fmt.Errorf("failed to resolve --root path: %w", err)
		}
		if err := ensureShelfConfig(rootAbs); err != nil {
			return "", err
		}
		return rootAbs, nil
	}

	dir, err := filepath.Abs(cwd)
	if err != nil {
		return "", fmt.Errorf("failed to resolve cwd: %w", err)
	}

	for {
		if err := ensureShelfConfig(dir); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	globalCfg, err := paths.LoadGlobalConfig()
	if err != nil {
		if errors.Is(err, paths.ErrGlobalConfigNotFound) {
			return "", ErrShelfNotFound
		}
		return "", err
	}
	if err := ensureShelfConfig(globalCfg.DefaultRoot); err != nil {
		return "", err
	}
	return globalCfg.DefaultRoot, nil
}

func ensureShelfConfig(root string) error {
	p := filepath.Join(root, ShelfDirName)
	info, err := os.Stat(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w under %s", ErrShelfNotFound, root)
		}
		return fmt.Errorf("failed to access %s: %w", p, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", p)
	}
	configPath := filepath.Join(p, "config.toml")
	info, err = os.Stat(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: missing %s", ErrShelfNotFound, configPath)
		}
		return fmt.Errorf("failed to access %s: %w", configPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s exists but is a directory", configPath)
	}
	return nil
}
