package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

func selectTaskIDIfMissing(
	ctx *commandContext,
	args []string,
	prompt string,
	filterFn func(shelf.Task) bool,
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

	options := make([]interactive.Option, 0, len(candidates))
	for _, task := range candidates {
		options = append(options, interactive.Option{
			Value:      task.ID,
			Label:      fmt.Sprintf("[%s] %s  (%s/%s)", shelf.ShortID(task.ID), task.Title, task.Kind, task.State),
			SearchText: fmt.Sprintf("%s %s %s", task.ID, shelf.ShortID(task.ID), task.Title),
		})
	}
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

	options := []interactive.Option{{
		Value:      "root",
		Label:      "0: [root] 親なし",
		SearchText: "root",
	}}
	for _, task := range tasks {
		if task.ID == currentID {
			continue
		}
		options = append(options, interactive.Option{
			Value:      task.ID,
			Label:      fmt.Sprintf("[%s] %s  (%s/%s)", shelf.ShortID(task.ID), task.Title, task.Kind, task.State),
			SearchText: fmt.Sprintf("%s %s %s", task.ID, shelf.ShortID(task.ID), task.Title),
		})
	}
	selected, err := interactive.Select("Parent を選択してください（0=root）", options)
	if err != nil {
		return "", err
	}
	return selected.Value, nil
}
