package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newShowCommand(ctx *commandContext) *cobra.Command {
	var (
		asJSON   bool
		format   string
		fields   string
		header   bool
		noHeader bool
	)

	cmd := &cobra.Command{
		Use:     "show <task-id>",
		Short:   "Show one task with inspector-style details",
		Args:    cobra.ExactArgs(1),
		Example: "  shelf show 01AAA\n  shelf show 01AAA --json\n  shelf show 01AAA --format tsv --fields id,title,file\n  shelf show 01AAA --format csv",
		RunE: func(_ *cobra.Command, args []string) error {
			if err := validateFormat(format, []string{"compact", "tsv", "csv", "jsonl"}); err != nil {
				return err
			}
			if strings.TrimSpace(fields) != "" && format != "tsv" && format != "csv" {
				return fmt.Errorf("--fields requires --format tsv or csv")
			}
			taskID := strings.TrimSpace(args[0])
			store := shelf.NewTaskStore(ctx.rootDir)
			task, err := store.Get(taskID)
			if err != nil {
				return err
			}
			tasks, err := store.List()
			if err != nil {
				return err
			}
			byID := make(map[string]shelf.Task, len(tasks))
			for _, candidate := range tasks {
				byID[candidate.ID] = candidate
			}
			outbound, inbound, err := shelf.ListLinks(ctx.rootDir, taskID)
			if err != nil {
				return err
			}

			if asJSON {
				data, err := json.MarshalIndent(buildShowTaskPayload(ctx.rootDir, task, byID, outbound, inbound), "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			record := buildTaskQueryRecord(ctx.rootDir, task, byID)

			if format == "jsonl" {
				text, err := renderJSONL([]taskQueryRecord{record})
				if err != nil {
					return err
				}
				fmt.Print(text)
				return nil
			}

			if format == "tsv" {
				selectedFields, err := resolveTSVFields(fields, defaultShowTSVFields(), allowedShowTSVFields())
				if err != nil {
					return err
				}
				includeHeader, err := resolveTabularHeader(format, header, noHeader)
				if err != nil {
					return err
				}
				row := record.TSVFields()
				row["outbound_count"] = fmt.Sprintf("%d", len(outbound))
				row["inbound_count"] = fmt.Sprintf("%d", len(inbound))
				if includeHeader {
					fmt.Println(strings.Join(selectedFields, "\t"))
				}
				fmt.Println(joinTSVFields(selectedFields, row))
				return nil
			}

			if format == "csv" {
				selectedFields, err := resolveTSVFields(fields, defaultShowCSVFields(), allowedShowTSVFields())
				if err != nil {
					return err
				}
				includeHeader, err := resolveTabularHeader(format, header, noHeader)
				if err != nil {
					return err
				}
				text, err := renderCSV([]taskQueryRecord{record}, selectedFields, includeHeader)
				if err != nil {
					return err
				}
				fmt.Print(text)
				return nil
			}

			printTaskDetails(task, byID, outbound, inbound, ctx.showID)
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

type showTaskPayload struct {
	ID          string            `json:"id"`
	File        string            `json:"file"`
	Title       string            `json:"title"`
	Path        string            `json:"path"`
	Kind        string            `json:"kind"`
	Status      string            `json:"status"`
	Tags        []string          `json:"tags,omitempty"`
	DueOn       string            `json:"due_on,omitempty"`
	RepeatEvery string            `json:"repeat_every,omitempty"`
	ArchivedAt  string            `json:"archived_at,omitempty"`
	Parent      string            `json:"parent,omitempty"`
	ParentTitle string            `json:"parent_title,omitempty"`
	ParentPath  string            `json:"parent_path,omitempty"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	Body        string            `json:"body,omitempty"`
	Outbound    []showLinkPayload `json:"outbound"`
	Inbound     []showLinkPayload `json:"inbound"`
}

type showLinkPayload struct {
	ID    string `json:"id"`
	File  string `json:"file,omitempty"`
	Title string `json:"title,omitempty"`
	Path  string `json:"path,omitempty"`
	Type  string `json:"type"`
}

func buildShowTaskPayload(rootDir string, task shelf.Task, byID map[string]shelf.Task, outbound []shelf.Edge, inbound []shelf.InboundEdge) showTaskPayload {
	record := buildTaskQueryRecord(rootDir, task, byID)
	payload := showTaskPayload{
		ID:          record.ID,
		File:        record.File,
		Title:       record.Title,
		Path:        record.Path,
		Kind:        record.Kind,
		Status:      record.Status,
		Tags:        record.Tags,
		DueOn:       record.DueOn,
		RepeatEvery: record.RepeatEvery,
		ArchivedAt:  record.ArchivedAt,
		Parent:      record.Parent,
		ParentTitle: record.ParentTitle,
		ParentPath:  record.ParentPath,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
		Body:        record.Body,
		Outbound:    make([]showLinkPayload, 0, len(outbound)),
		Inbound:     make([]showLinkPayload, 0, len(inbound)),
	}
	for _, edge := range outbound {
		payload.Outbound = append(payload.Outbound, buildShowLinkPayload(rootDir, edge.To, edge.Type, byID))
	}
	for _, edge := range inbound {
		payload.Inbound = append(payload.Inbound, buildShowLinkPayload(rootDir, edge.From, edge.Type, byID))
	}
	return payload
}

func buildShowLinkPayload(rootDir, taskID string, linkType shelf.LinkType, byID map[string]shelf.Task) showLinkPayload {
	payload := showLinkPayload{
		ID:   taskID,
		Type: string(linkType),
	}
	if task, ok := byID[taskID]; ok {
		payload.File = taskFilePath(rootDir, task.ID)
		payload.Title = task.Title
		payload.Path = buildTaskPath(task, byID)
	}
	return payload
}

func printTaskDetails(task shelf.Task, byID map[string]shelf.Task, outbound []shelf.Edge, inbound []shelf.InboundEdge, showID bool) {
	fmt.Printf("Task: %s\n", formatTaskPathLabel(task, byID, showID))
	fmt.Printf("Title: %s\n", task.Title)
	fmt.Printf("ID: %s\n", task.ID)
	fmt.Printf("Kind: %s\n", task.Kind)
	fmt.Printf("Status: %s\n", task.Status)
	fmt.Printf("Tags: %s\n", formatTagSummary(task.Tags))
	fmt.Printf("Due: %s\n", blankAsDash(task.DueOn))
	fmt.Printf("Repeat: %s\n", blankAsDash(task.RepeatEvery))
	fmt.Printf("Archived: %s\n", blankAsDash(task.ArchivedAt))
	if task.Parent == "" {
		fmt.Println("Parent: root")
	} else if parent, ok := byID[task.Parent]; ok {
		fmt.Printf("Parent: %s\n", formatTaskPathLabel(parent, byID, showID))
	} else {
		fmt.Printf("Parent: %s\n", task.Parent)
	}
	fmt.Printf("Created: %s\n", task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	fmt.Printf("Updated: %s\n", task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	fmt.Println("Body:")
	if strings.TrimSpace(task.Body) == "" {
		fmt.Println("  (empty)")
	} else {
		for _, line := range strings.Split(task.Body, "\n") {
			fmt.Printf("  %s\n", line)
		}
	}
	printLinkSection("Outbound", task.ID, outbound, byID, showID)
	printInboundLinkSection("Inbound", task.ID, inbound, byID, showID)
}

func blankAsDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func defaultShowTSVFields() []string {
	return []string{"id", "title", "path", "kind", "status", "due_on", "repeat_every", "parent", "parent_path", "tags", "file", "body"}
}

func defaultShowCSVFields() []string {
	return []string{"id", "title", "path", "kind", "status", "due_on", "repeat_every", "parent", "parent_path", "tags", "file", "body"}
}

func allowedShowTSVFields() map[string]struct{} {
	return map[string]struct{}{
		"id": {}, "title": {}, "path": {}, "kind": {}, "status": {}, "tags": {}, "due_on": {},
		"repeat_every": {}, "archived_at": {}, "parent": {}, "parent_path": {}, "file": {},
		"created_at": {}, "updated_at": {}, "body": {}, "outbound_count": {}, "inbound_count": {},
	}
}
