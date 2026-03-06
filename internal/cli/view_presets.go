package cli

import (
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func resolveTaskView(rootDir string, name string) (shelf.TaskFilter, error) {
	view := strings.TrimSpace(name)
	if view == "" {
		return shelf.TaskFilter{}, nil
	}
	if filter, ok := builtinTaskView(view); ok {
		return filter, nil
	}
	cfg, err := shelf.LoadConfig(rootDir)
	if err != nil {
		return shelf.TaskFilter{}, err
	}
	if custom, ok := cfg.Views[view]; ok {
		return shelf.TaskFilter{
			Kinds:       custom.Kinds,
			Statuses:    custom.Statuses,
			Tags:        custom.Tags,
			NotKinds:    custom.NotKinds,
			NotStatuses: custom.NotStatuses,
			NotTags:     custom.NotTags,
			ReadyOnly:   custom.ReadyOnly,
			DepsBlocked: custom.DepsBlocked,
			DueBefore:   custom.DueBefore,
			DueAfter:    custom.DueAfter,
			Overdue:     custom.Overdue,
			NoDue:       custom.NoDue,
			Parent:      custom.Parent,
			Search:      custom.Search,
			Limit:       custom.Limit,
		}, nil
	}
	return shelf.TaskFilter{}, fmt.Errorf("unknown view: %s", view)
}

func builtinTaskView(name string) (shelf.TaskFilter, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "active":
		return shelf.TaskFilter{
			NotStatuses: []shelf.Status{"done", "cancelled"},
		}, true
	case "ready":
		return shelf.TaskFilter{
			ReadyOnly: true,
		}, true
	case "blocked":
		return shelf.TaskFilter{
			DepsBlocked: true,
		}, true
	case "overdue":
		return shelf.TaskFilter{
			Overdue: true,
		}, true
	default:
		return shelf.TaskFilter{}, false
	}
}

func mergeTaskFilterWithView(base shelf.TaskFilter, preset shelf.TaskFilter, changed map[string]bool) shelf.TaskFilter {
	filter := base

	if len(filter.Kinds) == 0 && len(preset.Kinds) > 0 {
		filter.Kinds = preset.Kinds
	}
	if len(filter.Statuses) == 0 && len(preset.Statuses) > 0 {
		filter.Statuses = preset.Statuses
	}
	if len(filter.NotKinds) == 0 && len(preset.NotKinds) > 0 {
		filter.NotKinds = preset.NotKinds
	}
	if len(filter.Tags) == 0 && len(preset.Tags) > 0 {
		filter.Tags = preset.Tags
	}
	if len(filter.NotStatuses) == 0 && len(preset.NotStatuses) > 0 {
		filter.NotStatuses = preset.NotStatuses
	}
	if len(filter.NotTags) == 0 && len(preset.NotTags) > 0 {
		filter.NotTags = preset.NotTags
	}

	if !changed["ready"] {
		filter.ReadyOnly = preset.ReadyOnly
	}
	if !changed["blocked-by-deps"] {
		filter.DepsBlocked = preset.DepsBlocked
	}
	if !changed["due-before"] && strings.TrimSpace(filter.DueBefore) == "" {
		filter.DueBefore = preset.DueBefore
	}
	if !changed["due-after"] && strings.TrimSpace(filter.DueAfter) == "" {
		filter.DueAfter = preset.DueAfter
	}
	if !changed["overdue"] {
		filter.Overdue = preset.Overdue
	}
	if !changed["no-due"] {
		filter.NoDue = preset.NoDue
	}
	if !changed["parent"] && strings.TrimSpace(filter.Parent) == "" {
		filter.Parent = preset.Parent
	}
	if !changed["search"] && strings.TrimSpace(filter.Search) == "" {
		filter.Search = preset.Search
	}
	if !changed["limit"] && filter.Limit == 50 && preset.Limit > 0 {
		filter.Limit = preset.Limit
	}

	return filter
}

func treeOptionsFromFilter(base shelf.TreeOptions, filter shelf.TaskFilter) (shelf.TreeOptions, error) {
	if filter.ReadyOnly || filter.DepsBlocked || filter.DueBefore != "" || filter.DueAfter != "" || filter.Overdue || filter.NoDue || filter.Parent != "" || filter.Search != "" || filter.Limit > 0 {
		return shelf.TreeOptions{}, fmt.Errorf("view is not supported for tree: contains non-tree filters")
	}
	opts := base
	if len(opts.Kinds) == 0 && len(filter.Kinds) > 0 {
		opts.Kinds = filter.Kinds
	}
	if len(opts.Statuses) == 0 && len(filter.Statuses) > 0 {
		opts.Statuses = filter.Statuses
	}
	if len(opts.Tags) == 0 && len(filter.Tags) > 0 {
		opts.Tags = filter.Tags
	}
	if len(opts.NotKinds) == 0 && len(filter.NotKinds) > 0 {
		opts.NotKinds = filter.NotKinds
	}
	if len(opts.NotStatuses) == 0 && len(filter.NotStatuses) > 0 {
		opts.NotStatuses = filter.NotStatuses
	}
	if len(opts.NotTags) == 0 && len(filter.NotTags) > 0 {
		opts.NotTags = filter.NotTags
	}
	return opts, nil
}
