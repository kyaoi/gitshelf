package cli

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newDepsCommand(ctx *commandContext) *cobra.Command {
	var (
		transitive bool
		reverse    bool
		graph      bool
		asJSON     bool
	)
	cmd := &cobra.Command{
		Use:   "deps <id>",
		Short: "Show prerequisites and dependents by depends_on links",
		Example: "  shelf deps <id>\n" +
			"  shelf deps <id> --transitive\n" +
			"  shelf deps <id> --graph --transitive\n" +
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
				if graph {
					titleByIDPlain := make(map[string]string, len(titleByID))
					for idKey, title := range titleByID {
						titleByIDPlain[idKey] = title
					}
					maxDepth := 1
					if transitive {
						maxDepth = 0
					}
					prereqAdj, dependAdj, err := buildDependsAdjacency(ctx.rootDir, tasks)
					if err != nil {
						return err
					}
					payload["prerequisites_graph"] = renderDepGraphLines(id, prereqAdj, titleByIDPlain, ctx.showID, maxDepth)
					payload["dependents_graph"] = renderDepGraphLines(id, dependAdj, titleByIDPlain, ctx.showID, maxDepth)
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
				if graph {
					if err := printDepGraph(ctx.rootDir, "Dependents", id, tasks, titleByID, ctx.showID, transitive, false); err != nil {
						return err
					}
					if err := printDepGraph(ctx.rootDir, "Prerequisites", id, tasks, titleByID, ctx.showID, transitive, true); err != nil {
						return err
					}
				} else {
					printList("Dependents", dependents)
					printList("Prerequisites", prerequisites)
				}
			} else {
				if graph {
					if err := printDepGraph(ctx.rootDir, "Prerequisites", id, tasks, titleByID, ctx.showID, transitive, true); err != nil {
						return err
					}
					if err := printDepGraph(ctx.rootDir, "Dependents", id, tasks, titleByID, ctx.showID, transitive, false); err != nil {
						return err
					}
				} else {
					printList("Prerequisites", prerequisites)
					printList("Dependents", dependents)
				}
			}
			fmt.Println("depends_on の向き: A depends_on B = AをやるにはBが先")
			return nil
		},
	}
	cmd.Flags().BoolVar(&transitive, "transitive", false, "Show transitive closure")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "Print dependents first")
	cmd.Flags().BoolVar(&graph, "graph", false, "Render dependency graph")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func printDepGraph(rootDir string, heading string, taskID string, tasks []shelf.Task, titleByID map[string]string, showID bool, transitive bool, prereqDirection bool) error {
	maxDepth := 1
	if transitive {
		maxDepth = 0
	}
	prereqAdj, dependAdj, err := buildDependsAdjacency(rootDir, tasks)
	if err != nil {
		return err
	}
	adj := dependAdj
	if prereqDirection {
		adj = prereqAdj
	}
	titleByIDPlain := make(map[string]string, len(titleByID))
	for id, title := range titleByID {
		titleByIDPlain[id] = title
	}
	fmt.Println(uiHeading(heading + " Graph:"))
	lines := renderDepGraphLines(taskID, adj, titleByIDPlain, showID, maxDepth)
	for _, line := range lines {
		fmt.Println(line)
	}
	return nil
}

func buildDependsAdjacency(rootDir string, tasks []shelf.Task) (map[string][]string, map[string][]string, error) {
	prereqAdj := make(map[string][]string, len(tasks))
	dependAdj := make(map[string][]string, len(tasks))
	for _, task := range tasks {
		prereqAdj[task.ID] = []string{}
		dependAdj[task.ID] = []string{}
	}
	edgeStore := shelf.NewEdgeStore(rootDir)
	for _, task := range tasks {
		edges, err := edgeStore.ListOutbound(task.ID)
		if err != nil {
			return nil, nil, err
		}
		for _, edge := range edges {
			if edge.Type != "depends_on" {
				continue
			}
			prereqAdj[task.ID] = append(prereqAdj[task.ID], edge.To)
			dependAdj[edge.To] = append(dependAdj[edge.To], task.ID)
		}
	}
	for id := range prereqAdj {
		slices.Sort(prereqAdj[id])
		prereqAdj[id] = slices.Compact(prereqAdj[id])
	}
	for id := range dependAdj {
		slices.Sort(dependAdj[id])
		dependAdj[id] = slices.Compact(dependAdj[id])
	}
	return prereqAdj, dependAdj, nil
}

func renderDepGraphLines(rootID string, adjacency map[string][]string, titleByID map[string]string, showID bool, maxDepth int) []string {
	lines := []string{depGraphLabel(rootID, titleByID, showID)}
	children := adjacency[rootID]
	if len(children) == 0 {
		return append(lines, uiMuted("  (none)"))
	}
	visited := map[string]bool{rootID: true}
	var visit func(parent string, prefix string, depth int)
	visit = func(parent string, prefix string, depth int) {
		children := adjacency[parent]
		for i, child := range children {
			isLast := i == len(children)-1
			branch := "├─ "
			nextPrefix := prefix + "│  "
			if isLast {
				branch = "└─ "
				nextPrefix = prefix + "   "
			}

			label := depGraphLabel(child, titleByID, showID)
			if visited[child] {
				lines = append(lines, prefix+branch+label+uiMuted(" (cycle)"))
				continue
			}
			lines = append(lines, prefix+branch+label)
			if maxDepth > 0 && depth >= maxDepth {
				continue
			}
			visited[child] = true
			visit(child, nextPrefix, depth+1)
			delete(visited, child)
		}
	}
	visit(rootID, "", 1)
	return lines
}

func depGraphLabel(id string, titleByID map[string]string, showID bool) string {
	title := strings.TrimSpace(titleByID[id])
	if title == "" {
		if showID {
			return uiShortID(shelf.ShortID(id))
		}
		return uiMuted("(missing)")
	}
	label := uiPrimary(title)
	if showID {
		return fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(id)), label)
	}
	return label
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
