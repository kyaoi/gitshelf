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
		title      string
		kind       string
		status     string
		due        string
		clearDue   bool
		parent     string
		body       string
		appendBody string
	)

	cmd := &cobra.Command{
		Use:   "set <id>",
		Short: "Update task fields",
		Args:  cobra.MaximumNArgs(1),
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
			if cmd.Flags().Changed("parent") {
				input.Parent = &parent
			}
			if cmd.Flags().Changed("body") {
				input.Body = &body
			}
			if cmd.Flags().Changed("append-body") {
				input.AppendBody = &appendBody
			}

			if input.Title == nil && input.Kind == nil && input.Status == nil && input.DueOn == nil && input.Parent == nil && input.Body == nil && input.AppendBody == nil {
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
					return errors.New("更新対象がありません。--title/--kind/--status/--due/--clear-due/--parent/--body/--append-body を指定してください")
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
	currentParent := task.Parent
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
		options := []interactive.Option{
			{Value: "title", Label: fmt.Sprintf("Title: %s", currentTitle)},
			{Value: "kind", Label: fmt.Sprintf("Kind: %s", currentKind)},
			{Value: "status", Label: fmt.Sprintf("Status: %s", currentStatus)},
			{Value: "due", Label: fmt.Sprintf("Due: %s", dueLabel)},
			{Value: "parent", Label: fmt.Sprintf("Parent: %s", parentLabel)},
			{Value: "save", Label: "Save changes"},
			{Value: "cancel", Label: "Cancel"},
		}
		selected, err := interactive.Select("更新項目を選択してください", options)
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
			selectedKind, err := interactive.Select("Kind を選択してください", kindOptions)
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
			selectedStatus, err := interactive.Select("Status を選択してください", statusOptions)
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
		case "parent":
			parent, err := selectParentIfMissing(ctx, id, "")
			if err != nil {
				return shelf.SetTaskInput{}, err
			}
			currentParent = parent
			input.Parent = &parent
			changed = true
		case "save":
			if !changed {
				return shelf.SetTaskInput{}, errors.New("更新対象がありません")
			}
			return input, nil
		case "cancel":
			return shelf.SetTaskInput{}, interactive.ErrCanceled
		default:
			return shelf.SetTaskInput{}, fmt.Errorf("未知の選択肢です: %s", selected.Value)
		}
	}
}

func newMvCommand(ctx *commandContext) *cobra.Command {
	var parent string
	cmd := &cobra.Command{
		Use:   "mv <id>",
		Short: "Move task under a new parent",
		Args:  cobra.MaximumNArgs(1),
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
		Use:   use + " <id>",
		Short: short,
		Args:  cobra.MaximumNArgs(1),
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
