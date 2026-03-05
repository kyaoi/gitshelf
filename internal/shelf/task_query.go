package shelf

import (
	"fmt"
	"slices"
	"strings"
)

type TaskFilter struct {
	Kind   Kind
	Status Status
	Parent string
	Search string
	Limit  int
}

func ListTasks(rootDir string, filter TaskFilter) ([]Task, error) {
	store := NewTaskStore(rootDir)
	tasks, err := store.List()
	if err != nil {
		return nil, err
	}

	filtered := make([]Task, 0, len(tasks))
	search := strings.ToLower(strings.TrimSpace(filter.Search))
	parent := normalizeParent(filter.Parent)

	for _, task := range tasks {
		if filter.Kind != "" && task.Kind != filter.Kind {
			continue
		}
		if filter.Status != "" && task.Status != filter.Status {
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
