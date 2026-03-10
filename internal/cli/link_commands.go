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
	var (
		asJSON   bool
		format   string
		fields   string
		header   bool
		noHeader bool
	)
	cmd := &cobra.Command{
		Use:     "links <task-id>",
		Short:   "Show outbound and inbound links for a task",
		Args:    cobra.ExactArgs(1),
		Example: "  shelf links 01AAA\n  shelf links 01AAA --json\n  shelf links 01AAA --format tsv --fields direction,type,other_path\n  shelf links 01AAA --format csv",
		RunE: func(_ *cobra.Command, args []string) error {
			if err := validateFormat(format, []string{"compact", "tsv", "csv", "jsonl"}); err != nil {
				return err
			}
			if strings.TrimSpace(fields) != "" && format != "tsv" && format != "csv" {
				return fmt.Errorf("--fields requires --format tsv or csv")
			}
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
					File  string `json:"file,omitempty"`
					Title string `json:"title,omitempty"`
					Path  string `json:"path,omitempty"`
					Type  string `json:"type"`
				}
				payload := struct {
					TaskID   string      `json:"task_id"`
					Task     linkTaskRef `json:"task"`
					Outbound []edgeItem  `json:"outbound"`
					Inbound  []edgeItem  `json:"inbound"`
				}{
					TaskID:   taskID,
					Task:     buildLinkTaskRef(ctx.rootDir, taskID, byID),
					Outbound: make([]edgeItem, 0, len(outbound)),
					Inbound:  make([]edgeItem, 0, len(inbound)),
				}
				for _, edge := range outbound {
					record := buildEdgeQueryRecord(ctx.rootDir, "outbound", taskID, edge.To, edge.Type, byID)
					payload.Outbound = append(payload.Outbound, edgeItem{ID: record.Other.ID, File: record.Other.File, Title: record.Other.Title, Path: record.Other.Path, Type: record.Type})
				}
				for _, edge := range inbound {
					record := buildEdgeQueryRecord(ctx.rootDir, "inbound", edge.From, taskID, edge.Type, byID)
					payload.Inbound = append(payload.Inbound, edgeItem{ID: record.Other.ID, File: record.Other.File, Title: record.Other.Title, Path: record.Other.Path, Type: record.Type})
				}
				data, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			if format == "jsonl" {
				records := make([]edgeQueryRecord, 0, len(outbound)+len(inbound))
				for _, edge := range outbound {
					records = append(records, buildEdgeQueryRecord(ctx.rootDir, "outbound", taskID, edge.To, edge.Type, byID))
				}
				for _, edge := range inbound {
					records = append(records, buildEdgeQueryRecord(ctx.rootDir, "inbound", edge.From, taskID, edge.Type, byID))
				}
				text, err := renderJSONL(records)
				if err != nil {
					return err
				}
				fmt.Print(text)
				return nil
			}

			if format == "tsv" {
				selectedFields, err := resolveTSVFields(fields, defaultLinksTSVFields(), allowedLinksTSVFields())
				if err != nil {
					return err
				}
				includeHeader, err := resolveTabularHeader(format, header, noHeader)
				if err != nil {
					return err
				}
				if includeHeader {
					fmt.Println(strings.Join(selectedFields, "\t"))
				}
				for _, edge := range outbound {
					fmt.Println(joinTSVFields(selectedFields, buildEdgeQueryRecord(ctx.rootDir, "outbound", taskID, edge.To, edge.Type, byID).TSVFields()))
				}
				for _, edge := range inbound {
					fmt.Println(joinTSVFields(selectedFields, buildEdgeQueryRecord(ctx.rootDir, "inbound", edge.From, taskID, edge.Type, byID).TSVFields()))
				}
				return nil
			}

			if format == "csv" {
				selectedFields, err := resolveTSVFields(fields, defaultLinksTSVFields(), allowedLinksTSVFields())
				if err != nil {
					return err
				}
				includeHeader, err := resolveTabularHeader(format, header, noHeader)
				if err != nil {
					return err
				}
				records := make([]edgeQueryRecord, 0, len(outbound)+len(inbound))
				for _, edge := range outbound {
					records = append(records, buildEdgeQueryRecord(ctx.rootDir, "outbound", taskID, edge.To, edge.Type, byID))
				}
				for _, edge := range inbound {
					records = append(records, buildEdgeQueryRecord(ctx.rootDir, "inbound", edge.From, taskID, edge.Type, byID))
				}
				text, err := renderCSV(records, selectedFields, includeHeader)
				if err != nil {
					return err
				}
				fmt.Print(text)
				return nil
			}

			printLinkSection("Outbound", taskID, outbound, byID, ctx.showID)
			printInboundLinkSection("Inbound", taskID, inbound, byID, ctx.showID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&format, "format", "compact", "Output format: compact|tsv|csv|jsonl")
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated field names for --format tsv or csv")
	cmd.Flags().BoolVar(&header, "header", false, "Include a header row for tabular output")
	cmd.Flags().BoolVar(&noHeader, "no-header", false, "Omit the header row for tabular output")
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

func defaultLinksTSVFields() []string {
	return []string{"direction", "type", "task_id", "task_path", "other_id", "other_path", "other_file"}
}

func allowedLinksTSVFields() map[string]struct{} {
	return map[string]struct{}{
		"direction": {}, "type": {}, "task_id": {}, "task_title": {}, "task_path": {}, "task_file": {},
		"other_id": {}, "other_title": {}, "other_path": {}, "other_file": {},
	}
}
