package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newShowCommand(ctx *commandContext) *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:     "show <task-id>",
		Short:   "Show one task with inspector-style details",
		Args:    cobra.ExactArgs(1),
		Example: "  shelf show 01AAA\n  shelf show 01AAA --json",
		RunE: func(_ *cobra.Command, args []string) error {
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
				data, err := json.MarshalIndent(buildShowTaskPayload(task, byID, outbound, inbound), "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			printTaskDetails(task, byID, outbound, inbound, ctx.showID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

type showTaskPayload struct {
	ID          string            `json:"id"`
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
	Title string `json:"title,omitempty"`
	Path  string `json:"path,omitempty"`
	Type  string `json:"type"`
}

func buildShowTaskPayload(task shelf.Task, byID map[string]shelf.Task, outbound []shelf.Edge, inbound []shelf.InboundEdge) showTaskPayload {
	payload := showTaskPayload{
		ID:          task.ID,
		Title:       task.Title,
		Path:        buildTaskPath(task, byID),
		Kind:        string(task.Kind),
		Status:      string(task.Status),
		Tags:        append([]string{}, task.Tags...),
		DueOn:       task.DueOn,
		RepeatEvery: task.RepeatEvery,
		ArchivedAt:  task.ArchivedAt,
		Parent:      task.Parent,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Body:        task.Body,
		Outbound:    make([]showLinkPayload, 0, len(outbound)),
		Inbound:     make([]showLinkPayload, 0, len(inbound)),
	}
	if task.Parent != "" {
		if parent, ok := byID[task.Parent]; ok {
			payload.ParentTitle = parent.Title
			payload.ParentPath = buildTaskPath(parent, byID)
		}
	}
	for _, edge := range outbound {
		payload.Outbound = append(payload.Outbound, buildShowLinkPayload(edge.To, edge.Type, byID))
	}
	for _, edge := range inbound {
		payload.Inbound = append(payload.Inbound, buildShowLinkPayload(edge.From, edge.Type, byID))
	}
	return payload
}

func buildShowLinkPayload(taskID string, linkType shelf.LinkType, byID map[string]shelf.Task) showLinkPayload {
	payload := showLinkPayload{
		ID:   taskID,
		Type: string(linkType),
	}
	if task, ok := byID[taskID]; ok {
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
