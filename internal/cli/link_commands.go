package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newLinkCommand(ctx *commandContext) *cobra.Command {
	var (
		from     string
		to       string
		linkType string
	)
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Create a task link",
		Example: "  shelf link --from 01AAA --to 01BBB --type depends_on\n" +
			"  shelf link --from 01AAA --to 01CCC --type related",
		RunE: func(_ *cobra.Command, _ []string) error {
			resolvedType, err := resolveCLIBlockingLinkType(ctx.rootDir, linkType)
			if err != nil {
				return err
			}
			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "cli-link"); err != nil {
					return err
				}
				if err := shelf.LinkTasks(ctx.rootDir, from, to, resolvedType); err != nil {
					return err
				}
				fmt.Printf("Linked %s --%s--> %s\n", from, resolvedType, to)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "Source task ID")
	cmd.Flags().StringVar(&to, "to", "", "Target task ID")
	cmd.Flags().StringVar(&linkType, "type", "", "Link type name from config (defaults to blocking type)")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")
	return cmd
}

func newUnlinkCommand(ctx *commandContext) *cobra.Command {
	var (
		from     string
		to       string
		linkType string
	)
	cmd := &cobra.Command{
		Use:   "unlink",
		Short: "Remove a task link",
		Example: "  shelf unlink --from 01AAA --to 01BBB --type depends_on\n" +
			"  shelf unlink --from 01AAA --to 01CCC --type related",
		RunE: func(_ *cobra.Command, _ []string) error {
			resolvedType, err := resolveCLIBlockingLinkType(ctx.rootDir, linkType)
			if err != nil {
				return err
			}
			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "cli-unlink"); err != nil {
					return err
				}
				removed, err := shelf.UnlinkTasks(ctx.rootDir, from, to, resolvedType)
				if err != nil {
					return err
				}
				if !removed {
					return fmt.Errorf("link not found: %s --%s--> %s", from, resolvedType, to)
				}
				fmt.Printf("Removed %s --%s--> %s\n", from, resolvedType, to)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "Source task ID")
	cmd.Flags().StringVar(&to, "to", "", "Target task ID")
	cmd.Flags().StringVar(&linkType, "type", "", "Link type name from config (defaults to blocking type)")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")
	return cmd
}

func newLinksCommand(ctx *commandContext) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "links <task-id>",
		Short:   "Show outbound and inbound links for a task",
		Args:    cobra.ExactArgs(1),
		Example: "  shelf links 01AAA\n  shelf links 01AAA --json",
		RunE: func(_ *cobra.Command, args []string) error {
			taskID := strings.TrimSpace(args[0])
			outbound, inbound, err := shelf.ListLinks(ctx.rootDir, taskID)
			if err != nil {
				return err
			}
			tasks, err := shelf.NewTaskStore(ctx.rootDir).List()
			if err != nil {
				return err
			}
			byID := make(map[string]shelf.Task, len(tasks))
			for _, task := range tasks {
				byID[task.ID] = task
			}

			if asJSON {
				type edgeItem struct {
					ID    string `json:"id"`
					Title string `json:"title,omitempty"`
					Type  string `json:"type"`
				}
				payload := struct {
					TaskID   string     `json:"task_id"`
					Outbound []edgeItem `json:"outbound"`
					Inbound  []edgeItem `json:"inbound"`
				}{
					TaskID:   taskID,
					Outbound: make([]edgeItem, 0, len(outbound)),
					Inbound:  make([]edgeItem, 0, len(inbound)),
				}
				for _, edge := range outbound {
					payload.Outbound = append(payload.Outbound, edgeItem{ID: edge.To, Title: byID[edge.To].Title, Type: string(edge.Type)})
				}
				for _, edge := range inbound {
					payload.Inbound = append(payload.Inbound, edgeItem{ID: edge.From, Title: byID[edge.From].Title, Type: string(edge.Type)})
				}
				data, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			printLinkSection("Outbound", taskID, outbound, byID, ctx.showID)
			printInboundLinkSection("Inbound", taskID, inbound, byID, ctx.showID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func resolveCLIBlockingLinkType(rootDir, flagValue string) (shelf.LinkType, error) {
	if value := shelf.LinkType(strings.TrimSpace(flagValue)); value != "" {
		return value, nil
	}
	cfg, err := shelf.LoadConfig(rootDir)
	if err != nil {
		return "", err
	}
	return cfg.BlockingLinkType(), nil
}

func printLinkSection(title, taskID string, outbound []shelf.Edge, byID map[string]shelf.Task, showID bool) {
	fmt.Println(uiHeading(title + ":"))
	if len(outbound) == 0 {
		fmt.Println(uiMuted("  (none)"))
		return
	}
	source := formatLinkEndpoint(taskID, byID, showID)
	for _, edge := range outbound {
		target := formatLinkEndpoint(edge.To, byID, showID)
		fmt.Printf("  %s --%s--> %s\n", source, edge.Type, target)
	}
}

func printInboundLinkSection(title, taskID string, inbound []shelf.InboundEdge, byID map[string]shelf.Task, showID bool) {
	fmt.Println(uiHeading(title + ":"))
	if len(inbound) == 0 {
		fmt.Println(uiMuted("  (none)"))
		return
	}
	target := formatLinkEndpoint(taskID, byID, showID)
	for _, edge := range inbound {
		source := formatLinkEndpoint(edge.From, byID, showID)
		fmt.Printf("  %s --%s--> %s\n", source, edge.Type, target)
	}
}

func formatLinkEndpoint(taskID string, byID map[string]shelf.Task, showID bool) string {
	task, ok := byID[taskID]
	if !ok || strings.TrimSpace(task.Title) == "" {
		return taskID
	}
	return formatTaskPathLabel(task, byID, showID)
}
