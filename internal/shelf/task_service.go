package shelf

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type AddTaskInput struct {
	Title  string
	Kind   Kind
	Status Status
	Parent string
	Body   string
}

func AddTask(rootDir string, input AddTaskInput) (Task, error) {
	cfg, err := LoadConfig(rootDir)
	if err != nil {
		return Task{}, err
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return Task{}, errors.New("title は必須です")
	}

	kind := input.Kind
	if kind == "" {
		kind = cfg.DefaultKind
	}
	if err := cfg.ValidateKind(kind); err != nil {
		return Task{}, err
	}

	status := input.Status
	if status == "" {
		status = cfg.DefaultStatus
	}
	if err := cfg.ValidateStatus(status); err != nil {
		return Task{}, err
	}

	parentID := normalizeParent(input.Parent)
	store := NewTaskStore(rootDir)
	if parentID != "" {
		if _, err := store.Get(parentID); err != nil {
			return Task{}, fmt.Errorf("parent が存在しません: %s", parentID)
		}
	}

	now := time.Now().Local().Round(time.Second)
	task := Task{
		ID:        NewID(),
		Title:     title,
		Kind:      kind,
		Status:    status,
		Parent:    parentID,
		CreatedAt: now,
		UpdatedAt: now,
		Body:      input.Body,
	}

	if err := store.Create(task); err != nil {
		return Task{}, err
	}
	return task, nil
}

func normalizeParent(parent string) string {
	value := strings.TrimSpace(parent)
	if value == "" || strings.EqualFold(value, "root") {
		return ""
	}
	return value
}
