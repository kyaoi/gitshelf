package cli

import (
	"errors"
	"strings"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

func resolveAddInputInteractive(ctx *commandContext, body string, initialStatus string) (shelf.AddTaskInput, error) {
	if !interactive.IsTTY() {
		return shelf.AddTaskInput{}, errors.New("非TTYでは対話入力できません。--title を指定してください")
	}

	cfg, err := shelf.LoadConfig(ctx.rootDir)
	if err != nil {
		return shelf.AddTaskInput{}, err
	}

	title, err := interactive.PromptText("Title を入力してください")
	if err != nil {
		return shelf.AddTaskInput{}, err
	}
	if strings.TrimSpace(title) == "" {
		return shelf.AddTaskInput{}, errors.New("title は必須です")
	}

	kindOptions := make([]interactive.Option, 0, len(cfg.Kinds))
	for _, kind := range cfg.Kinds {
		kindOptions = append(kindOptions, interactive.Option{
			Value:      string(kind),
			Label:      string(kind),
			SearchText: string(kind),
		})
	}
	kindSelected, err := interactive.Select("Kind を選択してください", kindOptions)
	if err != nil {
		return shelf.AddTaskInput{}, err
	}

	selectedStatus := strings.TrimSpace(initialStatus)
	if selectedStatus == "" {
		statusOptions := make([]interactive.Option, 0, len(cfg.Statuses))
		for _, status := range cfg.Statuses {
			statusOptions = append(statusOptions, interactive.Option{
				Value:      string(status),
				Label:      string(status),
				SearchText: string(status),
			})
		}
		statusSelected, err := interactive.Select("Status を選択してください", statusOptions)
		if err != nil {
			return shelf.AddTaskInput{}, err
		}
		selectedStatus = statusSelected.Value
	}

	taskStore := shelf.NewTaskStore(ctx.rootDir)
	tasks, err := taskStore.List()
	if err != nil {
		return shelf.AddTaskInput{}, err
	}

	parentOptions := buildParentSelectionOptions(tasks, "", ctx.showID)
	parentSelected, err := interactive.Select("Parent を選択してください", parentOptions)
	if err != nil {
		return shelf.AddTaskInput{}, err
	}

	return shelf.AddTaskInput{
		Title:  title,
		Kind:   shelf.Kind(kindSelected.Value),
		Status: shelf.Status(selectedStatus),
		Parent: parentSelected.Value,
		Body:   body,
	}, nil
}
