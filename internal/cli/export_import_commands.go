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

type importSummary struct {
	Mode          string `json:"mode"`
	CurrentTasks  int    `json:"current_tasks"`
	IncomingTasks int    `json:"incoming_tasks"`
	ResultTasks   int    `json:"result_tasks"`
	CreateTasks   int    `json:"create_tasks"`
	UpdateTasks   int    `json:"update_tasks"`
	CurrentEdges  int    `json:"current_edge_files"`
	IncomingEdges int    `json:"incoming_edge_files"`
	ResultEdges   int    `json:"result_edge_files"`
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
	var (
		inPath       string
		dryRun       bool
		validateOnly bool
		merge        bool
		replace      bool
	)
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import config/tasks/edges from JSON export",
		Example: "  shelf import --in backup.json\n" +
			"  shelf import --dry-run --in backup.json\n" +
			"  shelf import --merge --in backup.json",
		RunE: func(_ *cobra.Command, _ []string) error {
			if merge && replace {
				return fmt.Errorf("--merge と --replace は同時に指定できません")
			}
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

			mode := "replace"
			if merge {
				mode = "merge"
			}
			if !merge && !replace {
				replace = true
			}

			resultPayload := payload
			if merge {
				merged, err := buildMergedImport(ctx.rootDir, payload)
				if err != nil {
					return err
				}
				resultPayload = merged
			}
			if err := validateImportPayload(resultPayload); err != nil {
				return err
			}

			summary, err := buildImportSummary(ctx.rootDir, payload, resultPayload, mode)
			if err != nil {
				return err
			}
			if validateOnly {
				fmt.Printf("Import validation OK (%s): current_tasks=%d incoming_tasks=%d result_tasks=%d\n", summary.Mode, summary.CurrentTasks, summary.IncomingTasks, summary.ResultTasks)
				return nil
			}
			if dryRun {
				data, err := json.MarshalIndent(summary, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			if err := prepareUndoSnapshot(ctx.rootDir, "import"); err != nil {
				return err
			}
			if err := restoreFromExport(ctx.rootDir, resultPayload); err != nil {
				return err
			}
			fmt.Printf("Imported (%s): tasks=%d edge_files=%d\n", summary.Mode, summary.ResultTasks, summary.ResultEdges)
			return nil
		},
	}
	cmd.Flags().StringVar(&inPath, "in", "-", "Input path ('-' for stdin)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show import summary without writing")
	cmd.Flags().BoolVar(&validateOnly, "validate-only", false, "Validate import payload without writing")
	cmd.Flags().BoolVar(&merge, "merge", false, "Merge incoming data into current shelf (incoming wins on conflict)")
	cmd.Flags().BoolVar(&replace, "replace", false, "Replace current shelf data with incoming payload")
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
	if err := validateImportPayload(payload); err != nil {
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

func validateImportPayload(payload shelfExport) error {
	if err := payload.Config.Validate(); err != nil {
		return err
	}
	for _, task := range payload.Tasks {
		if _, err := shelf.FormatTaskMarkdown(task); err != nil {
			return err
		}
	}
	for srcID, edges := range payload.Edges {
		if strings.TrimSpace(srcID) == "" {
			return fmt.Errorf("edges source ID is required")
		}
		_ = shelf.FormatEdgesTOML(edges)
	}
	return nil
}

func buildMergedImport(rootDir string, incoming shelfExport) (shelfExport, error) {
	currentTasks, err := shelf.NewTaskStore(rootDir).List()
	if err != nil {
		return shelfExport{}, err
	}
	currentEdges, err := readAllEdgeFiles(rootDir)
	if err != nil {
		return shelfExport{}, err
	}

	taskMap := make(map[string]shelf.Task, len(currentTasks)+len(incoming.Tasks))
	for _, task := range currentTasks {
		taskMap[task.ID] = task
	}
	for _, task := range incoming.Tasks {
		taskMap[task.ID] = task
	}
	mergedTasks := make([]shelf.Task, 0, len(taskMap))
	for _, task := range taskMap {
		mergedTasks = append(mergedTasks, task)
	}
	slices.SortFunc(mergedTasks, func(a, b shelf.Task) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})

	mergedEdges := make(map[string][]shelf.Edge, len(currentEdges)+len(incoming.Edges))
	for srcID, edges := range currentEdges {
		mergedEdges[srcID] = edges
	}
	for srcID, edges := range incoming.Edges {
		mergedEdges[srcID] = edges
	}

	return shelfExport{
		Version:    incoming.Version,
		ExportedAt: incoming.ExportedAt,
		Config:     incoming.Config,
		Tasks:      mergedTasks,
		Edges:      mergedEdges,
	}, nil
}

func buildImportSummary(rootDir string, incoming shelfExport, result shelfExport, mode string) (importSummary, error) {
	currentTasks, err := shelf.NewTaskStore(rootDir).List()
	if err != nil {
		return importSummary{}, err
	}
	currentEdges, err := readAllEdgeFiles(rootDir)
	if err != nil {
		return importSummary{}, err
	}

	currentByID := make(map[string]shelf.Task, len(currentTasks))
	for _, task := range currentTasks {
		currentByID[task.ID] = task
	}
	createCount := 0
	updateCount := 0
	for _, task := range incoming.Tasks {
		if _, ok := currentByID[task.ID]; ok {
			updateCount++
		} else {
			createCount++
		}
	}

	return importSummary{
		Mode:          mode,
		CurrentTasks:  len(currentTasks),
		IncomingTasks: len(incoming.Tasks),
		ResultTasks:   len(result.Tasks),
		CreateTasks:   createCount,
		UpdateTasks:   updateCount,
		CurrentEdges:  len(currentEdges),
		IncomingEdges: len(incoming.Edges),
		ResultEdges:   len(result.Edges),
	}, nil
}
