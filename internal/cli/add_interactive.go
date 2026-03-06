package cli

import (
	"errors"
	"strings"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

func resolveAddInputInteractive(ctx *commandContext, body string, initialStatus string, initialDue string, initialRepeatEvery string) (shelf.AddTaskInput, error) {
	if !interactive.IsTTY() {
		return shelf.AddTaskInput{}, errors.New("非TTYでは対話入力できません。--title を指定してください")
	}

	cfg, err := shelf.LoadConfig(ctx.rootDir)
	if err != nil {
		return shelf.AddTaskInput{}, err
	}

	title := ""
	kind := cfg.DefaultKind
	status := cfg.DefaultStatus
	due := strings.TrimSpace(initialDue)
	repeatEvery := strings.TrimSpace(initialRepeatEvery)
	parent := "root"
	if initial := strings.TrimSpace(initialStatus); initial != "" {
		status = shelf.Status(initial)
	}

	for {
		titleLabel := title
		if strings.TrimSpace(titleLabel) == "" {
			titleLabel = "(required)"
		}
		dueLabel := due
		if strings.TrimSpace(dueLabel) == "" {
			dueLabel = "(none)"
		}
		repeatLabel := repeatEvery
		if strings.TrimSpace(repeatLabel) == "" {
			repeatLabel = "(none)"
		}
		parentLabel := parent
		if strings.TrimSpace(parentLabel) == "" {
			parentLabel = "root"
		}

		options := []interactive.Option{
			{Value: "title", Label: "Title: " + titleLabel},
			{Value: "kind", Label: "Kind: " + string(kind)},
			{Value: "status", Label: "Status: " + string(status)},
			{Value: "due", Label: "Due: " + dueLabel},
			{Value: "repeat", Label: "Repeat: " + repeatLabel},
			{Value: "parent", Label: "Parent: " + parentLabel},
			{Value: "save", Label: "Create task"},
			{Value: "cancel", Label: "Cancel"},
		}
		selected, err := selectEnumOption("追加項目を入力してください", options)
		if err != nil {
			return shelf.AddTaskInput{}, err
		}

		switch selected.Value {
		case "title":
			next, err := interactive.PromptText("Title を入力してください")
			if err != nil {
				return shelf.AddTaskInput{}, err
			}
			title = next
		case "kind":
			kindOptions := make([]interactive.Option, 0, len(cfg.Kinds))
			for _, value := range cfg.Kinds {
				kindOptions = append(kindOptions, interactive.Option{
					Value:      string(value),
					Label:      string(value),
					SearchText: string(value),
				})
			}
			kindSelected, err := selectEnumOption("Kind を選択してください", kindOptions)
			if err != nil {
				return shelf.AddTaskInput{}, err
			}
			kind = shelf.Kind(kindSelected.Value)
		case "status":
			statusOptions := make([]interactive.Option, 0, len(cfg.Statuses))
			for _, value := range cfg.Statuses {
				statusOptions = append(statusOptions, interactive.Option{
					Value:      string(value),
					Label:      string(value),
					SearchText: string(value),
				})
			}
			statusSelected, err := selectEnumOption("Status を選択してください", statusOptions)
			if err != nil {
				return shelf.AddTaskInput{}, err
			}
			status = shelf.Status(statusSelected.Value)
		case "due":
			dueSelected, err := selectEnumOption("期限を選択してください", []interactive.Option{
				{Value: "none", Label: "(none)", SearchText: "none"},
				{Value: "today", Label: "today", SearchText: "today"},
				{Value: "tomorrow", Label: "tomorrow", SearchText: "tomorrow"},
				{Value: "custom", Label: "custom (YYYY-MM-DD)", SearchText: "custom YYYY-MM-DD"},
			})
			if err != nil {
				return shelf.AddTaskInput{}, err
			}
			switch dueSelected.Value {
			case "none":
				due = ""
			case "today", "tomorrow":
				due = dueSelected.Value
			case "custom":
				customDue, err := interactive.PromptText("期限を入力してください (YYYY-MM-DD, today, tomorrow, 空でクリア)")
				if err != nil {
					return shelf.AddTaskInput{}, err
				}
				due = strings.TrimSpace(customDue)
			}
		case "repeat":
			repeat, err := interactive.PromptText("繰り返し間隔を入力してください (<N>d|<N>w|<N>m|<N>y, 空でクリア)")
			if err != nil {
				return shelf.AddTaskInput{}, err
			}
			repeatEvery = strings.TrimSpace(repeat)
		case "parent":
			taskStore := shelf.NewTaskStore(ctx.rootDir)
			tasks, err := taskStore.List()
			if err != nil {
				return shelf.AddTaskInput{}, err
			}
			parentOptions := buildParentSelectionOptions(tasks, "", ctx.showID)
			parentSelected, err := selectTaskOption("Parent を選択してください", parentOptions)
			if err != nil {
				return shelf.AddTaskInput{}, err
			}
			parent = parentSelected.Value
		case "save":
			if strings.TrimSpace(title) == "" {
				return shelf.AddTaskInput{}, errors.New("title は必須です")
			}
			return shelf.AddTaskInput{
				Title:       title,
				Kind:        kind,
				Status:      status,
				DueOn:       due,
				RepeatEvery: repeatEvery,
				Parent:      parent,
				Body:        body,
			}, nil
		case "cancel":
			return shelf.AddTaskInput{}, interactive.ErrCanceled
		default:
			return shelf.AddTaskInput{}, errors.New("未知の選択です")
		}
	}
}
