package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newSetCommand(ctx *commandContext) *cobra.Command {
	var (
		title       string
		kind        string
		status      string
		due         string
		clearDue    bool
		repeatEvery string
		clearRepeat bool
		parent      string
		body        string
		appendBody  string
	)

	cmd := &cobra.Command{
		Use:   "set <id>",
		Short: "Update task fields",
		Example: "  shelf set 01ABCDEFG... --status blocked\n" +
			"  shelf set 01ABCDEFG... --due 2026-03-31\n" +
			"  shelf set 01ABCDEFG... --clear-due",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "更新するタスクを選択", nil, true)
			if err != nil {
				return err
			}

			input := shelf.SetTaskInput{}
			if cmd.Flags().Changed("title") {
				input.Title = &title
			}
			if cmd.Flags().Changed("kind") {
				k := shelf.Kind(kind)
				input.Kind = &k
			}
			if cmd.Flags().Changed("status") {
				s := shelf.Status(status)
				input.Status = &s
			}
			if cmd.Flags().Changed("due") && clearDue {
				return errors.New("--due と --clear-due は同時に指定できません")
			}
			if cmd.Flags().Changed("due") {
				input.DueOn = &due
			}
			if clearDue {
				empty := ""
				input.DueOn = &empty
			}
			if cmd.Flags().Changed("repeat-every") && clearRepeat {
				return errors.New("--repeat-every と --clear-repeat は同時に指定できません")
			}
			if cmd.Flags().Changed("repeat-every") {
				input.RepeatEvery = &repeatEvery
			}
			if clearRepeat {
				empty := ""
				input.RepeatEvery = &empty
			}
			if cmd.Flags().Changed("parent") {
				input.Parent = &parent
			}
			if cmd.Flags().Changed("body") {
				input.Body = &body
			}
			if cmd.Flags().Changed("append-body") {
				input.AppendBody = &appendBody
			}

			if input.Title == nil && input.Kind == nil && input.Status == nil && input.DueOn == nil && input.RepeatEvery == nil && input.Parent == nil && input.Body == nil && input.AppendBody == nil {
				if interactive.IsTTY() {
					task, err := shelf.EnsureTaskExists(ctx.rootDir, id)
					if err != nil {
						return err
					}
					cfg, err := shelf.LoadConfig(ctx.rootDir)
					if err != nil {
						return err
					}
					interactiveInput, err := resolveSetInputInteractive(ctx, id, task, cfg)
					if err != nil {
						return err
					}
					input = interactiveInput
				} else {
					return errors.New("更新対象がありません。--title/--kind/--status/--due/--clear-due/--repeat-every/--clear-repeat/--parent/--body/--append-body を指定してください")
				}
			}

			task, err := shelf.SetTask(ctx.rootDir, id, input)
			if err != nil {
				return err
			}
			fmt.Printf("Updated: [%s] %s\n", shelf.ShortID(task.ID), task.Title)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&kind, "kind", "", "New kind")
	cmd.Flags().StringVar(&status, "status", "", "New status")
	cmd.Flags().StringVar(&due, "due", "", "New due date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&clearDue, "clear-due", false, "Clear due date")
	cmd.Flags().StringVar(&repeatEvery, "repeat-every", "", "Repeat interval (<N>d|<N>w|<N>m|<N>y)")
	cmd.Flags().BoolVar(&clearRepeat, "clear-repeat", false, "Clear repeat interval")
	cmd.Flags().StringVar(&parent, "parent", "", "New parent task ID or root")
	cmd.Flags().StringVar(&body, "body", "", "Replace body")
	cmd.Flags().StringVar(&appendBody, "append-body", "", "Append text to body")
	return cmd
}

