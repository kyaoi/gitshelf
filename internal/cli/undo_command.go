package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

const maxHistoryEntries = 50

type snapshotMeta struct {
	ID        string `json:"id"`
	Action    string `json:"action"`
	CreatedAt string `json:"created_at"`
}

type actionLogEntry struct {
	Action     string `json:"action"`
	Event      string `json:"event"`
	SnapshotID string `json:"snapshot_id,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type historyIndex struct {
	Undo []snapshotMeta `json:"undo"`
	Redo []snapshotMeta `json:"redo"`
}

func newUndoCommand(ctx *commandContext) *cobra.Command {
	var steps int
	cmd := &cobra.Command{
		Use:     "undo",
		Short:   "Undo mutating actions",
		Example: "  shelf undo\n  shelf undo --steps 3",
		RunE: func(_ *cobra.Command, _ []string) error {
			if steps <= 0 {
				return fmt.Errorf("--steps must be >= 1")
			}
			return withWriteLock(ctx.rootDir, func() error {
				last, err := restoreUndoSnapshots(ctx.rootDir, steps)
				if err != nil {
					return err
				}
				fmt.Printf("Undone (%d): %s\n", steps, last.Action)
				return nil
			})
		},
	}
	cmd.Flags().IntVar(&steps, "steps", 1, "Number of actions to undo")
	return cmd
}

func newRedoCommand(ctx *commandContext) *cobra.Command {
	var steps int
	cmd := &cobra.Command{
		Use:     "redo",
		Short:   "Redo undone mutating actions",
		Example: "  shelf redo\n  shelf redo --steps 2",
		RunE: func(_ *cobra.Command, _ []string) error {
			if steps <= 0 {
				return fmt.Errorf("--steps must be >= 1")
			}
			return withWriteLock(ctx.rootDir, func() error {
				last, err := restoreRedoSnapshots(ctx.rootDir, steps)
				if err != nil {
					return err
				}
				fmt.Printf("Redone (%d): %s\n", steps, last.Action)
				return nil
			})
		},
	}
	cmd.Flags().IntVar(&steps, "steps", 1, "Number of actions to redo")
	return cmd
}

func prepareUndoSnapshot(rootDir string, action string) error {
	return nil
}

func restoreUndoSnapshots(rootDir string, steps int) (snapshotMeta, error) {
	idx, err := loadHistoryIndex(rootDir)
	if err != nil {
		return snapshotMeta{}, err
	}
	if len(idx.Undo) == 0 {
		return snapshotMeta{}, fmt.Errorf("undo history is empty")
	}
	if steps > len(idx.Undo) {
		return snapshotMeta{}, fmt.Errorf("undo history has only %d entries", len(idx.Undo))
	}

	var last snapshotMeta
	for i := 0; i < steps; i++ {
		target := idx.Undo[len(idx.Undo)-1]
		idx.Undo = idx.Undo[:len(idx.Undo)-1]

		current, err := captureSnapshot(rootDir, target.Action)
		if err != nil {
			return snapshotMeta{}, err
		}
		idx.Redo = append(idx.Redo, current)
		idx.Redo, err = truncateHistory(rootDir, idx.Redo)
		if err != nil {
			return snapshotMeta{}, err
		}

		if err := restoreSnapshot(rootDir, target.ID); err != nil {
			return snapshotMeta{}, err
		}
		if err := removeSnapshotDir(rootDir, target.ID); err != nil {
			return snapshotMeta{}, err
		}
		if err := appendActionLog(rootDir, actionLogEntry{
			Action:     target.Action,
			Event:      "undo",
			SnapshotID: target.ID,
			CreatedAt:  time.Now().Local().Round(time.Second).Format(time.RFC3339),
		}); err != nil {
			return snapshotMeta{}, err
		}
		last = target
	}

	if err := saveHistoryIndex(rootDir, idx); err != nil {
		return snapshotMeta{}, err
	}
	return last, nil
}

func restoreRedoSnapshots(rootDir string, steps int) (snapshotMeta, error) {
	idx, err := loadHistoryIndex(rootDir)
	if err != nil {
		return snapshotMeta{}, err
	}
	if len(idx.Redo) == 0 {
		return snapshotMeta{}, fmt.Errorf("redo history is empty")
	}
	if steps > len(idx.Redo) {
		return snapshotMeta{}, fmt.Errorf("redo history has only %d entries", len(idx.Redo))
	}

	var last snapshotMeta
	for i := 0; i < steps; i++ {
		target := idx.Redo[len(idx.Redo)-1]
		idx.Redo = idx.Redo[:len(idx.Redo)-1]

		current, err := captureSnapshot(rootDir, target.Action)
		if err != nil {
			return snapshotMeta{}, err
		}
		idx.Undo = append(idx.Undo, current)
		idx.Undo, err = truncateHistory(rootDir, idx.Undo)
		if err != nil {
			return snapshotMeta{}, err
		}

		if err := restoreSnapshot(rootDir, target.ID); err != nil {
			return snapshotMeta{}, err
		}
		if err := removeSnapshotDir(rootDir, target.ID); err != nil {
			return snapshotMeta{}, err
		}
		if err := appendActionLog(rootDir, actionLogEntry{
			Action:     target.Action,
			Event:      "redo",
			SnapshotID: target.ID,
			CreatedAt:  time.Now().Local().Round(time.Second).Format(time.RFC3339),
		}); err != nil {
			return snapshotMeta{}, err
		}
		last = target
	}

	if err := saveHistoryIndex(rootDir, idx); err != nil {
		return snapshotMeta{}, err
	}
	return last, nil
}

func historyDir(rootDir string) string {
	return filepath.Join(rootDir, ".shelf", "history")
}

func snapshotsDir(rootDir string) string {
	return filepath.Join(historyDir(rootDir), "snapshots")
}

func historyIndexPath(rootDir string) string {
	return filepath.Join(historyDir(rootDir), "index.json")
}

func historyLogPath(rootDir string) string {
	return filepath.Join(historyDir(rootDir), "actions.log")
}

func loadHistoryIndex(rootDir string) (historyIndex, error) {
	history := historyDir(rootDir)
	if err := os.MkdirAll(history, 0o755); err != nil {
		return historyIndex{}, err
	}
	path := historyIndexPath(rootDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return historyIndex{}, nil
		}
		return historyIndex{}, err
	}
	var idx historyIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return historyIndex{}, err
	}
	return idx, nil
}

func saveHistoryIndex(rootDir string, idx historyIndex) error {
	if err := os.MkdirAll(historyDir(rootDir), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(historyIndexPath(rootDir), data, 0o644)
}

func appendActionLog(rootDir string, entry actionLogEntry) error {
	if err := os.MkdirAll(historyDir(rootDir), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(historyLogPath(rootDir), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

func captureSnapshot(rootDir string, action string) (snapshotMeta, error) {
	now := time.Now().Local()
	ts := now.Round(time.Second)
	id := strconv.FormatInt(now.UnixNano(), 10)
	snapshot := filepath.Join(snapshotsDir(rootDir), id)
	if err := os.MkdirAll(snapshot, 0o755); err != nil {
		return snapshotMeta{}, err
	}

	items := map[string]string{
		"config.toml": shelf.ConfigPath(rootDir),
		"tasks":       shelf.TasksDir(rootDir),
		"edges":       shelf.EdgesDir(rootDir),
	}
	for rel, src := range items {
		dst := filepath.Join(snapshot, rel)
		info, err := os.Stat(src)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return snapshotMeta{}, err
		}
		if info.IsDir() {
			if err := copyDir(src, dst); err != nil {
				return snapshotMeta{}, err
			}
			continue
		}
		if err := copyFile(src, dst, info.Mode()); err != nil {
			return snapshotMeta{}, err
		}
	}
	return snapshotMeta{
		ID:        id,
		Action:    action,
		CreatedAt: ts.Format(time.RFC3339),
	}, nil
}

func restoreSnapshot(rootDir string, snapshotID string) error {
	snapshot := filepath.Join(snapshotsDir(rootDir), snapshotID)
	if _, err := os.Stat(snapshot); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("history snapshot is missing: %s", snapshotID)
		}
		return err
	}

	for _, dir := range []string{shelf.TasksDir(rootDir), shelf.EdgesDir(rootDir)} {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}

	items := map[string]string{
		"config.toml": shelf.ConfigPath(rootDir),
		"tasks":       shelf.TasksDir(rootDir),
		"edges":       shelf.EdgesDir(rootDir),
	}
	for rel, dst := range items {
		src := filepath.Join(snapshot, rel)
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
	return nil
}

func removeSnapshotDir(rootDir string, snapshotID string) error {
	return os.RemoveAll(filepath.Join(snapshotsDir(rootDir), snapshotID))
}

func truncateHistory(rootDir string, items []snapshotMeta) ([]snapshotMeta, error) {
	if len(items) <= maxHistoryEntries {
		return items, nil
	}
	excess := len(items) - maxHistoryEntries
	for i := 0; i < excess; i++ {
		if err := removeSnapshotDir(rootDir, items[i].ID); err != nil {
			return nil, err
		}
	}
	return items[excess:], nil
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
