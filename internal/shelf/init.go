package shelf

import (
	"fmt"
	"os"
	"path/filepath"
)

type InitResult struct {
	RootDir       string
	ShelfDir      string
	ConfigCreated bool
	ConfigForced  bool
}

func Initialize(rootDir string, force bool) (InitResult, error) {
	shelfDir := ShelfDir(rootDir)
	tasksDir := TasksDir(rootDir)
	edgesDir := EdgesDir(rootDir)

	for _, dir := range []string{shelfDir, tasksDir, edgesDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return InitResult{}, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	for _, legacyDir := range []string{
		filepath.Join(shelfDir, "templates"),
		filepath.Join(shelfDir, "history"),
	} {
		if err := os.RemoveAll(legacyDir); err != nil {
			return InitResult{}, fmt.Errorf("failed to remove legacy directory %s: %w", legacyDir, err)
		}
	}

	result := InitResult{
		RootDir:  rootDir,
		ShelfDir: shelfDir,
	}

	cfgPath := ConfigPath(rootDir)
	_, err := os.Stat(cfgPath)
	switch {
	case err == nil && !force:
		return result, nil
	case err == nil && force:
		result.ConfigForced = true
	case os.IsNotExist(err):
		result.ConfigCreated = true
	case err != nil:
		return InitResult{}, fmt.Errorf("failed to access config %s: %w", cfgPath, err)
	}

	if err := SaveConfig(rootDir, DefaultConfig()); err != nil {
		return InitResult{}, err
	}
	return result, nil
}
