package cli

import (
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newLsCommand(ctx *commandContext) *cobra.Command {
	var (
		kind   string
		state  string
		parent string
		limit  int
		search string
	)

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List tasks",
		RunE: func(_ *cobra.Command, _ []string) error {
			tasks, err := shelf.ListTasks(ctx.rootDir, shelf.TaskFilter{
				Kind:   shelf.Kind(kind),
				State:  shelf.State(state),
				Parent: parent,
				Limit:  limit,
				Search: search,
			})
			if err != nil {
				return err
			}
			for _, task := range tasks {
				parentLabel := "root"
				if task.Parent != "" {
					parentLabel = shelf.ShortID(task.Parent)
				}
				fmt.Printf("[%s] %s  (%s/%s) parent=%s\n", shelf.ShortID(task.ID), task.Title, task.Kind, task.State, parentLabel)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&kind, "kind", "", "Filter by kind")
	cmd.Flags().StringVar(&state, "state", "", "Filter by state")
	cmd.Flags().StringVar(&parent, "parent", "", "Filter by parent task ID or root")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of items")
	cmd.Flags().StringVar(&search, "search", "", "Search by title/body")
	return cmd
}

func newShowCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show task details",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			task, err := shelf.EnsureTaskExists(ctx.rootDir, args[0])
			if err != nil {
				return err
			}

			fmt.Println("+++")
			fmt.Printf("id = %q\n", task.ID)
			fmt.Printf("title = %q\n", task.Title)
			fmt.Printf("kind = %q\n", task.Kind)
			fmt.Printf("state = %q\n", task.State)
			if task.Parent != "" {
				fmt.Printf("parent = %q\n", task.Parent)
			}
			fmt.Printf("created_at = %q\n", task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
			fmt.Printf("updated_at = %q\n", task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
			fmt.Println("+++")
			fmt.Println()
			fmt.Println(task.Body)
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
		state    string
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
				State:    shelf.State(state),
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
	cmd.Flags().StringVar(&state, "state", "", "Filter by state")
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

	fmt.Printf("%s%s[%s] %s (%s/%s)\n", prefix, branch, shelf.ShortID(node.Task.ID), node.Task.Title, node.Task.Kind, node.Task.State)
	for i, child := range node.Children {
		printTreeNode(child, nextPrefix, i == len(node.Children)-1)
	}
}
