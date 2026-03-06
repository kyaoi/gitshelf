package shelf

import "slices"

type TreeNode struct {
	Task     Task
	Children []TreeNode
}

type TreeOptions struct {
	FromID          string
	Status          Status
	Kinds           []Kind
	Statuses        []Status
	Tags            []string
	NotKinds        []Kind
	NotStatuses     []Status
	NotTags         []string
	IncludeArchived bool
	OnlyArchived    bool
	MaxDepth        int
}

func BuildTree(rootDir string, options TreeOptions) ([]TreeNode, error) {
	tasks, err := NewTaskStore(rootDir).List()
	if err != nil {
		return nil, err
	}

	byParent := make(map[string][]Task)
	byID := make(map[string]Task, len(tasks))
	for _, task := range tasks {
		byID[task.ID] = task
		byParent[task.Parent] = append(byParent[task.Parent], task)
	}
	for parent := range byParent {
		slices.SortFunc(byParent[parent], func(a, b Task) int {
			if a.ID < b.ID {
				return -1
			}
			if a.ID > b.ID {
				return 1
			}
			return 0
		})
	}

	startTasks := byParent[""]
	if options.FromID != "" {
		start, ok := byID[options.FromID]
		if !ok {
			return nil, nil
		}
		startTasks = []Task{start}
	}

	nodes := make([]TreeNode, 0, len(startTasks))
	for _, task := range startTasks {
		node := buildTreeNode(task, byParent, options, 0, map[string]struct{}{})
		if node == nil {
			continue
		}
		nodes = append(nodes, *node)
	}
	return nodes, nil
}

func buildTreeNode(task Task, byParent map[string][]Task, options TreeOptions, depth int, path map[string]struct{}) *TreeNode {
	if options.MaxDepth > 0 && depth > options.MaxDepth {
		return nil
	}
	if options.OnlyArchived {
		if task.ArchivedAt == "" {
			return nil
		}
	} else if !options.IncludeArchived && task.ArchivedAt != "" {
		return nil
	}
	if _, ok := path[task.ID]; ok {
		return nil
	}

	nextPath := make(map[string]struct{}, len(path)+1)
	for k, v := range path {
		nextPath[k] = v
	}
	nextPath[task.ID] = struct{}{}

	children := make([]TreeNode, 0)
	for _, child := range byParent[task.ID] {
		childNode := buildTreeNode(child, byParent, options, depth+1, nextPath)
		if childNode != nil {
			children = append(children, *childNode)
		}
	}

	if !matchesTreeFilters(task, options) {
		if len(children) == 0 {
			return nil
		}
		return &TreeNode{
			Task:     task,
			Children: children,
		}
	}
	return &TreeNode{
		Task:     task,
		Children: children,
	}
}

func matchesTreeFilters(task Task, options TreeOptions) bool {
	statuses := options.Statuses
	if len(statuses) == 0 && options.Status != "" {
		statuses = []Status{options.Status}
	}

	if len(options.Kinds) > 0 && !slices.Contains(options.Kinds, task.Kind) {
		return false
	}
	if len(statuses) > 0 && !slices.Contains(statuses, task.Status) {
		return false
	}
	if slices.Contains(options.NotKinds, task.Kind) {
		return false
	}
	if slices.Contains(options.NotStatuses, task.Status) {
		return false
	}
	if len(options.Tags) > 0 && !containsAnyTag(task.Tags, options.Tags) {
		return false
	}
	if containsAnyTag(task.Tags, options.NotTags) {
		return false
	}
	return true
}
