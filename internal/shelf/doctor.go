package shelf

import (
	"bytes"
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
	Issues   []DoctorIssue
	Warnings []DoctorIssue
}

type DoctorOptions struct {
	Strict bool
}

func (r DoctorReport) HasIssues() bool {
	return len(r.Issues) > 0
}

func RunDoctor(rootDir string) (DoctorReport, error) {
	return RunDoctorWithOptions(rootDir, DoctorOptions{})
}

func RunDoctorWithOptions(rootDir string, opts DoctorOptions) (DoctorReport, error) {
	report := DoctorReport{
		Issues:   make([]DoctorIssue, 0),
		Warnings: make([]DoctorIssue, 0),
	}

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
		if err := cfg.ValidateStatus(task.Status); err != nil {
			report.Issues = append(report.Issues, DoctorIssue{Path: path, TaskID: id, Message: err.Error()})
		}
		for _, tag := range NormalizeTags(task.Tags) {
			if err := cfg.ValidateTag(tag); err != nil {
				report.Issues = append(report.Issues, DoctorIssue{Path: path, TaskID: id, Message: err.Error()})
			}
		}
		if task.Parent != "" {
			if _, ok := tasks[task.Parent]; !ok {
				report.Issues = append(report.Issues, DoctorIssue{Path: path, TaskID: id, Message: fmt.Sprintf("parent does not exist: %s", task.Parent)})
			}
		}
		if opts.Strict && task.Kind == "todo" && strings.TrimSpace(task.DueOn) == "" {
			report.Warnings = append(report.Warnings, DoctorIssue{Path: path, TaskID: id, Message: "todo task has no due_on"})
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

func RunDoctorWithFix(rootDir string) (DoctorReport, int, error) {
	return RunDoctorWithFixOptions(rootDir, DoctorOptions{})
}

func RunDoctorWithFixOptions(rootDir string, opts DoctorOptions) (DoctorReport, int, error) {
	fixed, err := applyDoctorFixes(rootDir)
	if err != nil {
		return DoctorReport{}, fixed, err
	}
	report, doctorErr := RunDoctorWithOptions(rootDir, opts)
	return report, fixed, doctorErr
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

func applyDoctorFixes(rootDir string) (int, error) {
	fixedCount := 0
	cfg, err := LoadConfig(rootDir)
	if err != nil {
		return fixedCount, err
	}

	taskDir := TasksDir(rootDir)
	taskEntries, err := os.ReadDir(taskDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fixedCount, fmt.Errorf("failed to read tasks directory %s: %w", taskDir, err)
	}
	knownTaskIDs := map[string]struct{}{}
	for _, entry := range taskEntries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		knownTaskIDs[strings.TrimSuffix(entry.Name(), ".md")] = struct{}{}
	}

	taskStore := NewTaskStore(rootDir)
	for _, entry := range taskEntries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(taskDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		task, err := ParseTaskMarkdown(data)
		if err != nil {
			continue
		}
		changed := false
		if task.Parent != "" {
			if _, ok := knownTaskIDs[task.Parent]; !ok {
				task.Parent = ""
				changed = true
			}
		}
		normalized, err := FormatTaskMarkdown(task)
		if err != nil {
			continue
		}
		if !bytes.Equal(data, normalized) {
			changed = true
		}
		if !changed {
			continue
		}
		if err := taskStore.Update(task); err != nil {
			return fixedCount, err
		}
		fixedCount++
	}

	edgeDir := EdgesDir(rootDir)
	edgeEntries, err := os.ReadDir(edgeDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fixedCount, fmt.Errorf("failed to read edges directory %s: %w", edgeDir, err)
	}
	for _, entry := range edgeEntries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		srcID := strings.TrimSuffix(entry.Name(), ".toml")
		path := filepath.Join(edgeDir, entry.Name())
		if _, ok := knownTaskIDs[srcID]; !ok {
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fixedCount, err
			}
			fixedCount++
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		edges, err := ParseEdgesTOML(data)
		if err != nil {
			continue
		}
		filtered := make([]Edge, 0, len(edges))
		for _, edge := range edges {
			if _, ok := knownTaskIDs[edge.To]; !ok {
				continue
			}
			if err := cfg.ValidateLinkType(edge.Type); err != nil {
				continue
			}
			filtered = append(filtered, edge)
		}
		normalized := FormatEdgesTOML(filtered)
		if len(filtered) == 0 {
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fixedCount, err
			}
			fixedCount++
			continue
		}
		if bytes.Equal(data, normalized) {
			continue
		}
		if err := atomicWriteFile(path, normalized, 0o644); err != nil {
			return fixedCount, err
		}
		fixedCount++
	}

	return fixedCount, nil
}
