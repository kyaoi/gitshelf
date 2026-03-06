package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func newHistoryCommand(ctx *commandContext) *cobra.Command {
	var (
		limit  int
		asJSON bool
	)
	cmd := &cobra.Command{
		Use:     "history",
		Short:   "Show mutating action history",
		Example: "  shelf history\n  shelf history --limit 20\n  shelf history --json\n  shelf history show 1",
		RunE: func(_ *cobra.Command, _ []string) error {
			entries, err := listHistoryEntries(ctx.rootDir, limit)
			if err != nil {
				return err
			}

			if asJSON {
				data, err := json.MarshalIndent(entries, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}
			if len(entries) == 0 {
				fmt.Println("(none)")
				return nil
			}
			for i, entry := range entries {
				if strings.TrimSpace(entry.SnapshotID) != "" {
					fmt.Printf("%d. %s [%s] %s snapshot=%s\n", i+1, entry.CreatedAt, entry.Event, entry.Action, entry.SnapshotID)
					continue
				}
				fmt.Printf("%d. %s [%s] %s\n", i+1, entry.CreatedAt, entry.Event, entry.Action)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of history entries")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.AddCommand(newHistoryShowCommand(ctx))
	return cmd
}

func newHistoryShowCommand(ctx *commandContext) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "show <entry>",
		Short:   "Show details of a history entry by index or snapshot ID",
		Example: "  shelf history show 1\n  shelf history show 1730982...\n  shelf history show 3 --json",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			entries, err := listHistoryEntries(ctx.rootDir, 0)
			if err != nil {
				return err
			}
			entry, index, err := resolveHistoryEntry(entries, args[0])
			if err != nil {
				return err
			}

			snapshotPath := ""
			snapshotExists := false
			if strings.TrimSpace(entry.SnapshotID) != "" {
				snapshotPath = filepath.Join(snapshotsDir(ctx.rootDir), entry.SnapshotID)
				if _, statErr := os.Stat(snapshotPath); statErr == nil {
					snapshotExists = true
				}
			}

			if asJSON {
				payload := map[string]any{
					"entry":          entry,
					"index":          index,
					"snapshot_path":  snapshotPath,
					"snapshot_exists": snapshotExists,
				}
				data, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("index: %d\n", index)
			fmt.Printf("created_at: %s\n", entry.CreatedAt)
			fmt.Printf("event: %s\n", entry.Event)
			fmt.Printf("action: %s\n", entry.Action)
			if strings.TrimSpace(entry.SnapshotID) == "" {
				fmt.Println("snapshot_id: (none)")
				return nil
			}
			fmt.Printf("snapshot_id: %s\n", entry.SnapshotID)
			fmt.Printf("snapshot_path: %s\n", snapshotPath)
			fmt.Printf("snapshot_exists: %t\n", snapshotExists)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func listHistoryEntries(rootDir string, limit int) ([]actionLogEntry, error) {
	if limit < 0 {
		return nil, fmt.Errorf("--limit must be >= 0")
	}
	entries, err := loadActionHistory(rootDir)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries, nil
}

func resolveHistoryEntry(entries []actionLogEntry, ref string) (actionLogEntry, int, error) {
	if len(entries) == 0 {
		return actionLogEntry{}, 0, fmt.Errorf("history is empty")
	}
	if n, err := strconv.Atoi(strings.TrimSpace(ref)); err == nil {
		if n < 1 || n > len(entries) {
			return actionLogEntry{}, 0, fmt.Errorf("history index out of range: %d", n)
		}
		return entries[n-1], n, nil
	}
	for i, entry := range entries {
		if entry.SnapshotID == ref {
			return entry, i + 1, nil
		}
	}
	return actionLogEntry{}, 0, fmt.Errorf("history entry not found: %s", ref)
}

func loadActionHistory(rootDir string) ([]actionLogEntry, error) {
	path := historyLogPath(rootDir)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []actionLogEntry{}, nil
		}
		return nil, err
	}
	defer f.Close()

	result := make([]actionLogEntry, 0, 64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry actionLogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		result = append(result, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
