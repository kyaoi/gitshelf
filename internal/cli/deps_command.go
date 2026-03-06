package cli

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newDepsCommand(ctx *commandContext) *cobra.Command {
	var (
		transitive bool
		reverse    bool
		asJSON     bool
	)
	cmd := &cobra.Command{
		Use:   "deps <id>",
		Short: "Show prerequisites and dependents by depends_on links",
		Example: "  shelf deps <id>\n" +
			"  shelf deps <id> --transitive\n" +
			"  shelf deps <id> --reverse --json",
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "依存関係を表示するタスクを選択", nil, true)
			if err != nil {
				return err
			}
			taskStore := shelf.NewTaskStore(ctx.rootDir)
			tasks, err := taskStore.List()
			if err != nil {
				return err
			}
			titleByID := make(map[string]string, len(tasks))
			for _, task := range tasks {
				titleByID[task.ID] = task.Title
			}

			prerequisites, err := listPrerequisites(ctx.rootDir, id, transitive)
			if err != nil {
				return err
			}
			dependents, err := listDependents(ctx.rootDir, id, transitive, tasks)
			if err != nil {
				return err
			}

			if asJSON {
				payload := map[string]any{
					"id":            id,
					"transitive":    transitive,
					"prerequisites": prerequisites,
					"dependents":    dependents,
				}
				data, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			printList := func(label string, ids []string) {
				fmt.Println(uiHeading(label + ":"))
				if len(ids) == 0 {
					fmt.Println(uiMuted("  (none)"))
					return
				}
				for _, depID := range ids {
					title := titleByID[depID]
					item := title
					if title == "" {
						item = "(missing)"
					}
					if ctx.showID {
						item = fmt.Sprintf("[%s] %s", shelf.ShortID(depID), item)
					}
					fmt.Printf("  - %s\n", item)
				}
			}

			if reverse {
				printList("Dependents", dependents)
				printList("Prerequisites", prerequisites)
			} else {
				printList("Prerequisites", prerequisites)
				printList("Dependents", dependents)
			}
			fmt.Println("depends_on の向き: A depends_on B = AをやるにはBが先")
			return nil
		},
	}
	cmd.Flags().BoolVar(&transitive, "transitive", false, "Show transitive closure")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "Print dependents first")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func listPrerequisites(rootDir string, id string, transitive bool) ([]string, error) {
	if transitive {
		return shelf.ListTransitiveDependencies(rootDir, id)
	}
	edges, err := shelf.NewEdgeStore(rootDir).ListOutbound(id)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(edges))
	for _, edge := range edges {
		if edge.Type != "depends_on" {
			continue
		}
		result = append(result, edge.To)
	}
	slices.Sort(result)
	return result, nil
}

func listDependents(rootDir string, id string, transitive bool, tasks []shelf.Task) ([]string, error) {
	if !transitive {
		inbound, err := shelf.NewEdgeStore(rootDir).FindInbound(id)
		if err != nil {
			return nil, err
		}
		result := make([]string, 0, len(inbound))
		for _, edge := range inbound {
			if edge.Type != "depends_on" {
				continue
			}
			result = append(result, edge.From)
		}
		slices.Sort(result)
		return result, nil
	}

	result := make([]string, 0)
	for _, task := range tasks {
		if task.ID == id {
			continue
		}
		deps, err := shelf.ListTransitiveDependencies(rootDir, task.ID)
		if err != nil {
			return nil, err
		}
		if slices.Contains(deps, id) {
			result = append(result, task.ID)
		}
	}
	slices.Sort(result)
	return result, nil
}
