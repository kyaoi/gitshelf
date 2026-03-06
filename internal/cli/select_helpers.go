package cli

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

func selectTaskIDIfMissing(
	ctx *commandContext,
	args []string,
	prompt string,
	filterFn func(shelf.Task) bool,
	hierarchical bool,
) (string, error) {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return args[0], nil
	}
	if !interactive.IsTTY() {
		return "", errors.New("非TTYでは対話入力できません。<id> を指定してください")
	}

	taskStore := shelf.NewTaskStore(ctx.rootDir)
	tasks, err := taskStore.List()
	if err != nil {
		return "", err
	}
	if len(tasks) == 0 {
		return "", errors.New("タスクがありません")
	}

	candidates := make([]shelf.Task, 0, len(tasks))
	if filterFn == nil {
		candidates = append(candidates, tasks...)
	} else {
		prioritized := make([]shelf.Task, 0, len(tasks))
		others := make([]shelf.Task, 0, len(tasks))
		for _, task := range tasks {
			if filterFn(task) {
				prioritized = append(prioritized, task)
			} else {
				others = append(others, task)
			}
		}
		candidates = append(candidates, prioritized...)
		candidates = append(candidates, others...)
	}
	if len(candidates) == 0 {
		return "", errors.New("選択可能なタスクがありません")
	}

	options := buildTaskSelectionOptions(candidates, taskSelectionBuildOptions{
		Hierarchical:  hierarchical,
		ShowID:        ctx.showID,
		IncludeOrphan: true,
	})
	selected, err := interactive.Select(prompt, options)
	if err != nil {
		return "", err
	}
	return selected.Value, nil
}

func selectParentIfMissing(ctx *commandContext, currentID string, parentFlag string) (string, error) {
	if strings.TrimSpace(parentFlag) != "" {
		return parentFlag, nil
	}
	if !interactive.IsTTY() {
		return "", errors.New("非TTYでは対話入力できません。--parent を指定してください")
	}

	taskStore := shelf.NewTaskStore(ctx.rootDir)
	tasks, err := taskStore.List()
	if err != nil {
		return "", err
	}

	options := buildParentSelectionOptions(tasks, currentID, ctx.showID)
	selected, err := interactive.Select("Parent を選択してください", options)
	if err != nil {
		return "", err
	}
	return selected.Value, nil
}

func buildParentSelectionOptions(tasks []shelf.Task, excludeID string, showID bool) []interactive.Option {
	options := []interactive.Option{{
		Value:      "root",
		Label:      "(root)",
		SearchText: "root",
	}}
	if len(tasks) == 0 {
		return options
	}

	filtered := make([]shelf.Task, 0, len(tasks))
	for _, task := range tasks {
		if task.ID == excludeID {
			continue
		}
		filtered = append(filtered, task)
	}
	if len(filtered) == 0 {
		return options
	}
	return append(options, buildTaskSelectionOptions(filtered, taskSelectionBuildOptions{
		Hierarchical:  true,
		ShowID:        showID,
		IncludeOrphan: false,
	})...)
}

type taskSelectionBuildOptions struct {
	Hierarchical  bool
	ShowID        bool
	IncludeOrphan bool
}

func buildTaskSelectionOptions(tasks []shelf.Task, opts taskSelectionBuildOptions) []interactive.Option {
	if !opts.Hierarchical {
		options := make([]interactive.Option, 0, len(tasks))
		for _, task := range tasks {
			label := fmt.Sprintf("%s  (%s/%s)", task.Title, task.Kind, task.Status)
			if opts.ShowID {
				label = fmt.Sprintf("[%s] %s", shelf.ShortID(task.ID), label)
			}
			options = append(options, interactive.Option{
				Value:      task.ID,
				Label:      label,
				SearchText: fmt.Sprintf("%s %s %s", task.ID, shelf.ShortID(task.ID), task.Title),
				Preview:    buildPreview(task),
			})
		}
		return options
	}

	options := make([]interactive.Option, 0, len(tasks))
	if len(tasks) == 0 {
		return options
	}

	byID := make(map[string]shelf.Task, len(tasks))
	children := make(map[string][]shelf.Task, len(tasks))
	titleCount := map[string]int{}
	titleKindStatusCount := map[string]int{}

	orderByID := make(map[string]int, len(tasks))
	for i, task := range tasks {
		byID[task.ID] = task
		orderByID[task.ID] = i
		titleCount[task.Title]++
		key := task.Title + "\x00" + string(task.Kind) + "\x00" + string(task.Status)
		titleKindStatusCount[key]++
	}
	for _, task := range tasks {
		if task.Parent == "" {
			children[""] = append(children[""], task)
			continue
		}
		if _, ok := byID[task.Parent]; !ok {
			if opts.IncludeOrphan {
				children[""] = append(children[""], task)
			}
			continue
		}
		children[task.Parent] = append(children[task.Parent], task)
	}
	for parent := range children {
		sort.Slice(children[parent], func(i, j int) bool {
			return orderByID[children[parent][i].ID] < orderByID[children[parent][j].ID]
		})
	}

	var visit func(parent string, prefix string, depth int)
	visit = func(parent string, prefix string, depth int) {
		siblings := children[parent]
		for i, task := range siblings {
			isLast := i == len(siblings)-1
			label := taskDisplayLabel(task, titleCount, titleKindStatusCount)
			nextPrefix := prefix

			if depth > 0 {
				branch := "├─ "
				nextPrefix = prefix + "│  "
				if isLast {
					branch = "└─ "
					nextPrefix = prefix + "   "
				}
				label = prefix + branch + label
			}

			if opts.ShowID {
				label = fmt.Sprintf("[%s] %s", shelf.ShortID(task.ID), label)
			}
			options = append(options, interactive.Option{
				Value:      task.ID,
				Label:      label,
				SearchText: fmt.Sprintf("%s %s %s %s %s", task.ID, shelf.ShortID(task.ID), task.Title, task.Kind, task.Status),
				Preview:    buildPreview(task),
			})
			visit(task.ID, nextPrefix, depth+1)
		}
	}
	visit("", "", 0)
	return options
}

func taskDisplayLabel(task shelf.Task, titleCount map[string]int, tksCount map[string]int) string {
	label := task.Title
	if titleCount[task.Title] <= 1 {
		return label
	}
	label = fmt.Sprintf("%s (%s/%s)", task.Title, task.Kind, task.Status)
	key := task.Title + "\x00" + string(task.Kind) + "\x00" + string(task.Status)
	if tksCount[key] > 1 {
		label = fmt.Sprintf("%s [%s]", label, shelf.ShortID(task.ID))
	}
	return label
}

func buildPreview(task shelf.Task) string {
	body := strings.TrimSpace(task.Body)
	if body == "" {
		return "(empty body)"
	}
	return body
}
