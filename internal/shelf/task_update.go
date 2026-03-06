package shelf

import (
	"fmt"
	"strings"
	"time"
)

type SetTaskInput struct {
	Title       *string
	Kind        *Kind
	Status      *Status
	DueOn       *string
	RepeatEvery *string
	ArchivedAt  *string
	Parent      *string
	Body        *string
	AppendBody  *string
}

func SetTask(rootDir, taskID string, input SetTaskInput) (Task, error) {
	store := NewTaskStore(rootDir)
	task, err := store.Get(taskID)
	if err != nil {
		return Task{}, err
	}

	cfg, err := LoadConfig(rootDir)
	if err != nil {
		return Task{}, err
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return Task{}, fmt.Errorf("title は空にできません")
		}
		task.Title = title
	}

	if input.Kind != nil {
		if err := cfg.ValidateKind(*input.Kind); err != nil {
			return Task{}, err
		}
		task.Kind = *input.Kind
	}

	if input.Status != nil {
		if err := cfg.ValidateStatus(*input.Status); err != nil {
			return Task{}, err
		}
		task.Status = *input.Status
	}

	if input.DueOn != nil {
		dueOn, err := NormalizeDueOn(*input.DueOn)
		if err != nil {
			return Task{}, err
		}
		task.DueOn = dueOn
	}
	if input.RepeatEvery != nil {
		repeatEvery, err := NormalizeRepeatEvery(*input.RepeatEvery)
		if err != nil {
			return Task{}, err
		}
		task.RepeatEvery = repeatEvery
	}
	if input.ArchivedAt != nil {
		archivedAt, err := normalizeArchivedAt(*input.ArchivedAt)
		if err != nil {
			return Task{}, err
		}
		task.ArchivedAt = archivedAt
	}

	if input.Parent != nil {
		parent := normalizeParent(*input.Parent)
		if err := validateParentUpdate(rootDir, taskID, parent); err != nil {
			return Task{}, err
		}
		task.Parent = parent
	}

	if input.Body != nil {
		task.Body = *input.Body
	}
	if input.AppendBody != nil {
		appendix := *input.AppendBody
		if task.Body != "" && appendix != "" && !strings.HasSuffix(task.Body, "\n") {
			task.Body += "\n"
		}
		task.Body += appendix
	}

	task.UpdatedAt = time.Now().Local().Round(time.Second)
	if err := store.Update(task); err != nil {
		return Task{}, err
	}
	return task, nil
}

func validateParentUpdate(rootDir, taskID, parent string) error {
	if parent == "" {
		return nil
	}
	if parent == taskID {
		return fmt.Errorf("parent に自分自身は指定できません")
	}

	store := NewTaskStore(rootDir)
	tasks, err := store.List()
	if err != nil {
		return err
	}

	byID := make(map[string]Task, len(tasks))
	for _, task := range tasks {
		byID[task.ID] = task
	}
	if _, ok := byID[parent]; !ok {
		return fmt.Errorf("parent が存在しません: %s", parent)
	}

	cur := parent
	visited := map[string]struct{}{}
	for cur != "" {
		if cur == taskID {
			return fmt.Errorf("parent 循環が発生します: %s -> %s", taskID, parent)
		}
		if _, ok := visited[cur]; ok {
			break
		}
		visited[cur] = struct{}{}
		next, ok := byID[cur]
		if !ok {
			break
		}
		cur = next.Parent
	}
	return nil
}
