package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newLsCommand(ctx *commandContext) *cobra.Command {
	var (
		kinds       []string
		statuses    []string
		notKinds    []string
		notStatuses []string
		ready       bool
		depsBlocked bool
		dueBefore   string
		dueAfter    string
		overdue     bool
		noDue       bool
		asJSON      bool
		parent      string
		limit       int
		search      string
	)

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List tasks",
		Example: "  shelf ls\n" +
			"  shelf ls --kind todo --status open --status in_progress\n" +
			"  shelf ls --ready --overdue\n" +
			"  shelf ls --json",
		RunE: func(_ *cobra.Command, _ []string) error {
			tasks, err := shelf.ListTasks(ctx.rootDir, shelf.TaskFilter{
				Kinds:       toKinds(kinds),
				Statuses:    toStatuses(statuses),
				NotKinds:    toKinds(notKinds),
				NotStatuses: toStatuses(notStatuses),
				ReadyOnly:   ready,
				DepsBlocked: depsBlocked,
				DueBefore:   dueBefore,
				DueAfter:    dueAfter,
				Overdue:     overdue,
				NoDue:       noDue,
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

			if asJSON {
				type lsItem struct {
					ID          string `json:"id"`
					Title       string `json:"title"`
					Kind        string `json:"kind"`
					Status      string `json:"status"`
					DueOn       string `json:"due_on,omitempty"`
					Parent      string `json:"parent,omitempty"`
					ParentTitle string `json:"parent_title,omitempty"`
				}
				items := make([]lsItem, 0, len(tasks))
				for _, task := range tasks {
					parentTitle := ""
					if task.Parent != "" {
						parentTitle = titleByID[task.Parent]
					}
					items = append(items, lsItem{
						ID:          task.ID,
						Title:       task.Title,
						Kind:        string(task.Kind),
						Status:      string(task.Status),
						DueOn:       task.DueOn,
						Parent:      task.Parent,
						ParentTitle: parentTitle,
					})
				}
				data, err := json.MarshalIndent(items, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			for _, task := range tasks {
				parentLabel := "root"
				if task.Parent != "" {
					if title, ok := titleByID[task.Parent]; ok {
						parentLabel = uiPrimary(title)
						if ctx.showID {
							parentLabel = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(task.Parent)), uiPrimary(title))
						}
					} else {
						parentLabel = uiMuted("(missing)")
					}
				} else {
					parentLabel = uiMuted(parentLabel)
				}
				label := uiPrimary(task.Title)
				if ctx.showID {
					label = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(task.ID)), uiPrimary(task.Title))
				}
				dueText := ""
				if task.DueOn != "" {
					dueText = fmt.Sprintf(" due=%s", uiDue(task.DueOn))
				}
				fmt.Printf("%s  (%s/%s)%s parent=%s\n", label, uiKind(task.Kind), uiStatus(task.Status), dueText, parentLabel)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&kinds, "kind", nil, "Include kind (repeatable)")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Include status (repeatable)")
	cmd.Flags().StringArrayVar(&notKinds, "not-kind", nil, "Exclude kind (repeatable)")
	cmd.Flags().StringArrayVar(&notStatuses, "not-status", nil, "Exclude status (repeatable)")
	cmd.Flags().BoolVar(&ready, "ready", false, "Include only actionable tasks")
	cmd.Flags().BoolVar(&depsBlocked, "blocked-by-deps", false, "Include only tasks blocked by unresolved dependencies")
	cmd.Flags().StringVar(&dueBefore, "due-before", "", "Include only tasks due before this date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&dueAfter, "due-after", "", "Include only tasks due after this date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&overdue, "overdue", false, "Include only overdue tasks")
	cmd.Flags().BoolVar(&noDue, "no-due", false, "Include only tasks without due date")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&parent, "parent", "", "Filter by parent task ID or root")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of items")
	cmd.Flags().StringVar(&search, "search", "", "Search by title/body")
	return cmd
}

