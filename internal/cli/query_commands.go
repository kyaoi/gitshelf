package cli

import (
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newLsCommand(ctx *commandContext) *cobra.Command {
	var (
		kinds       []string
		statuses    []string
		notKinds    []string
		notStatuses []string
		parent      string
		limit       int
		search      string
	)

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List tasks",
		RunE: func(_ *cobra.Command, _ []string) error {
			tasks, err := shelf.ListTasks(ctx.rootDir, shelf.TaskFilter{
				Kinds:       toKinds(kinds),
				Statuses:    toStatuses(statuses),
				NotKinds:    toKinds(notKinds),
				NotStatuses: toStatuses(notStatuses),
				Parent:      parent,
				Limit:       limit,
				Search:      search,
			})
			if err != nil {
				return err
			}
			allTasks, err := shelf.NewTaskStore(ctx.rootDir).List()
			if err != nil {
				return err
			}
			titleByID := make(map[string]string, len(allTasks))
			for _, task := range allTasks {
				titleByID[task.ID] = task.Title
			}
			for _, task := range tasks {
				parentLabel := "root"
				if task.Parent != "" {
					if title, ok := titleByID[task.Parent]; ok {
						parentLabel = title
					} else {
						parentLabel = "(missing)"
					}
				}
				fmt.Printf("%s  (%s/%s) parent=%s\n", task.Title, task.Kind, task.Status, parentLabel)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&kinds, "kind", nil, "Include kind (repeatable)")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Include status (repeatable)")
	cmd.Flags().StringArrayVar(&notKinds, "not-kind", nil, "Exclude kind (repeatable)")
	cmd.Flags().StringArrayVar(&notStatuses, "not-status", nil, "Exclude status (repeatable)")
	cmd.Flags().StringVar(&parent, "parent", "", "Filter by parent task ID or root")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of items")
	cmd.Flags().StringVar(&search, "search", "", "Search by title/body")
	return cmd
}

func newShowCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show task details",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "表示するタスクを選択", nil)
			if err != nil {
				return err
			}

			task, err := shelf.EnsureTaskExists(ctx.rootDir, id)
			if err != nil {
				return err
			}

			fmt.Println("+++")
			fmt.Printf("id = %q\n", task.ID)
			fmt.Printf("title = %q\n", task.Title)
			fmt.Printf("kind = %q\n", task.Kind)
			fmt.Printf("status = %q\n", task.Status)
			if task.Parent != "" {
				fmt.Printf("parent = %q\n", task.Parent)
			}
			fmt.Printf("created_at = %q\n", task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
			fmt.Printf("updated_at = %q\n", task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
			fmt.Println("+++")
			fmt.Println()
			fmt.Println(task.Body)
			fmt.Println()

			allTasks, err := shelf.NewTaskStore(ctx.rootDir).List()
			if err != nil {
				return err
			}
			byID := make(map[string]shelf.Task, len(allTasks))
			for _, item := range allTasks {
				byID[item.ID] = item
			}
			fmt.Println("Hierarchy:")
			fmt.Printf("Path: %s\n", buildTaskPath(task, byID))
			subtree, err := shelf.BuildTree(ctx.rootDir, shelf.TreeOptions{FromID: task.ID})
			if err != nil {
				return err
			}
			fmt.Println("Subtree:")
			if len(subtree) == 0 {
				fmt.Println("  (none)")
			} else {
				for i, node := range subtree {
					printTreeNode(node, "", i == len(subtree)-1)
				}
			}
			fmt.Println()

			edgeStore := shelf.NewEdgeStore(ctx.rootDir)
			outbound, err := edgeStore.ListOutbound(task.ID)
			if err != nil {
				return err
			}
			inbound, err := edgeStore.FindInbound(task.ID)
			if err != nil {
				return err
			}

			fmt.Println("Outbound Links:")
			if len(outbound) == 0 {
				fmt.Println("  (none)")
			} else {
				for _, edge := range outbound {
					fmt.Printf("  [%s] --%s--> [%s]\n", shelf.ShortID(task.ID), edge.Type, shelf.ShortID(edge.To))
				}
			}
			fmt.Println("Inbound Links:")
			if len(inbound) == 0 {
				fmt.Println("  (none)")
			} else {
				for _, edge := range inbound {
					fmt.Printf("  [%s] --%s--> [%s]\n", shelf.ShortID(edge.From), edge.Type, shelf.ShortID(task.ID))
				}
			}
			return nil
		},
	}
	return cmd
}

func newTreeCommand(ctx *commandContext) *cobra.Command {
	var (
		from     string
		maxDepth int
		status   string
	)

	cmd := &cobra.Command{
		Use:   "tree",
		Short: "Show task tree",
		RunE: func(_ *cobra.Command, _ []string) error {
			fromID := ""
			if strings.TrimSpace(from) != "" && !strings.EqualFold(from, "root") {
				fromID = from
			}
			nodes, err := shelf.BuildTree(ctx.rootDir, shelf.TreeOptions{
				FromID:   fromID,
				Status:   shelf.Status(status),
				MaxDepth: maxDepth,
			})
			if err != nil {
				return err
			}
			for i, node := range nodes {
				printTreeNode(node, "", i == len(nodes)-1)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "root", "Start from task ID or root")
	cmd.Flags().IntVar(&maxDepth, "max-depth", 0, "Maximum depth (0 means unlimited)")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	return cmd
}

func printTreeNode(node shelf.TreeNode, prefix string, isLast bool) {
	branch := "├─ "
	nextPrefix := prefix + "│  "
	if isLast {
		branch = "└─ "
		nextPrefix = prefix + "   "
	}
	if prefix == "" {
		branch = ""
	}

	fmt.Printf("%s%s%s (%s/%s)\n", prefix, branch, node.Task.Title, node.Task.Kind, node.Task.Status)
	for i, child := range node.Children {
		printTreeNode(child, nextPrefix, i == len(node.Children)-1)
	}
}

func toKinds(values []string) []shelf.Kind {
	kinds := make([]shelf.Kind, len(values))
	for i, value := range values {
		kinds[i] = shelf.Kind(value)
	}
	return kinds
}

func toStatuses(values []string) []shelf.Status {
	statuses := make([]shelf.Status, len(values))
	for i, value := range values {
		statuses[i] = shelf.Status(value)
	}
	return statuses
}

func buildTaskPath(task shelf.Task, byID map[string]shelf.Task) string {
	titles := []string{task.Title}
	current := task.Parent
	seen := map[string]struct{}{}
	for current != "" {
		if _, ok := seen[current]; ok {
			titles = append([]string{"(cycle)"}, titles...)
			break
		}
		seen[current] = struct{}{}

		parent, ok := byID[current]
		if !ok {
			titles = append([]string{"(missing)"}, titles...)
			break
		}
		titles = append([]string{parent.Title}, titles...)
		current = parent.Parent
	}
	return "root > " + strings.Join(titles, " > ")
}
