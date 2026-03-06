package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

type shelfExport struct {
	Version    int                     `json:"version"`
	ExportedAt string                  `json:"exported_at"`
	Config     shelf.Config            `json:"config"`
	Tasks      []shelf.Task            `json:"tasks"`
	Edges      map[string][]shelf.Edge `json:"edges"`
}

func newExportCommand(ctx *commandContext) *cobra.Command {
	var outPath string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export config/tasks/edges as JSON",
		Example: "  shelf export\n" +
			"  shelf export --out backup.json",
		RunE: func(_ *cobra.Command, _ []string) error {
			payload, err := buildShelfExport(ctx.rootDir)
			if err != nil {
				return err
			}
			data, err := json.MarshalIndent(payload, "", "  ")
			if err != nil {
				return err
			}
			data = append(data, '\n')

			if strings.TrimSpace(outPath) == "" || outPath == "-" {
				fmt.Print(string(data))
				return nil
			}
			if err := os.WriteFile(outPath, data, 0o644); err != nil {
				return err
			}
			fmt.Printf("Exported: %s\n", outPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&outPath, "out", "-", "Output path ('-' for stdout)")
	return cmd
}

func newImportCommand(ctx *commandContext) *cobra.Command {
	var inPath string
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import config/tasks/edges from JSON export",
		Example: "  shelf import --in backup.json\n" +
			"  shelf import < backup.json",
		RunE: func(_ *cobra.Command, _ []string) error {
			data, err := readImportData(inPath)
			if err != nil {
				return err
			}

			var payload shelfExport
			if err := json.Unmarshal(data, &payload); err != nil {
				return fmt.Errorf("invalid import JSON: %w", err)
			}
			if payload.Version != 1 {
				return fmt.Errorf("unsupported export version: %d", payload.Version)
			}

			if err := prepareUndoSnapshot(ctx.rootDir, "import"); err != nil {
				return err
			}
			if err := restoreFromExport(ctx.rootDir, payload); err != nil {
				return err
			}
			fmt.Printf("Imported: tasks=%d edge_files=%d\n", len(payload.Tasks), len(payload.Edges))
			return nil
		},
	}
	cmd.Flags().StringVar(&inPath, "in", "-", "Input path ('-' for stdin)")
	return cmd
}

func buildShelfExport(rootDir string) (shelfExport, error) {
	cfg, err := shelf.LoadConfig(rootDir)
	if err != nil {
		return shelfExport{}, err
	}
	tasks, err := shelf.NewTaskStore(rootDir).List()
	if err != nil {
		return shelfExport{}, err
	}
	edges, err := readAllEdgeFiles(rootDir)
	if err != nil {
		return shelfExport{}, err
	}
	return shelfExport{
		Version:    1,
		ExportedAt: time.Now().Local().Round(time.Second).Format(time.RFC3339),
		Config:     cfg,
		Tasks:      tasks,
		Edges:      edges,
	}, nil
}

func readAllEdgeFiles(rootDir string) (map[string][]shelf.Edge, error) {
	result := map[string][]shelf.Edge{}
	edgeDir := shelf.EdgesDir(rootDir)
	entries, err := os.ReadDir(edgeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		srcID := strings.TrimSuffix(entry.Name(), ".toml")
		data, err := os.ReadFile(filepath.Join(edgeDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		edges, err := shelf.ParseEdgesTOML(data)
		if err != nil {
			return nil, err
		}
		result[srcID] = edges
	}
	return result, nil
}

func readImportData(inPath string) ([]byte, error) {
	if strings.TrimSpace(inPath) == "" || inPath == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(inPath)
}

func restoreFromExport(rootDir string, payload shelfExport) error {
	if err := payload.Config.Validate(); err != nil {
		return err
	}
	if err := os.RemoveAll(shelf.TasksDir(rootDir)); err != nil {
		return err
	}
	if err := os.RemoveAll(shelf.EdgesDir(rootDir)); err != nil {
		return err
	}
	if err := os.MkdirAll(shelf.TasksDir(rootDir), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(shelf.EdgesDir(rootDir), 0o755); err != nil {
		return err
	}
	if err := shelf.SaveConfig(rootDir, payload.Config); err != nil {
		return err
	}

	taskStore := shelf.NewTaskStore(rootDir)
	slices.SortFunc(payload.Tasks, func(a, b shelf.Task) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	for _, task := range payload.Tasks {
		if err := taskStore.Upsert(task); err != nil {
			return err
		}
	}

	edgeStore := shelf.NewEdgeStore(rootDir)
	srcIDs := make([]string, 0, len(payload.Edges))
	for srcID := range payload.Edges {
		srcIDs = append(srcIDs, srcID)
	}
	slices.Sort(srcIDs)
	for _, srcID := range srcIDs {
		if err := edgeStore.SetOutbound(srcID, payload.Edges[srcID]); err != nil {
			return err
		}
	}
	return nil
}