func newNextCommand(ctx *commandContext) *cobra.Command {
	var (
		limit  int
		asJSON bool
	)

	cmd := &cobra.Command{
		Use:   "next",
		Short: "List actionable tasks (ready to work on)",
		Example: "  shelf next\n" +
			"  shelf next --limit 20",
		RunE: func(_ *cobra.Command, _ []string) error {
			tasks, err := shelf.NewTaskStore(ctx.rootDir).List()
			if err != nil {
				return err
			}
			readiness, err := shelf.BuildTaskReadiness(ctx.rootDir)
			if err != nil {
				return err
			}

			titleByID := make(map[string]string, len(tasks))
			for _, task := range tasks {
				titleByID[task.ID] = task.Title
			}

			if asJSON {
				type nextItem struct {
					ID          string `json:"id"`
					Title       string `json:"title"`
					Kind        string `json:"kind"`
					Status      string `json:"status"`
					DueOn       string `json:"due_on,omitempty"`
					Parent      string `json:"parent,omitempty"`
					ParentTitle string `json:"parent_title,omitempty"`
				}
				items := make([]nextItem, 0)
				for _, task := range tasks {
					info, ok := readiness[task.ID]
					if !ok || !info.Ready {
						continue
					}
					parentTitle := ""
					if task.Parent != "" {
						parentTitle = titleByID[task.Parent]
					}
					items = append(items, nextItem{
						ID:          task.ID,
						Title:       task.Title,
						Kind:        string(task.Kind),
						Status:      string(task.Status),
						DueOn:       task.DueOn,
						Parent:      task.Parent,
						ParentTitle: parentTitle,
					})
					if limit > 0 && len(items) >= limit {
						break
					}
				}
				data, err := json.MarshalIndent(items, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			count := 0
			for _, task := range tasks {
				info, ok := readiness[task.ID]
				if !ok || !info.Ready {
					continue
				}
				parentLabel := uiMuted("root")
				if task.Parent != "" {
					if title, ok := titleByID[task.Parent]; ok {
						parentLabel = uiPrimary(title)
						if ctx.showID {
							parentLabel = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(task.Parent)), uiPrimary(title))
						}
					} else {
						parentLabel = uiMuted("(missing)")
					}
				}
				label := uiPrimary(task.Title)
				if ctx.showID {
					label = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(task.ID)), uiPrimary(task.Title))
				}
				dueText := ""
				if task.DueOn != "" {
					dueText = fmt.Sprintf(" due=%s", uiDue(task.DueOn))
				}
				fmt.Printf("%s  (%s/%s)%s parent=%s\n", label, uiKind(task.Kind), uiStatus(task.Status), dueText, parentLabel)
				count++
				if limit > 0 && count >= limit {
					break
				}
			}
			if count == 0 {
				fmt.Println(uiMuted("(none)"))
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of items")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newShowCommand(ctx *commandContext) *cobra.Command {
	var (
		noBody   bool
		onlyBody bool
		asJSON   bool
	)

	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show task details",
		Example: "  shelf show 01ABCDEFG...\n" +
			"  shelf show 01ABCDEFG... --no-body\n" +
			"  shelf show 01ABCDEFG... --json",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if noBody && onlyBody {
				return fmt.Errorf("--no-body と --only-body は同時に指定できません")
			}
			id, err := selectTaskIDIfMissing(ctx, args, "表示するタスクを選択", nil, true)
			if err != nil {
				return err
			}

			task, err := shelf.EnsureTaskExists(ctx.rootDir, id)
			if err != nil {
				return err
			}

			if onlyBody {
				if asJSON {
					payload := map[string]string{
						"id":   task.ID,
						"body": task.Body,
					}
					data, err := json.MarshalIndent(payload, "", "  ")
					if err != nil {
						return err
					}
					fmt.Println(string(data))
					return nil
				}
				fmt.Println(task.Body)
				return nil
			}

			allTasks, err := shelf.NewTaskStore(ctx.rootDir).List()
			if err != nil {
				return err
			}
			byID := make(map[string]shelf.Task, len(allTasks))
			for _, item := range allTasks {
				byID[item.ID] = item
			}
			path := buildTaskPath(task, byID)
			subtree, err := shelf.BuildTree(ctx.rootDir, shelf.TreeOptions{FromID: task.ID})
			if err != nil {
				return err
			}

			edgeStore := shelf.NewEdgeStore(ctx.rootDir)
			outbound, err := edgeStore.ListOutbound(task.ID)
			if err != nil {
				return err
			}
			inbound, err := edgeStore.FindInbound(task.ID)
			if err != nil {
				return err
			}

			if asJSON {
				type jsonTreeNode struct {
					ID       string         `json:"id"`
					Title    string         `json:"title"`
					Kind     string         `json:"kind"`
					Status   string         `json:"status"`
					DueOn    string         `json:"due_on,omitempty"`
					Parent   string         `json:"parent,omitempty"`
					Children []jsonTreeNode `json:"children,omitempty"`
				}
				var convert func(node shelf.TreeNode) jsonTreeNode
				convert = func(node shelf.TreeNode) jsonTreeNode {
					children := make([]jsonTreeNode, 0, len(node.Children))
					for _, child := range node.Children {
						children = append(children, convert(child))
					}
					return jsonTreeNode{
						ID:       node.Task.ID,
						Title:    node.Task.Title,
						Kind:     string(node.Task.Kind),
						Status:   string(node.Task.Status),
						DueOn:    node.Task.DueOn,
						Parent:   node.Task.Parent,
						Children: children,
					}
				}
				subtreePayload := make([]jsonTreeNode, 0, len(subtree))
				for _, node := range subtree {
					subtreePayload = append(subtreePayload, convert(node))
				}

				taskPayload := map[string]string{
					"id":         task.ID,
					"title":      task.Title,
					"kind":       string(task.Kind),
					"status":     string(task.Status),
					"due_on":     task.DueOn,
					"parent":     task.Parent,
					"created_at": task.CreatedAt.Format(time.RFC3339),
					"updated_at": task.UpdatedAt.Format(time.RFC3339),
				}
				if !noBody {
					taskPayload["body"] = task.Body
				}

				payload := map[string]any{
					"task":     taskPayload,
					"path":     path,
					"subtree":  subtreePayload,
					"outbound": outbound,
					"inbound":  inbound,
				}
				data, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Println("+++")
			fmt.Printf("id = %q\n", task.ID)
			fmt.Printf("title = %q\n", task.Title)
			fmt.Printf("kind = %q\n", task.Kind)
			fmt.Printf("status = %q\n", task.Status)
			if task.DueOn != "" {
				fmt.Printf("due_on = %q\n", task.DueOn)
			}
			if task.Parent != "" {
				fmt.Printf("parent = %q\n", task.Parent)
			}
			fmt.Printf("created_at = %q\n", task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
			fmt.Printf("updated_at = %q\n", task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
			fmt.Println("+++")
			if !noBody {
				fmt.Println()
				fmt.Println(task.Body)
				fmt.Println()
			}

			fmt.Println(uiHeading("Hierarchy:"))
			fmt.Printf("%s %s\n", uiMuted("Path:"), uiPrimary(path))
			fmt.Println(uiHeading("Subtree:"))
			if len(subtree) == 0 {
				fmt.Println(uiMuted("  (none)"))
			} else {
				for i, node := range subtree {
					printTreeNode(node, "", i == len(subtree)-1, ctx.showID)
				}
			}
			fmt.Println()

			fmt.Println(uiHeading("Outbound Links:"))
			if len(outbound) == 0 {
				fmt.Println(uiMuted("  (none)"))
			} else {
				for _, edge := range outbound {
					fmt.Printf("  %s --%s--> %s\n", uiShortID(shelf.ShortID(task.ID)), uiLinkType(edge.Type), uiShortID(shelf.ShortID(edge.To)))
				}
			}
			fmt.Println(uiHeading("Inbound Links:"))
			if len(inbound) == 0 {
				fmt.Println(uiMuted("  (none)"))
			} else {
				for _, edge := range inbound {
					fmt.Printf("  %s --%s--> %s\n", uiShortID(shelf.ShortID(edge.From)), uiLinkType(edge.Type), uiShortID(shelf.ShortID(task.ID)))
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&noBody, "no-body", false, "Hide body section")
	cmd.Flags().BoolVar(&onlyBody, "only-body", false, "Show only body")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newTreeCommand(ctx *commandContext) *cobra.Command {
	var (
		from     string
		maxDepth int
		kinds    []string
		statuses []string
		notKinds []string
		notStats []string
		asJSON   bool
	)

	cmd := &cobra.Command{
		Use:   "tree",
		Short: "Show task tree",
		Example: "  shelf tree\n" +
			"  shelf tree --kind todo --not-status done\n" +
			"  shelf tree --from root --max-depth 2 --json",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			for _, kind := range toKinds(kinds) {
				if err := cfg.ValidateKind(kind); err != nil {
					return err
				}
			}
			for _, kind := range toKinds(notKinds) {
				if err := cfg.ValidateKind(kind); err != nil {
					return err
				}
			}
			for _, status := range toStatuses(statuses) {
				if err := cfg.ValidateStatus(status); err != nil {
					return err
				}
			}
			for _, status := range toStatuses(notStats) {
				if err := cfg.ValidateStatus(status); err != nil {
					return err
				}
			}

			fromID := ""
			if strings.TrimSpace(from) != "" && !strings.EqualFold(from, "root") {
				fromID = from
			}
			nodes, err := shelf.BuildTree(ctx.rootDir, shelf.TreeOptions{
				FromID:      fromID,
				Kinds:       toKinds(kinds),
				Statuses:    toStatuses(statuses),
				NotKinds:    toKinds(notKinds),
				NotStatuses: toStatuses(notStats),
				MaxDepth:    maxDepth,
			})
			if err != nil {
				return err
			}

			if asJSON {
				type jsonTreeNode struct {
					ID       string         `json:"id"`
					Title    string         `json:"title"`
					Kind     string         `json:"kind"`
					Status   string         `json:"status"`
					DueOn    string         `json:"due_on,omitempty"`
					Parent   string         `json:"parent,omitempty"`
					Children []jsonTreeNode `json:"children,omitempty"`
				}
				var convert func(node shelf.TreeNode) jsonTreeNode
				convert = func(node shelf.TreeNode) jsonTreeNode {
					children := make([]jsonTreeNode, 0, len(node.Children))
					for _, child := range node.Children {
						children = append(children, convert(child))
					}
					return jsonTreeNode{
						ID:       node.Task.ID,
						Title:    node.Task.Title,
						Kind:     string(node.Task.Kind),
						Status:   string(node.Task.Status),
						DueOn:    node.Task.DueOn,
						Parent:   node.Task.Parent,
						Children: children,
					}
				}

				rows := make([]jsonTreeNode, 0, len(nodes))
				for _, node := range nodes {
					rows = append(rows, convert(node))
				}
				data, err := json.MarshalIndent(rows, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			for i, node := range nodes {
				printTreeNode(node, "", i == len(nodes)-1, ctx.showID)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "root", "Start from task ID or root")
	cmd.Flags().IntVar(&maxDepth, "max-depth", 0, "Maximum depth (0 means unlimited)")
	cmd.Flags().StringArrayVar(&kinds, "kind", nil, "Include kind (repeatable)")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Include status (repeatable)")
	cmd.Flags().StringArrayVar(&notKinds, "not-kind", nil, "Exclude kind (repeatable)")
	cmd.Flags().StringArrayVar(&notStats, "not-status", nil, "Exclude status (repeatable)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func printTreeNode(node shelf.TreeNode, prefix string, isLast bool, showID bool) {
	branch := "├─ "
	nextPrefix := prefix + "│  "
	if isLast {
		branch = "└─ "
		nextPrefix = prefix + "   "
	}
	if prefix == "" {
		branch = ""
	}

	label := uiPrimary(node.Task.Title)
	if showID {
		label = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(node.Task.ID)), uiPrimary(node.Task.Title))
	}
	dueText := ""
	if node.Task.DueOn != "" {
		dueText = fmt.Sprintf(" due=%s", uiDue(node.Task.DueOn))
	}
	fmt.Printf("%s%s%s (%s/%s)%s\n", uiMuted(prefix), uiMuted(branch), label, uiKind(node.Task.Kind), uiStatus(node.Task.Status), dueText)
	for i, child := range node.Children {
		printTreeNode(child, nextPrefix, i == len(node.Children)-1, showID)
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
