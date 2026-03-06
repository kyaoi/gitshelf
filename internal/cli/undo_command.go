package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

type undoMeta struct {
	Action    string `json:"action"`
	CreatedAt string `json:"created_at"`
}

func newUndoCommand(ctx *commandContext) *cobra.Command {
	return &cobra.Command{
		Use:     "undo",
		Short:   "Undo last mutating action",
		Example: "  shelf undo",
		RunE: func(_ *cobra.Command, _ []string) error {
			meta, err := restoreUndoSnapshot(ctx.rootDir)
			if err != nil {
				return err
			}
			fmt.Printf("Undone: %s\n", meta.Action)
			return nil
		},
	}
}

func prepareUndoSnapshot(rootDir string, action string) error {
	history := filepath.Join(rootDir, ".shelf", "history")
	snapshot := filepath.Join(history, "snapshot")
	metaPath := filepath.Join(history, "last_action.json")

	if err := os.MkdirAll(history, 0o755); err != nil {
		return err
	}
	if err := os.RemoveAll(snapshot); err != nil {
		return err
	}
	if err := os.MkdirAll(snapshot, 0o755); err != nil {
		return err
	}

	shelfDir := filepath.Join(rootDir, ".shelf")
	for _, rel := range []string{"config.toml", "tasks", "edges"} {
		src := filepath.Join(shelfDir, rel)
		dst := filepath.Join(snapshot, rel)
		info, err := os.Stat(src)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if info.IsDir() {
			if err := copyDir(src, dst); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(src, dst, info.Mode()); err != nil {
			return err
		}
	}

	meta := undoMeta{
		Action:    action,
		CreatedAt: time.Now().Local().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath, data, 0o644)
}

func restoreUndoSnapshot(rootDir string) (undoMeta, error) {
	history := filepath.Join(rootDir, ".shelf", "history")
	snapshot := filepath.Join(history, "snapshot")
	metaPath := filepath.Join(history, "last_action.json")
	var meta undoMeta

	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return meta, fmt.Errorf("undo history is empty")
		}
		return meta, err
	}
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return meta, err
	}
	if _, err := os.Stat(snapshot); err != nil {
		if os.IsNotExist(err) {
			return meta, fmt.Errorf("undo snapshot is missing")
		}
		return meta, err
	}

	shelfDir := filepath.Join(rootDir, ".shelf")
	for _, rel := range []string{"tasks", "edges"} {
		if err := os.RemoveAll(filepath.Join(shelfDir, rel)); err != nil {
			return meta, err
		}
	}

	for _, rel := range []string{"config.toml", "tasks", "edges"} {
		src := filepath.Join(snapshot, rel)
		dst := filepath.Join(shelfDir, rel)
		info, err := os.Stat(src)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return meta, err
		}
		if info.IsDir() {
			if err := copyDir(src, dst); err != nil {
				return meta, err
			}
			continue
		}
		if err := copyFile(src, dst, info.Mode()); err != nil {
			return meta, err
		}
	}

	if err := os.RemoveAll(snapshot); err != nil {
		return meta, err
	}
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		return meta, err
	}
	return meta, nil
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(srcPath, dstPath, info.Mode()); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}
