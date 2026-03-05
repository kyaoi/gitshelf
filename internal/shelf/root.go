package shelf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
		if err := ensureShelfDir(rootAbs); err != nil {
			return "", err
		}
		return rootAbs, nil
	}

	dir, err := filepath.Abs(cwd)
	if err != nil {
		return "", fmt.Errorf("failed to resolve cwd: %w", err)
	}

	for {
		if err := ensureShelfDir(dir); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", ErrShelfNotFound
}

func ensureShelfDir(root string) error {
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
	return nil
}
