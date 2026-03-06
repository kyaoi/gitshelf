package shelf

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type TaskFilter struct {
	Kinds       []Kind
	Statuses    []Status
	NotKinds    []Kind
	NotStatuses []Status
	Parent      string
	Search      string
	Limit       int
}

type TaskReadiness struct {
	Ready               bool
	BlockedByDeps       bool
	UnresolvedDependsOn []string
}

func ListTasks(rootDir string, filter TaskFilter) ([]Task, error) {
	if err := validateTaskFilter(rootDir, filter); err != nil {
		return nil, err
	}

	store := NewTaskStore(rootDir)
	tasks, err := store.List()
	if err != nil {
		return nil, err
	}

	filtered := make([]Task, 0, len(tasks))
	search := strings.ToLower(strings.TrimSpace(filter.Search))
	parent := normalizeParent(filter.Parent)

	for _, task := range tasks {
		if len(filter.Kinds) > 0 && !slices.Contains(filter.Kinds, task.Kind) {
			continue
		}
		if len(filter.Statuses) > 0 && !slices.Contains(filter.Statuses, task.Status) {
			continue
		}
		if slices.Contains(filter.NotKinds, task.Kind) {
			continue
		}
		if slices.Contains(filter.NotStatuses, task.Status) {
			continue
		}
		if filter.Parent != "" {
			if parent == "" && task.Parent != "" {
				continue
			}
			if parent != "" && task.Parent != parent {
				continue
			}
		}
		if search != "" {
			target := strings.ToLower(task.Title + "\n" + task.Body)
			if !strings.Contains(target, search) {
				continue
			}
		}
		filtered = append(filtered, task)
	}

	slices.SortFunc(filtered, func(a, b Task) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})

	if filter.Limit > 0 && len(filtered) > filter.Limit {
		filtered = filtered[:filter.Limit]
	}
	return filtered, nil
}

func EnsureTaskExists(rootDir string, taskID string) (Task, error) {
	task, err := NewTaskStore(rootDir).Get(taskID)
	if err != nil {
		return Task{}, fmt.Errorf("task %s の取得に失敗しました: %w", taskID, err)
	}
	return task, nil
}

func BuildTaskReadiness(rootDir string) (map[string]TaskReadiness, error) {
	tasks, err := NewTaskStore(rootDir).List()
	if err != nil {
		return nil, err
	}
	byID := make(map[string]Task, len(tasks))
	for _, task := range tasks {
		byID[task.ID] = task
	}

	dependsOnByTask, err := loadDependsOnEdges(rootDir)
	if err != nil {
		return nil, err
	}

	readiness := make(map[string]TaskReadiness, len(tasks))
	for _, task := range tasks {
		dependencies := dependsOnByTask[task.ID]
		unresolved := make([]string, 0, len(dependencies))
		for _, depID := range dependencies {
			depTask, ok := byID[depID]
			if !ok || !isDependencyResolved(depTask.Status) {
				unresolved = append(unresolved, depID)
			}
		}
		slices.Sort(unresolved)
		blocked := len(unresolved) > 0
		ready := (task.Status == Status("open") || task.Status == Status("in_progress")) && !blocked
		readiness[task.ID] = TaskReadiness{
			Ready:               ready,
			BlockedByDeps:       blocked,
			UnresolvedDependsOn: unresolved,
		}
	}
	return readiness, nil
}

func isDependencyResolved(status Status) bool {
	return status == Status("done") || status == Status("cancelled")
}

func loadDependsOnEdges(rootDir string) (map[string][]string, error) {
	result := map[string][]string{}
	dir := EdgesDir(rootDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, fmt.Errorf("failed to read edges directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		srcID := strings.TrimSuffix(entry.Name(), ".toml")
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		edges, err := ParseEdgesTOML(data)
		if err != nil {
			return nil, err
		}
		for _, edge := range edges {
			if edge.Type != LinkType("depends_on") {
				continue
			}
			result[srcID] = append(result[srcID], edge.To)
		}
	}
	for src := range result {
		slices.Sort(result[src])
	}
	return result, nil
}

func validateTaskFilter(rootDir string, filter TaskFilter) error {
	cfg, err := LoadConfig(rootDir)
	if err != nil {
		return err
	}
	for _, kind := range filter.Kinds {
		if err := cfg.ValidateKind(kind); err != nil {
			return err
		}
	}
	for _, kind := range filter.NotKinds {
		if err := cfg.ValidateKind(kind); err != nil {
			return err
		}
	}
	for _, status := range filter.Statuses {
		if err := cfg.ValidateStatus(status); err != nil {
			return err
		}
	}
	for _, status := range filter.NotStatuses {
		if err := cfg.ValidateStatus(status); err != nil {
			return err
		}
	}
	return nil
}
