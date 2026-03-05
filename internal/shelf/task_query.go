package shelf

import (
	"fmt"
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
