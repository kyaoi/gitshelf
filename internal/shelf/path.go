package shelf

import "path/filepath"

func ShelfDir(rootDir string) string {
	return filepath.Join(rootDir, ShelfDirName)
}

func StorageRootDir(rootDir string) string {
	cfg, err := LoadConfig(rootDir)
	if err == nil {
		if resolved, resolveErr := ResolveStorageRootDir(rootDir, cfg.StorageRoot); resolveErr == nil {
			return resolved
		}
	}
	resolved, _ := ResolveStorageRootDir(rootDir, DefaultConfig().StorageRoot)
	return resolved
}

func TasksDir(rootDir string) string {
	return filepath.Join(StorageRootDir(rootDir), "tasks")
}

func EdgesDir(rootDir string) string {
	return filepath.Join(StorageRootDir(rootDir), "edges")
}

func ConfigPath(rootDir string) string {
	return filepath.Join(ShelfDir(rootDir), "config.toml")
}
