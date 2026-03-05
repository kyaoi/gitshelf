package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

func resolveAddInputInteractive(ctx *commandContext, body string, initialState string) (shelf.AddTaskInput, error) {
	if !interactive.IsTTY() {
		return shelf.AddTaskInput{}, errors.New("非TTYでは対話入力できません。--title を指定してください")
	}

	cfg, err := shelf.LoadConfig(ctx.rootDir)
	if err != nil {
		return shelf.AddTaskInput{}, err
	}

	title, err := promptLine("Title: ")
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

	taskStore := shelf.NewTaskStore(ctx.rootDir)
	tasks, err := taskStore.List()
	if err != nil {
		return shelf.AddTaskInput{}, err
	}

	parentOptions := []interactive.Option{
		{
			Value:      "root",
			Label:      "0: [root] 親なし",
			SearchText: "root",
		},
	}
	for _, task := range tasks {
		label := fmt.Sprintf("[%s] %s  (%s/%s)", shelf.ShortID(task.ID), task.Title, task.Kind, task.State)
		parentOptions = append(parentOptions, interactive.Option{
			Value:      task.ID,
			Label:      label,
			SearchText: fmt.Sprintf("%s %s %s", task.ID, shelf.ShortID(task.ID), task.Title),
		})
	}
	parentSelected, err := interactive.Select("Parent を選択してください（0=root）", parentOptions)
	if err != nil {
		return shelf.AddTaskInput{}, err
	}

	return shelf.AddTaskInput{
		Title:  title,
		Kind:   shelf.Kind(kindSelected.Value),
		State:  shelf.State(initialState),
		Parent: parentSelected.Value,
		Body:   body,
	}, nil
}

func promptLine(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}
