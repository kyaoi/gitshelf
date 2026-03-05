package shelf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type TaskStore struct {
	rootDir string
}

func NewTaskStore(rootDir string) *TaskStore {
	return &TaskStore{rootDir: rootDir}
}

func (s *TaskStore) Create(task Task) error {
	path := s.taskPath(task.ID)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("task already exists: %s", task.ID)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to access task file %s: %w", path, err)
	}

	return s.writeTask(path, task)
}

func (s *TaskStore) Update(task Task) error {
	path := s.taskPath(task.ID)
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("task not found: %s", task.ID)
		}
		return fmt.Errorf("failed to access task file %s: %w", path, err)
	}
	return s.writeTask(path, task)
}

func (s *TaskStore) Upsert(task Task) error {
	return s.writeTask(s.taskPath(task.ID), task)
}

func (s *TaskStore) Get(id string) (Task, error) {
	path := s.taskPath(id)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Task{}, fmt.Errorf("task not found: %s", id)
		}
		return Task{}, fmt.Errorf("failed to read task file %s: %w", path, err)
	}
	task, err := ParseTaskMarkdown(data)
	if err != nil {
		return Task{}, fmt.Errorf("%s: %w", path, err)
	}
	if task.ID != id {
		return Task{}, fmt.Errorf("%s: front matter id %q does not match filename %q", path, task.ID, id)
	}
	return task, nil
}

func (s *TaskStore) List() ([]Task, error) {
	dir := TasksDir(s.rootDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read tasks directory %s: %w", dir, err)
	}

	tasks := make([]Task, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		task, err := s.Get(id)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return tasks, nil
}

func (s *TaskStore) taskPath(id string) string {
	return filepath.Join(TasksDir(s.rootDir), id+".md")
}

func (s *TaskStore) writeTask(path string, task Task) error {
	data, err := FormatTaskMarkdown(task)
	if err != nil {
		return fmt.Errorf("failed to format task %s: %w", task.ID, err)
	}
	if err := atomicWriteFile(path, data, 0o644); err != nil {
		return err
	}
	return nil
}
