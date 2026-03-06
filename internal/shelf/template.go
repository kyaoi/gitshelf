package shelf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type TaskTemplate struct {
	Version int                `json:"version"`
	Name    string             `json:"name"`
	Tasks   []TaskTemplateTask `json:"tasks"`
}

type TaskTemplateTask struct {
	Key         string   `json:"key"`
	ParentKey   string   `json:"parent_key,omitempty"`
	Title       string   `json:"title"`
	Kind        Kind     `json:"kind"`
	Status      Status   `json:"status"`
	Tags        []string `json:"tags,omitempty"`
	DueOn       string   `json:"due_on,omitempty"`
	RepeatEvery string   `json:"repeat_every,omitempty"`
	Body        string   `json:"body,omitempty"`
}

func SaveTemplate(rootDir string, tpl TaskTemplate) error {
	tpl.Name = strings.TrimSpace(tpl.Name)
	if err := validateTemplate(tpl); err != nil {
		return err
	}
	data, err := json.MarshalIndent(tpl, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return atomicWriteFile(templatePath(rootDir, tpl.Name), data, 0o644)
}

func LoadTemplate(rootDir string, name string) (TaskTemplate, error) {
	path := templatePath(rootDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return TaskTemplate{}, err
	}
	var tpl TaskTemplate
	if err := json.Unmarshal(data, &tpl); err != nil {
		return TaskTemplate{}, err
	}
	if err := validateTemplate(tpl); err != nil {
		return TaskTemplate{}, err
	}
	return tpl, nil
}

func DeleteTemplate(rootDir string, name string) error {
	return os.Remove(templatePath(rootDir, name))
}

func ListTemplateNames(rootDir string) ([]string, error) {
	entries, err := os.ReadDir(TemplatesDir(rootDir))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), ".json"))
	}
	sort.Strings(names)
	return names, nil
}

func BuildTemplateFromTask(rootDir string, name string, taskID string) (TaskTemplate, error) {
	nodes, err := BuildTree(rootDir, TreeOptions{FromID: taskID, IncludeArchived: true})
	if err != nil {
		return TaskTemplate{}, err
	}
	if len(nodes) == 0 {
		return TaskTemplate{}, fmt.Errorf("task not found: %s", taskID)
	}
	tpl := TaskTemplate{
		Version: 1,
		Name:    strings.TrimSpace(name),
		Tasks:   make([]TaskTemplateTask, 0),
	}
	included := map[string]struct{}{}
	var visit func(node TreeNode)
	visit = func(node TreeNode) {
		included[node.Task.ID] = struct{}{}
		tpl.Tasks = append(tpl.Tasks, TaskTemplateTask{
			Key:         node.Task.ID,
			ParentKey:   node.Task.Parent,
			Title:       node.Task.Title,
			Kind:        node.Task.Kind,
			Status:      node.Task.Status,
			Tags:        NormalizeTags(node.Task.Tags),
			DueOn:       node.Task.DueOn,
			RepeatEvery: node.Task.RepeatEvery,
			Body:        node.Task.Body,
		})
		for _, child := range node.Children {
			visit(child)
		}
	}
	for _, node := range nodes {
		visit(node)
	}
	for i := range tpl.Tasks {
		if _, ok := included[tpl.Tasks[i].ParentKey]; !ok {
			tpl.Tasks[i].ParentKey = ""
		}
	}
	return tpl, nil
}

func ApplyTemplate(rootDir string, tpl TaskTemplate, parentID string, titlePrefix string) ([]Task, error) {
	if err := validateTemplate(tpl); err != nil {
		return nil, err
	}
	parentID = normalizeParent(parentID)
	if parentID != "" {
		if _, err := NewTaskStore(rootDir).Get(parentID); err != nil {
			return nil, fmt.Errorf("parent が存在しません: %s", parentID)
		}
	}

	created := make([]Task, 0, len(tpl.Tasks))
	createdByKey := make(map[string]Task, len(tpl.Tasks))
	for _, item := range tpl.Tasks {
		resolvedParent := parentID
		if item.ParentKey != "" {
			parentTask, ok := createdByKey[item.ParentKey]
			if ok {
				resolvedParent = parentTask.ID
			}
		}
		title := item.Title
		if titlePrefix != "" {
			title = strings.TrimSpace(titlePrefix + title)
		}
		task, err := AddTask(rootDir, AddTaskInput{
			Title:       title,
			Kind:        item.Kind,
			Status:      item.Status,
			Tags:        item.Tags,
			DueOn:       item.DueOn,
			RepeatEvery: item.RepeatEvery,
			Parent:      resolvedParent,
			Body:        item.Body,
		})
		if err != nil {
			return nil, err
		}
		created = append(created, task)
		createdByKey[item.Key] = task
	}
	return created, nil
}

func templatePath(rootDir string, name string) string {
	return filepath.Join(TemplatesDir(rootDir), strings.TrimSpace(name)+".json")
}

func validateTemplate(tpl TaskTemplate) error {
	if strings.TrimSpace(tpl.Name) == "" {
		return fmt.Errorf("template name is required")
	}
	if tpl.Version == 0 {
		tpl.Version = 1
	}
	if tpl.Version != 1 {
		return fmt.Errorf("unsupported template version: %d", tpl.Version)
	}
	if len(tpl.Tasks) == 0 {
		return fmt.Errorf("template tasks is empty")
	}
	seen := map[string]struct{}{}
	rootCount := 0
	for _, item := range tpl.Tasks {
		if strings.TrimSpace(item.Key) == "" {
			return fmt.Errorf("template task key is required")
		}
		if _, ok := seen[item.Key]; ok {
			return fmt.Errorf("duplicate template task key: %s", item.Key)
		}
		seen[item.Key] = struct{}{}
		if strings.TrimSpace(item.Title) == "" {
			return fmt.Errorf("template task title is required")
		}
		if strings.TrimSpace(string(item.Kind)) == "" {
			return fmt.Errorf("template task kind is required")
		}
		if strings.TrimSpace(string(item.Status)) == "" {
			return fmt.Errorf("template task status is required")
		}
		if strings.TrimSpace(item.ParentKey) == "" {
			rootCount++
		}
	}
	if rootCount == 0 {
		return fmt.Errorf("template must have at least one root task")
	}
	for _, item := range tpl.Tasks {
		if item.ParentKey == "" {
			continue
		}
		if _, ok := seen[item.ParentKey]; !ok {
			return fmt.Errorf("template parent key does not exist: %s", item.ParentKey)
		}
	}
	return nil
}