func resolveSetInputInteractive(ctx *commandContext, id string, task shelf.Task, cfg shelf.Config) (shelf.SetTaskInput, error) {
	input := shelf.SetTaskInput{}
	currentTitle := task.Title
	currentKind := task.Kind
	currentStatus := task.Status
	currentDue := task.DueOn
	currentRepeatEvery := task.RepeatEvery
	currentParent := task.Parent
	currentBody := task.Body
	changed := false

	for {
		parentLabel := "root"
		if currentParent != "" {
			parentLabel = shelf.ShortID(currentParent)
		}
		dueLabel := currentDue
		if strings.TrimSpace(dueLabel) == "" {
			dueLabel = "(none)"
		}
		repeatLabel := currentRepeatEvery
		if strings.TrimSpace(repeatLabel) == "" {
			repeatLabel = "(none)"
		}
		bodyLabel := strings.TrimSpace(currentBody)
		if bodyLabel == "" {
			bodyLabel = "(empty)"
		} else {
			bodyLabel = strings.SplitN(bodyLabel, "\n", 2)[0]
		}
		options := []interactive.Option{
			{Value: "title", Label: fmt.Sprintf("Title: %s", currentTitle)},
			{Value: "kind", Label: fmt.Sprintf("Kind: %s", currentKind)},
			{Value: "status", Label: fmt.Sprintf("Status: %s", currentStatus)},
			{Value: "due", Label: fmt.Sprintf("Due: %s", dueLabel)},
			{Value: "repeat", Label: fmt.Sprintf("Repeat: %s", repeatLabel)},
			{Value: "parent", Label: fmt.Sprintf("Parent: %s", parentLabel)},
			{Value: "body_replace", Label: fmt.Sprintf("Body (replace): %s", bodyLabel)},
			{Value: "body_append", Label: "Body (append)"},
			{Value: "save", Label: "Save changes"},
			{Value: "cancel", Label: "Cancel"},
		}
		selected, err := selectEnumOption("更新項目を選択してください", options)
		if err != nil {
			return shelf.SetTaskInput{}, err
		}

		switch selected.Value {
		case "title":
			title, err := interactive.PromptText("新しい Title を入力してください")
			if err != nil {
				return shelf.SetTaskInput{}, err
			}
			if strings.TrimSpace(title) == "" {
				return shelf.SetTaskInput{}, errors.New("title は空にできません")
			}
			currentTitle = title
			input.Title = &title
			changed = true
		case "kind":
			kindOptions := make([]interactive.Option, 0, len(cfg.Kinds))
			for _, kind := range cfg.Kinds {
				kindOptions = append(kindOptions, interactive.Option{
					Value:      string(kind),
					Label:      string(kind),
					SearchText: string(kind),
				})
			}
			selectedKind, err := selectEnumOption("Kind を選択してください", kindOptions)
			if err != nil {
				return shelf.SetTaskInput{}, err
			}
			kind := shelf.Kind(selectedKind.Value)
			currentKind = kind
			input.Kind = &kind
			changed = true
		case "status":
			statusOptions := make([]interactive.Option, 0, len(cfg.Statuses))
			for _, status := range cfg.Statuses {
				statusOptions = append(statusOptions, interactive.Option{
					Value:      string(status),
					Label:      string(status),
					SearchText: string(status),
				})
			}
			selectedStatus, err := selectEnumOption("Status を選択してください", statusOptions)
			if err != nil {
				return shelf.SetTaskInput{}, err
			}
			status := shelf.Status(selectedStatus.Value)
			currentStatus = status
			input.Status = &status
			changed = true
		case "due":
			due, err := interactive.PromptText("期限を入力してください (YYYY-MM-DD, 空でクリア)")
			if err != nil {
				return shelf.SetTaskInput{}, err
			}
			currentDue = due
			input.DueOn = &due
			changed = true
		case "repeat":
			repeatEvery, err := interactive.PromptText("繰り返し間隔を入力してください (<N>d|<N>w|<N>m|<N>y, 空でクリア)")
			if err != nil {
				return shelf.SetTaskInput{}, err
			}
			currentRepeatEvery = repeatEvery
			input.RepeatEvery = &repeatEvery
			changed = true
		case "parent":
			parent, err := selectParentIfMissing(ctx, id, "")
			if err != nil {
				return shelf.SetTaskInput{}, err
			}
			currentParent = parent
			input.Parent = &parent
			changed = true
		case "body_replace":
			body, err := interactive.PromptText("Body を入力してください（全置換）")
			if err != nil {
				return shelf.SetTaskInput{}, err
			}
			currentBody = body
			input.Body = &body
			input.AppendBody = nil
			changed = true
		case "body_append":
			appendix, err := interactive.PromptText("Body への追記を入力してください")
			if err != nil {
				return shelf.SetTaskInput{}, err
			}
			if appendix == "" {
				continue
			}
			combined := appendix
			if input.AppendBody != nil && *input.AppendBody != "" {
				combined = *input.AppendBody + "\n" + appendix
			}
			input.AppendBody = &combined
			if currentBody != "" && !strings.HasSuffix(currentBody, "\n") {
				currentBody += "\n"
			}
			currentBody += appendix
			changed = true
		case "save":
			if !changed {
				return shelf.SetTaskInput{}, errors.New("更新対象がありません")
			}
			preview := buildSetChangePreview(task, input)
			confirm, err := interactive.SelectWithConfig(interactive.SelectConfig{
				Prompt: "変更内容を確認してください",
				Options: []interactive.Option{
					{Value: "apply", Label: "Apply changes", Preview: preview},
					{Value: "back", Label: "Back to edit", Preview: preview},
					{Value: "cancel", Label: "Cancel", Preview: preview},
				},
				ShowPreview:       true,
				MaxRows:           10,
				HelpText:          selectorHelpText,
				SearchPlaceholder: "確認",
			})
			if err != nil {
				return shelf.SetTaskInput{}, err
			}
			switch confirm.Value {
			case "apply":
				return input, nil
			case "back":
				continue
			case "cancel":
				return shelf.SetTaskInput{}, interactive.ErrCanceled
			default:
				return shelf.SetTaskInput{}, fmt.Errorf("未知の選択肢です: %s", confirm.Value)
			}
		case "cancel":
			return shelf.SetTaskInput{}, interactive.ErrCanceled
		default:
			return shelf.SetTaskInput{}, fmt.Errorf("未知の選択肢です: %s", selected.Value)
		}
	}
}

