package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

type taskQueryRecord struct {
	ID          string   `json:"id"`
	File        string   `json:"file"`
	Title       string   `json:"title"`
	Path        string   `json:"path"`
	Kind        string   `json:"kind"`
	Status      string   `json:"status"`
	Tags        []string `json:"tags,omitempty"`
	DueOn       string   `json:"due_on,omitempty"`
	RepeatEvery string   `json:"repeat_every,omitempty"`
	ArchivedAt  string   `json:"archived_at,omitempty"`
	Parent      string   `json:"parent,omitempty"`
	ParentTitle string   `json:"parent_title,omitempty"`
	ParentPath  string   `json:"parent_path,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
	Body        string   `json:"body,omitempty"`
}

type linkTaskRef struct {
	ID    string `json:"id"`
	File  string `json:"file,omitempty"`
	Title string `json:"title,omitempty"`
	Path  string `json:"path,omitempty"`
}

type edgeQueryRecord struct {
	Direction string      `json:"direction"`
	Type      string      `json:"type"`
	Task      linkTaskRef `json:"task"`
	Other     linkTaskRef `json:"other"`
}

type linkSummaryRecord struct {
	Direction string `json:"direction"`
	Type      string `json:"type"`
	Count     int    `json:"count"`
}

type copyPresetRecord struct {
	Name         string `json:"name"`
	Scope        string `json:"scope"`
	SubtreeStyle string `json:"subtree_style"`
	Template     string `json:"template"`
	JoinWith     string `json:"join_with,omitempty"`
}

func buildTaskQueryRecord(rootDir string, task shelf.Task, byID map[string]shelf.Task) taskQueryRecord {
	record := taskQueryRecord{
		ID:          task.ID,
		File:        taskFilePath(rootDir, task.ID),
		Title:       task.Title,
		Path:        buildTaskPath(task, byID),
		Kind:        string(task.Kind),
		Status:      string(task.Status),
		Tags:        append([]string{}, task.Tags...),
		DueOn:       task.DueOn,
		RepeatEvery: task.RepeatEvery,
		ArchivedAt:  task.ArchivedAt,
		Parent:      task.Parent,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Body:        task.Body,
	}
	if task.Parent != "" {
		if parent, ok := byID[task.Parent]; ok {
			record.ParentTitle = parent.Title
			record.ParentPath = buildTaskPath(parent, byID)
		}
	}
	return record
}

func (record taskQueryRecord) TSVFields() map[string]string {
	return map[string]string{
		"id":           record.ID,
		"title":        record.Title,
		"path":         record.Path,
		"kind":         record.Kind,
		"status":       record.Status,
		"due_on":       record.DueOn,
		"repeat_every": record.RepeatEvery,
		"archived_at":  record.ArchivedAt,
		"parent":       record.Parent,
		"parent_path":  record.ParentPath,
		"tags":         strings.Join(record.Tags, ","),
		"file":         record.File,
		"created_at":   record.CreatedAt,
		"updated_at":   record.UpdatedAt,
		"body":         record.Body,
	}
}

func buildLinkTaskRef(rootDir, taskID string, byID map[string]shelf.Task) linkTaskRef {
	ref := linkTaskRef{ID: taskID}
	if task, ok := byID[taskID]; ok {
		ref.File = taskFilePath(rootDir, task.ID)
		ref.Title = task.Title
		ref.Path = buildTaskPath(task, byID)
	}
	return ref
}

func buildEdgeQueryRecord(rootDir, direction string, sourceID string, targetID string, linkType shelf.LinkType, byID map[string]shelf.Task) edgeQueryRecord {
	if direction == "outbound" {
		return edgeQueryRecord{
			Direction: direction,
			Type:      string(linkType),
			Task:      buildLinkTaskRef(rootDir, sourceID, byID),
			Other:     buildLinkTaskRef(rootDir, targetID, byID),
		}
	}
	return edgeQueryRecord{
		Direction: direction,
		Type:      string(linkType),
		Task:      buildLinkTaskRef(rootDir, targetID, byID),
		Other:     buildLinkTaskRef(rootDir, sourceID, byID),
	}
}

func (record edgeQueryRecord) TSVFields() map[string]string {
	return map[string]string{
		"direction":   record.Direction,
		"type":        record.Type,
		"task_id":     record.Task.ID,
		"task_title":  record.Task.Title,
		"task_path":   record.Task.Path,
		"task_file":   record.Task.File,
		"other_id":    record.Other.ID,
		"other_title": record.Other.Title,
		"other_path":  record.Other.Path,
		"other_file":  record.Other.File,
	}
}

func (record linkSummaryRecord) TSVFields() map[string]string {
	return map[string]string{
		"direction": record.Direction,
		"type":      record.Type,
		"count":     fmt.Sprintf("%d", record.Count),
	}
}

func buildCopyPresetRecord(preset shelf.CopyPreset) copyPresetRecord {
	return copyPresetRecord{
		Name:         preset.Name,
		Scope:        string(preset.Scope),
		SubtreeStyle: string(preset.EffectiveSubtreeStyle()),
		Template:     preset.Template,
		JoinWith:     preset.JoinWith,
	}
}

func (record copyPresetRecord) TSVFields() map[string]string {
	return map[string]string{
		"name":          record.Name,
		"scope":         record.Scope,
		"subtree_style": record.SubtreeStyle,
		"template":      record.Template,
		"join_with":     record.JoinWith,
	}
}

func taskFilePath(rootDir, taskID string) string {
	return filepath.Join(shelf.TasksDir(rootDir), taskID+".md")
}
