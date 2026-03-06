package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

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
		Example: "  shelf history\n  shelf history --limit 20\n  shelf history --json",
		RunE: func(_ *cobra.Command, _ []string) error {
			if limit <= 0 {
				return fmt.Errorf("--limit must be >= 1")
			}
			entries, err := loadActionHistory(ctx.rootDir)
			if err != nil {
				return err
			}
			if len(entries) > limit {
				entries = entries[len(entries)-limit:]
			}
			for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
				entries[i], entries[j] = entries[j], entries[i]
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
			for _, entry := range entries {
				fmt.Printf("%s [%s] %s\n", entry.CreatedAt, entry.Event, entry.Action)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of history entries")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
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
