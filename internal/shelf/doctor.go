package shelf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

var ErrDoctorIssues = errors.New("doctor found issues")

type DoctorIssue struct {
	Path    string
	TaskID  string
	Message string
}

type DoctorReport struct {
	Issues []DoctorIssue
}

func (r DoctorReport) HasIssues() bool {
	return len(r.Issues) > 0
}

func RunDoctor(rootDir string) (DoctorReport, error) {
	report := DoctorReport{Issues: make([]DoctorIssue, 0)}

	cfg, err := LoadConfig(rootDir)
	if err != nil {
		return report, err
	}

	tasks, taskFiles, err := loadTasksForDoctor(rootDir, &report)
	if err != nil {
		return report, err
	}

	for id, task := range tasks {
		path := taskFiles[id]
		if err := cfg.ValidateKind(task.Kind); err != nil {
			report.Issues = append(report.Issues, DoctorIssue{Path: path, TaskID: id, Message: err.Error()})
		}
		if err := cfg.ValidateState(task.State); err != nil {
			report.Issues = append(report.Issues, DoctorIssue{Path: path, TaskID: id, Message: err.Error()})
		}
		if task.Parent != "" {
			if _, ok := tasks[task.Parent]; !ok {
				report.Issues = append(report.Issues, DoctorIssue{Path: path, TaskID: id, Message: fmt.Sprintf("parent does not exist: %s", task.Parent)})
			}
		}
	}

	for id := range tasks {
		if hasParentCycle(id, tasks) {
			report.Issues = append(report.Issues, DoctorIssue{
				Path:    taskFiles[id],
				TaskID:  id,
				Message: "parent cycle detected",
			})
		}
	}

	if err := validateEdgesForDoctor(rootDir, cfg, tasks, &report); err != nil {
		return report, err
	}

	if report.HasIssues() {
		return report, ErrDoctorIssues
	}
	return report, nil
}

func loadTasksForDoctor(rootDir string, report *DoctorReport) (map[string]Task, map[string]string, error) {
	dir := TasksDir(rootDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]Task{}, map[string]string{}, nil
		}
		return nil, nil, fmt.Errorf("failed to read tasks directory %s: %w", dir, err)
	}

	tasks := make(map[string]Task, len(entries))
	taskFiles := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			report.Issues = append(report.Issues, DoctorIssue{Path: path, TaskID: id, Message: fmt.Sprintf("failed to read file: %v", err)})
			continue
		}
		task, err := ParseTaskMarkdown(data)
		if err != nil {
			report.Issues = append(report.Issues, DoctorIssue{Path: path, TaskID: id, Message: err.Error()})
			continue
		}
		if task.ID != id {
			report.Issues = append(report.Issues, DoctorIssue{
				Path:    path,
				TaskID:  id,
				Message: fmt.Sprintf("front matter id %q does not match filename %q", task.ID, id),
			})
			continue
		}
		tasks[id] = task
		taskFiles[id] = path
	}
	return tasks, taskFiles, nil
}

func hasParentCycle(start string, tasks map[string]Task) bool {
	seen := map[string]struct{}{}
	cur := start
	for cur != "" {
		if _, ok := seen[cur]; ok {
			return true
		}
		seen[cur] = struct{}{}

		next, ok := tasks[cur]
		if !ok {
			return false
		}
		cur = next.Parent
	}
	return false
}

func validateEdgesForDoctor(rootDir string, cfg Config, tasks map[string]Task, report *DoctorReport) error {
	dir := EdgesDir(rootDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to read edges directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		srcID := strings.TrimSuffix(entry.Name(), ".toml")
		path := filepath.Join(dir, entry.Name())

		if _, ok := tasks[srcID]; !ok {
			report.Issues = append(report.Issues, DoctorIssue{
				Path:    path,
				TaskID:  srcID,
				Message: fmt.Sprintf("source task does not exist: %s", srcID),
			})
		}

		data, err := os.ReadFile(path)
		if err != nil {
			report.Issues = append(report.Issues, DoctorIssue{Path: path, TaskID: srcID, Message: fmt.Sprintf("failed to read file: %v", err)})
			continue
		}

		var raw edgeFile
		if _, err := toml.Decode(string(data), &raw); err != nil {
			report.Issues = append(report.Issues, DoctorIssue{Path: path, TaskID: srcID, Message: fmt.Sprintf("invalid TOML: %v", err)})
			continue
		}

		seen := map[string]struct{}{}
		for _, edge := range raw.Edges {
			if strings.TrimSpace(edge.To) == "" {
				report.Issues = append(report.Issues, DoctorIssue{
					Path:    path,
					TaskID:  srcID,
					Message: "edge.to is empty",
				})
				continue
			}
			if err := cfg.ValidateLinkType(edge.Type); err != nil {
				report.Issues = append(report.Issues, DoctorIssue{
					Path:    path,
					TaskID:  srcID,
					Message: err.Error(),
				})
			}
			if _, ok := tasks[edge.To]; !ok {
				report.Issues = append(report.Issues, DoctorIssue{
					Path:    path,
					TaskID:  srcID,
					Message: fmt.Sprintf("edge destination does not exist: %s", edge.To),
				})
			}

			key := string(edge.Type) + "\x00" + edge.To
			if _, ok := seen[key]; ok {
				report.Issues = append(report.Issues, DoctorIssue{
					Path:    path,
					TaskID:  srcID,
					Message: fmt.Sprintf("duplicate edge found: (%s, %s)", edge.Type, edge.To),
				})
			}
			seen[key] = struct{}{}
		}
	}
	return nil
}
