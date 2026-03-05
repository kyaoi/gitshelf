package shelf

import "path/filepath"

func ShelfDir(rootDir string) string {
	return filepath.Join(rootDir, ShelfDirName)
}

func TasksDir(rootDir string) string {
	return filepath.Join(ShelfDir(rootDir), "tasks")
}

func EdgesDir(rootDir string) string {
	return filepath.Join(ShelfDir(rootDir), "edges")
}

func ConfigPath(rootDir string) string {
	return filepath.Join(ShelfDir(rootDir), "config.toml")
}
