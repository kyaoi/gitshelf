package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newTriageCommand(ctx *commandContext) *cobra.Command {
	var (
		kind       string
		status     string
		limit      int
		autoAction string
	)

	cmd := &cobra.Command{
		Use:   "triage",
		Short: "Process inbox/open tasks quickly",
		Example: "  shelf triage\n" +
			"  shelf triage --auto done\n" +
			"  shelf triage --kind inbox --status open --limit 10",
		RunE: func(_ *cobra.Command, _ []string) error {
			candidates, err := listTriageCandidates(ctx.rootDir, kind, status, limit)
			if err != nil {
				return err
			}
			if len(candidates) == 0 {
				fmt.Println("No triage targets.")
				return nil
			}

			autoAction = strings.ToLower(strings.TrimSpace(autoAction))
			if autoAction != "" {
				updated, err := runTriageAuto(ctx.rootDir, candidates, autoAction)
				if err != nil {
					return err
				}
				fmt.Printf("Triage auto (%s): updated=%d\n", autoAction, updated)
				return nil
			}

			if !interactive.IsTTY() {
				return errors.New("非TTYでは --auto を指定してください")
			}
			return runTriageInteractive(ctx, candidates)
		},
	}

	cmd.Flags().StringVar(&kind, "kind", "inbox", "Target kind for triage")
	cmd.Flags().StringVar(&status, "status", "open", "Target status for triage")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of triage targets")
	cmd.Flags().StringVar(&autoAction, "auto", "", "Auto action: done|start|block|cancel|reopen|archive")
	return cmd
}

func listTriageCandidates(rootDir string, kind string, status string, limit int) ([]shelf.Task, error) {
	return shelf.ListTasks(rootDir, shelf.TaskFilter{
		Kinds:    []shelf.Kind{shelf.Kind(strings.TrimSpace(kind))},
		Statuses: []shelf.Status{shelf.Status(strings.TrimSpace(status))},
		Limit:    limit,
	})
}

func runTriageAuto(rootDir string, candidates []shelf.Task, action string) (int, error) {
	resolvedStatus, doArchive, err := resolveTriageAction(action)
	if err != nil {
		return 0, err
	}
	updated := 0
	err = withWriteLock(rootDir, func() error {
		if err := prepareUndoSnapshot(rootDir, "triage-auto"); err != nil {
			return err
		}
		for _, task := range candidates {
			input := shelf.SetTaskInput{}
			if doArchive {
				now := time.Now().Local().Round(time.Second).Format(time.RFC3339)
				input.ArchivedAt = &now
			} else {
				status := resolvedStatus
				input.Status = &status
			}
			if _, err := shelf.SetTask(rootDir, task.ID, input); err != nil {
				return err
			}
			updated++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return updated, nil
}

func runTriageInteractive(ctx *commandContext, candidates []shelf.Task) error {
	edited := 0
	statusUpdated := 0
	archived := 0
	skipped := 0

	for i, candidate := range candidates {
		latest, err := shelf.EnsureTaskExists(ctx.rootDir, candidate.ID)
		if err != nil {
			return err
		}

		action, err := interactive.SelectWithConfig(interactive.SelectConfig{
			Prompt: fmt.Sprintf("Triage %d/%d: %s", i+1, len(candidates), latest.Title),
			Options: []interactive.Option{
				{Value: "edit", Label: "Edit fields", Preview: buildPreview(latest)},
				{Value: "done", Label: "Set status done", Preview: buildPreview(latest)},
				{Value: "start", Label: "Set status in_progress", Preview: buildPreview(latest)},
				{Value: "block", Label: "Set status blocked", Preview: buildPreview(latest)},
				{Value: "cancel", Label: "Set status cancelled", Preview: buildPreview(latest)},
				{Value: "reopen", Label: "Set status open", Preview: buildPreview(latest)},
				{Value: "archive", Label: "Archive task", Preview: buildPreview(latest)},
				{Value: "skip", Label: "Skip", Preview: buildPreview(latest)},
				{Value: "quit", Label: "Quit triage", Preview: buildPreview(latest)},
			},
			ShowPreview:       true,
			MaxRows:           12,
			HelpText:          selectorHelpText,
			SearchPlaceholder: "操作検索",
		})
		if err != nil {
			return err
		}

		switch action.Value {
		case "edit":
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			input, err := resolveSetInputInteractive(ctx, latest.ID, latest, cfg)
			if err != nil {
				if errors.Is(err, interactive.ErrCanceled) {
					skipped++
					continue
				}
				return err
			}
			if err := withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "triage-edit"); err != nil {
					return err
				}
				_, err := shelf.SetTask(ctx.rootDir, latest.ID, input)
				return err
			}); err != nil {
				return err
			}
			edited++
		case "archive":
			now := time.Now().Local().Round(time.Second).Format(time.RFC3339)
			if err := withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "triage-archive"); err != nil {
					return err
				}
				_, err := shelf.SetTask(ctx.rootDir, latest.ID, shelf.SetTaskInput{ArchivedAt: &now})
				return err
			}); err != nil {
				return err
			}
			archived++
		case "skip":
			skipped++
		case "quit":
			fmt.Printf("Triage summary: edited=%d status_updated=%d archived=%d skipped=%d\n", edited, statusUpdated, archived, skipped)
			return nil
		default:
			nextStatus, _, err := resolveTriageAction(action.Value)
			if err != nil {
				return err
			}
			if err := withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "triage-status"); err != nil {
					return err
				}
				_, err := shelf.SetTask(ctx.rootDir, latest.ID, shelf.SetTaskInput{Status: &nextStatus})
				return err
			}); err != nil {
				return err
			}
			statusUpdated++
		}
	}

	fmt.Printf("Triage summary: edited=%d status_updated=%d archived=%d skipped=%d\n", edited, statusUpdated, archived, skipped)
	return nil
}

func resolveTriageAction(action string) (shelf.Status, bool, error) {
	switch action {
	case "done":
		return "done", false, nil
	case "start":
		return "in_progress", false, nil
	case "block":
		return "blocked", false, nil
	case "cancel":
		return "cancelled", false, nil
	case "reopen":
		return "open", false, nil
	case "archive":
		return "", true, nil
	default:
		return "", false, fmt.Errorf("invalid --auto action: %s (allowed: done|start|block|cancel|reopen|archive)", action)
	}
}
