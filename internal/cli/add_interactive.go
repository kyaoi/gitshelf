package cli

import (
	"errors"
	"strings"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

func resolveAddInputInteractive(ctx *commandContext, body string, initialStatus string, initialDue string, initialRepeatEvery string, initialTags []string) (shelf.AddTaskInput, error) {
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
	tags := shelf.NormalizeTags(initialTags)
	due := strings.TrimSpace(initialDue)
	repeatEvery := strings.TrimSpace(initialRepeatEvery)
	parent := "root"
	if initial := strings.TrimSpace(initialStatus); initial != "" {
		status = shelf.Status(initial)
	}

	kindSelected, err := selectEnumOption("Kind を選択してください", enumOptionsFromKinds(cfg.Kinds))
	if err != nil {
		return shelf.AddTaskInput{}, err
	}
	kind = shelf.Kind(kindSelected.Value)

	statusSelected, err := selectEnumOption("Status を選択してください", enumOptionsFromStatuses(cfg.Statuses))
	if err != nil {
		return shelf.AddTaskInput{}, err
	}
	status = shelf.Status(statusSelected.Value)

	for {
		titleLabel := title
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
			{Value: "tags", Label: "Tags: " + formatTagSummary(tags)},
			{Value: "due", Label: "Due: " + dueLabel},
			{Value: "repeat", Label: "Repeat: " + repeatLabel},
			{Value: "parent", Label: "Parent: " + parentLabel},
			{Value: "save", Label: "Create task", Preview: buildAddSummary(title, kind, status, tags, dueLabel, repeatLabel, parentLabel)},
			{Value: "cancel", Label: "Cancel"},
		}
		selected, err := interactive.SelectWithConfig(interactive.SelectConfig{
			Prompt:            "内容を確認してください",
			Options:           options,
			ShowPreview:       false,
			MaxRows:           15,
			HelpText:          selectorHelpText + "  Ctrl+S/Ctrl+Enter: 作成",
			SearchPlaceholder: "検索",
			SubmitValue:       "save",
			SubmitShortcuts:   []string{"Ctrl+S", "Ctrl+Enter"},
		})
		if err != nil {
			return shelf.AddTaskInput{}, err
		}

		switch selected.Value {
		case "title":
			next, err := interactive.PromptText("Title を入力してください")
			if err != nil {
				return shelf.AddTaskInput{}, err
			}
			title = strings.TrimSpace(next)
		case "kind":
			kindSelected, err := selectEnumOption("Kind を選択してください", enumOptionsFromKinds(cfg.Kinds))
			if err != nil {
				return shelf.AddTaskInput{}, err
			}
			kind = shelf.Kind(kindSelected.Value)
		case "status":
			statusSelected, err := selectEnumOption("Status を選択してください", enumOptionsFromStatuses(cfg.Statuses))
			if err != nil {
				return shelf.AddTaskInput{}, err
			}
			status = shelf.Status(statusSelected.Value)
		case "tags":
			selectedTags, err := selectTagsInteractive("Tags を選択してください", cfg.Tags, tags)
			if err != nil {
				return shelf.AddTaskInput{}, err
			}
			tags = selectedTags
		case "due":
			dueSelected, err := selectEnumOption("期限を選択してください", []interactive.Option{
				{Value: "none", Label: "(none)", SearchText: "none"},
				{Value: "today", Label: "today", SearchText: "today"},
				{Value: "tomorrow", Label: "tomorrow", SearchText: "tomorrow"},
				{Value: "next-week", Label: "next-week", SearchText: "next-week"},
				{Value: "custom", Label: "custom", SearchText: "custom YYYY-MM-DD +Nd -Nd mon..sun"},
			})
			if err != nil {
				return shelf.AddTaskInput{}, err
			}
			switch dueSelected.Value {
			case "none":
				due = ""
			case "today", "tomorrow", "next-week":
				due = dueSelected.Value
			case "custom":
				customDue, err := interactive.PromptText("期限を入力してください (YYYY-MM-DD, today, tomorrow, +Nd, -Nd, next-week, mon..sun, 空でクリア)")
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
				next, err := interactive.PromptText("Title は必須です。Title を入力してください")
				if err != nil {
					if errors.Is(err, interactive.ErrCanceled) {
						continue
					}
					return shelf.AddTaskInput{}, err
				}
				title = strings.TrimSpace(next)
				if title == "" {
					continue
				}
			}
			return shelf.AddTaskInput{
				Title:       title,
				Kind:        kind,
				Status:      status,
				Tags:        tags,
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

func enumOptionsFromKinds(values []shelf.Kind) []interactive.Option {
	options := make([]interactive.Option, 0, len(values))
	for _, value := range values {
		options = append(options, interactive.Option{
			Value:      string(value),
			Label:      string(value),
			SearchText: string(value),
		})
	}
	return options
}

func enumOptionsFromStatuses(values []shelf.Status) []interactive.Option {
	options := make([]interactive.Option, 0, len(values))
	for _, value := range values {
		options = append(options, interactive.Option{
			Value:      string(value),
			Label:      string(value),
			SearchText: string(value),
		})
	}
	return options
}

func buildAddSummary(title string, kind shelf.Kind, status shelf.Status, tags []string, due string, repeat string, parent string) string {
	return strings.Join([]string{
		"Title: " + title,
		"Kind: " + string(kind),
		"Status: " + string(status),
		"Tags: " + formatTagSummary(tags),
		"Due: " + due,
		"Repeat: " + repeat,
		"Parent: " + parent,
	}, "\n")
}