func buildSetChangePreview(orig shelf.Task, input shelf.SetTaskInput) string {
	lines := make([]string, 0, 8)
	if input.Title != nil {
		lines = append(lines, fmt.Sprintf("Title: %q -> %q", orig.Title, strings.TrimSpace(*input.Title)))
	}
	if input.Kind != nil {
		lines = append(lines, fmt.Sprintf("Kind: %q -> %q", orig.Kind, *input.Kind))
	}
	if input.Status != nil {
		lines = append(lines, fmt.Sprintf("Status: %q -> %q", orig.Status, *input.Status))
	}
	if input.DueOn != nil {
		before := orig.DueOn
		if strings.TrimSpace(before) == "" {
			before = "(none)"
		}
		after := strings.TrimSpace(*input.DueOn)
		if after == "" {
			after = "(none)"
		}
		lines = append(lines, fmt.Sprintf("Due: %q -> %q", before, after))
	}
	if input.RepeatEvery != nil {
		before := orig.RepeatEvery
		if strings.TrimSpace(before) == "" {
			before = "(none)"
		}
		after := strings.TrimSpace(*input.RepeatEvery)
		if after == "" {
			after = "(none)"
		}
		lines = append(lines, fmt.Sprintf("Repeat: %q -> %q", before, after))
	}
	if input.Parent != nil {
		before := orig.Parent
		if strings.TrimSpace(before) == "" {
			before = "root"
		}
		after := strings.TrimSpace(*input.Parent)
		if after == "" {
			after = "root"
		}
		lines = append(lines, fmt.Sprintf("Parent: %q -> %q", before, after))
	}
	if input.Body != nil {
		lines = append(lines, "Body: replace")
	}
	if input.AppendBody != nil {
		lines = append(lines, fmt.Sprintf("Body: append %d chars", len(*input.AppendBody)))
	}
	if len(lines) == 0 {
		return "(no changes)"
	}
	return strings.Join(lines, "\n")
}

func newMvCommand(ctx *commandContext) *cobra.Command {
	var parent string
	cmd := &cobra.Command{
		Use:   "mv <id>",
		Short: "Move task under a new parent",
		Example: "  shelf mv 01ABCDEFG... --parent root\n" +
			"  shelf mv 01ABCDEFG...",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "移動するタスクを選択", nil, true)
			if err != nil {
				return err
			}

			resolvedParent := parent
			if !cmd.Flags().Changed("parent") {
				resolvedParent, err = selectParentIfMissing(ctx, id, parent)
				if err != nil {
					return err
				}
			}

			task, err := shelf.SetTask(ctx.rootDir, id, shelf.SetTaskInput{
				Parent: &resolvedParent,
			})
			if err != nil {
				return err
			}
			parentLabel := "root"
			if task.Parent != "" {
				parentLabel = shelf.ShortID(task.Parent)
			}
			fmt.Printf("Moved: [%s] parent=%s\n", shelf.ShortID(task.ID), parentLabel)
			return nil
		},
	}

	cmd.Flags().StringVar(&parent, "parent", "", "New parent task ID or root")
	return cmd
}

func newDoneCommand(ctx *commandContext) *cobra.Command {
	return newStatusShortcutCommand(ctx, "done", "Shortcut to set --status done", "done", "done にするタスクを選択", "Done")
}

func newStartCommand(ctx *commandContext) *cobra.Command {
	return newStatusShortcutCommand(ctx, "start", "Shortcut to set --status in_progress", "in_progress", "in_progress にするタスクを選択", "Started")
}

func newBlockCommand(ctx *commandContext) *cobra.Command {
	return newStatusShortcutCommand(ctx, "block", "Shortcut to set --status blocked", "blocked", "blocked にするタスクを選択", "Blocked")
}

func newCancelCommand(ctx *commandContext) *cobra.Command {
	return newStatusShortcutCommand(ctx, "cancel", "Shortcut to set --status cancelled", "cancelled", "cancelled にするタスクを選択", "Cancelled")
}

func newStatusShortcutCommand(ctx *commandContext, use string, short string, targetStatus string, prompt string, actionLabel string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     use + " <id>",
		Short:   short,
		Example: fmt.Sprintf("  shelf %s 01ABCDEFG...\n  shelf %s", use, use),
		Args:    cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, prompt, func(task shelf.Task) bool {
				return task.Status != shelf.Status(targetStatus)
			}, true)
			if err != nil {
				return err
			}

			next := shelf.Status(targetStatus)
			task, err := shelf.SetTask(ctx.rootDir, id, shelf.SetTaskInput{
				Status: &next,
			})
			if err != nil {
				return err
			}
			fmt.Printf("%s: [%s] %s\n", actionLabel, shelf.ShortID(task.ID), task.Title)
			return nil
		},
	}
	return cmd
}
