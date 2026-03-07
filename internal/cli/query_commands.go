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
		kinds           []string
		statuses        []string
		tags            []string
		notKinds        []string
		notStatuses     []string
		notTags         []string
		presetName      string
		view            string
		includeArchived bool
		onlyArchived    bool
		format          string
		ready           bool
		depsBlocked     bool
		dueBefore       string
		dueAfter        string
		overdue         bool
		noDue           bool
		asJSON          bool
		parent          string
		limit           int
		search          string
	)

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List tasks",
		Example: "  shelf ls\n" +
			"  shelf ls --kind todo --status open --status in_progress\n" +
			"  shelf ls --tag backend --not-tag wip\n" +
			"  shelf ls --ready --overdue\n" +
			"  shelf ls --json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			outputPreset, err := loadOutputPreset(ctx.rootDir, presetName, "ls")
			if err != nil {
				return err
			}
			view = applyPresetString(view, cmd.Flags().Changed("view"), outputPreset.View)
			format = applyPresetString(format, cmd.Flags().Changed("format"), outputPreset.Format)
			limit = applyPresetInt(limit, cmd.Flags().Changed("limit"), outputPreset.Limit)

			if err := validateFormat(format, []string{"compact", "detail", "kanban"}); err != nil {
				return err
			}
			viewPreset, err := resolveTaskView(ctx.rootDir, view)
			if err != nil {
				return err
			}

			filter := shelf.TaskFilter{
				Kinds:           toKinds(kinds),
				Statuses:        toStatuses(statuses),
				Tags:            parseTagFlagValues(tags),
				NotKinds:        toKinds(notKinds),
				NotStatuses:     toStatuses(notStatuses),
				NotTags:         parseTagFlagValues(notTags),
				IncludeArchived: includeArchived,
				OnlyArchived:    onlyArchived,
				ReadyOnly:       ready,
				DepsBlocked:     depsBlocked,
				DueBefore:       dueBefore,
				DueAfter:        dueAfter,
				Overdue:         overdue,
				NoDue:           noDue,
				Parent:          parent,
				Limit:           limit,
				Search:          search,
			}
			filter = mergeTaskFilterWithView(filter, viewPreset, map[string]bool{
				"ready":           cmd.Flags().Changed("ready"),
				"blocked-by-deps": cmd.Flags().Changed("blocked-by-deps"),
				"due-before":      cmd.Flags().Changed("due-before"),
				"due-after":       cmd.Flags().Changed("due-after"),
				"overdue":         cmd.Flags().Changed("overdue"),
				"no-due":          cmd.Flags().Changed("no-due"),
				"parent":          cmd.Flags().Changed("parent"),
				"search":          cmd.Flags().Changed("search"),
				"limit":           cmd.Flags().Changed("limit"),
			})

			tasks, err := shelf.ListTasks(ctx.rootDir, filter)
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
					ID          string   `json:"id"`
					Title       string   `json:"title"`
					Kind        string   `json:"kind"`
					Status      string   `json:"status"`
					Tags        []string `json:"tags,omitempty"`
					GitHubURLs  []string `json:"github_urls,omitempty"`
					EstimateMin int      `json:"estimate_minutes,omitempty"`
					SpentMin    int      `json:"spent_minutes,omitempty"`
					TimerStart  string   `json:"timer_started_at,omitempty"`
					DueOn       string   `json:"due_on,omitempty"`
					RepeatEvery string   `json:"repeat_every,omitempty"`
					ArchivedAt  string   `json:"archived_at,omitempty"`
					Parent      string   `json:"parent,omitempty"`
					ParentTitle string   `json:"parent_title,omitempty"`
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
						Tags:        task.Tags,
						GitHubURLs:  task.GitHubURLs,
						EstimateMin: task.EstimateMin,
						SpentMin:    task.SpentMin,
						TimerStart:  task.TimerStart,
						DueOn:       task.DueOn,
						RepeatEvery: task.RepeatEvery,
						ArchivedAt:  task.ArchivedAt,
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

			if format == "kanban" {
				statusOrder := []shelf.Status{"open", "in_progress", "blocked", "done", "cancelled"}
				grouped := map[shelf.Status][]shelf.Task{}
				for _, task := range tasks {
					grouped[task.Status] = append(grouped[task.Status], task)
				}
				for _, status := range statusOrder {
					fmt.Println(uiHeading(string(status) + ":"))
					rows := grouped[status]
					if len(rows) == 0 {
						fmt.Println(uiMuted("  (none)"))
						continue
					}
					for _, task := range rows {
						label := uiPrimary(task.Title)
						if ctx.showID {
							label = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(task.ID)), uiPrimary(task.Title))
						}
						dueText := uiMuted("-")
						if task.DueOn != "" {
							dueText = uiDue(task.DueOn)
						}
						fmt.Printf("  %s (%s) due=%s\n", label, uiKind(task.Kind), dueText)
					}
				}
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
				tagText := ""
				if len(task.Tags) > 0 {
					tagText = fmt.Sprintf(" tags=%s", strings.Join(task.Tags, ","))
				}
				archivedText := ""
				if task.ArchivedAt != "" {
					archivedText = " " + uiMuted("[archived]")
				}
				if format == "detail" {
					repeatText := "-"
					if task.RepeatEvery != "" {
						repeatText = task.RepeatEvery
					}
					fmt.Printf("%s kind=%s status=%s tags=%s github=%d estimate=%s spent=%s due=%s repeat=%s archived_at=%q parent=%s\n", label, uiKind(task.Kind), uiStatus(task.Status), formatTagSummary(task.Tags), len(task.GitHubURLs), shelf.FormatWorkMinutes(task.EstimateMin), shelf.FormatWorkMinutes(task.SpentMin), uiDue(task.DueOn), repeatText, task.ArchivedAt, parentLabel)
					continue
				}
				fmt.Printf("%s  (%s/%s)%s%s%s parent=%s\n", label, uiKind(task.Kind), uiStatus(task.Status), dueText, tagText, archivedText, parentLabel)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&kinds, "kind", nil, "Include kind (repeatable)")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Include status (repeatable)")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Include tag (repeatable)")
	cmd.Flags().StringArrayVar(&notKinds, "not-kind", nil, "Exclude kind (repeatable)")
	cmd.Flags().StringArrayVar(&notStatuses, "not-status", nil, "Exclude status (repeatable)")
	cmd.Flags().StringArrayVar(&notTags, "not-tag", nil, "Exclude tag (repeatable)")
	cmd.Flags().StringVar(&presetName, "preset", "", "Apply output preset for ls")
	cmd.Flags().StringVar(&view, "view", "", "Apply built-in view preset (active|ready|blocked|overdue)")
	cmd.Flags().BoolVar(&includeArchived, "include-archived", false, "Include archived tasks")
	cmd.Flags().BoolVar(&onlyArchived, "only-archived", false, "Include only archived tasks")
	cmd.Flags().StringVar(&format, "format", "compact", "Output format: compact|detail|kanban")
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
		presetName      string
		view            string
		includeArchived bool
		onlyArchived    bool
		limit           int
		asJSON          bool
	)

	cmd := &cobra.Command{
		Use:   "next",
		Short: "List actionable tasks (ready to work on)",
		Example: "  shelf next\n" +
			"  shelf next --limit 20\n" +
			"  shelf next --view active",
		RunE: func(cmd *cobra.Command, _ []string) error {
			outputPreset, err := loadOutputPreset(ctx.rootDir, presetName, "next")
			if err != nil {
				return err
			}
			view = applyPresetString(view, cmd.Flags().Changed("view"), outputPreset.View)
			limit = applyPresetInt(limit, cmd.Flags().Changed("limit"), outputPreset.Limit)
			preset, err := resolveTaskView(ctx.rootDir, view)
			if err != nil {
				return err
			}

			filter := mergeTaskFilterWithView(shelf.TaskFilter{Limit: 0}, preset, map[string]bool{
				"limit": true,
			})
			filter.IncludeArchived = includeArchived
			filter.OnlyArchived = onlyArchived
			tasks, err := shelf.ListTasks(ctx.rootDir, filter)
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
					RepeatEvery string `json:"repeat_every,omitempty"`
					ArchivedAt  string `json:"archived_at,omitempty"`
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
						RepeatEvery: task.RepeatEvery,
						ArchivedAt:  task.ArchivedAt,
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

	cmd.Flags().StringVar(&presetName, "preset", "", "Apply output preset for next")
	cmd.Flags().StringVar(&view, "view", "", "Apply built-in view preset (active|ready|blocked|overdue)")
	cmd.Flags().BoolVar(&includeArchived, "include-archived", false, "Include archived tasks")
	cmd.Flags().BoolVar(&onlyArchived, "only-archived", false, "Include only archived tasks")
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
		Args: cobra.MaximumNArgs(1),
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
			readinessMap, err := shelf.BuildTaskReadiness(ctx.rootDir)
			if err != nil {
				return err
			}
			readiness, ok := readinessMap[task.ID]
			if !ok {
				readiness = shelf.TaskReadiness{}
			}

			if asJSON {
				type jsonTreeNode struct {
					ID          string         `json:"id"`
					Title       string         `json:"title"`
					Kind        string         `json:"kind"`
					Status      string         `json:"status"`
					Tags        []string       `json:"tags,omitempty"`
					GitHubURLs  []string       `json:"github_urls,omitempty"`
					EstimateMin int            `json:"estimate_minutes,omitempty"`
					SpentMin    int            `json:"spent_minutes,omitempty"`
					TimerStart  string         `json:"timer_started_at,omitempty"`
					DueOn       string         `json:"due_on,omitempty"`
					RepeatEvery string         `json:"repeat_every,omitempty"`
					ArchivedAt  string         `json:"archived_at,omitempty"`
					Parent      string         `json:"parent,omitempty"`
					Children    []jsonTreeNode `json:"children,omitempty"`
				}
				var convert func(node shelf.TreeNode) jsonTreeNode
				convert = func(node shelf.TreeNode) jsonTreeNode {
					children := make([]jsonTreeNode, 0, len(node.Children))
					for _, child := range node.Children {
						children = append(children, convert(child))
					}
					return jsonTreeNode{
						ID:          node.Task.ID,
						Title:       node.Task.Title,
						Kind:        string(node.Task.Kind),
						Status:      string(node.Task.Status),
						Tags:        node.Task.Tags,
						GitHubURLs:  node.Task.GitHubURLs,
						EstimateMin: node.Task.EstimateMin,
						SpentMin:    node.Task.SpentMin,
						TimerStart:  node.Task.TimerStart,
						DueOn:       node.Task.DueOn,
						RepeatEvery: node.Task.RepeatEvery,
						ArchivedAt:  node.Task.ArchivedAt,
						Parent:      node.Task.Parent,
						Children:    children,
					}
				}
				subtreePayload := make([]jsonTreeNode, 0, len(subtree))
				for _, node := range subtree {
					subtreePayload = append(subtreePayload, convert(node))
				}

				taskPayload := map[string]any{
					"id":               task.ID,
					"title":            task.Title,
					"kind":             string(task.Kind),
					"status":           string(task.Status),
					"tags":             task.Tags,
					"github_urls":      task.GitHubURLs,
					"estimate_minutes": task.EstimateMin,
					"spent_minutes":    task.SpentMin,
					"timer_started_at": task.TimerStart,
					"due_on":           task.DueOn,
					"repeat_every":     task.RepeatEvery,
					"archived_at":      task.ArchivedAt,
					"parent":           task.Parent,
					"created_at":       task.CreatedAt.Format(time.RFC3339),
					"updated_at":       task.UpdatedAt.Format(time.RFC3339),
				}
				if !noBody {
					taskPayload["body"] = task.Body
				}

				payload := map[string]any{
					"task":      taskPayload,
					"path":      path,
					"subtree":   subtreePayload,
					"readiness": readiness,
					"outbound":  outbound,
					"inbound":   inbound,
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
			if len(task.Tags) > 0 {
				fmt.Printf("tags = [%q", task.Tags[0])
				for i := 1; i < len(task.Tags); i++ {
					fmt.Printf(", %q", task.Tags[i])
				}
				fmt.Println("]")
			}
			if len(task.GitHubURLs) > 0 {
				fmt.Printf("github_urls = [%q", task.GitHubURLs[0])
				for i := 1; i < len(task.GitHubURLs); i++ {
					fmt.Printf(", %q", task.GitHubURLs[i])
				}
				fmt.Println("]")
			}
			if task.EstimateMin > 0 {
				fmt.Printf("estimate_minutes = %d\n", task.EstimateMin)
			}
			if task.SpentMin > 0 {
				fmt.Printf("spent_minutes = %d\n", task.SpentMin)
			}
			if strings.TrimSpace(task.TimerStart) != "" {
				fmt.Printf("timer_started_at = %q\n", task.TimerStart)
			}
			if task.DueOn != "" {
				fmt.Printf("due_on = %q\n", task.DueOn)
			}
			if task.RepeatEvery != "" {
				fmt.Printf("repeat_every = %q\n", task.RepeatEvery)
			}
			if task.ArchivedAt != "" {
				fmt.Printf("archived_at = %q\n", task.ArchivedAt)
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
			fmt.Println(uiHeading("Context Tree:"))
			fullTree, err := shelf.BuildTree(ctx.rootDir, shelf.TreeOptions{})
			if err != nil {
				return err
			}
			if len(fullTree) == 0 {
				fmt.Println(uiMuted("  (none)"))
			} else {
				if !printContextTree(fullTree, task.ID, "", true, ctx.showID) {
					for i, node := range subtree {
						printTreeNode(node, "", i == len(subtree)-1, ctx.showID, "compact")
					}
				}
			}
			fmt.Println()
			fmt.Println(uiHeading("Readiness:"))
			if readiness.Ready {
				fmt.Println("  ready=true")
			} else {
				fmt.Println("  ready=false")
			}
			if len(readiness.UnresolvedDependsOn) == 0 {
				fmt.Println("  blocked_by_dependencies=(none)")
			} else {
				fmt.Println("  blocked_by_dependencies:")
				for _, depID := range readiness.UnresolvedDependsOn {
					depLabel := depID
					if depTask, ok := byID[depID]; ok {
						depLabel = depTask.Title
						if ctx.showID {
							depLabel = fmt.Sprintf("[%s] %s", shelf.ShortID(depID), depTask.Title)
						}
					}
					fmt.Printf("    - %s\n", depLabel)
				}
			}
			fmt.Println()

			fmt.Println(uiHeading("Outbound Links:"))
			if len(outbound) == 0 {
				fmt.Println(uiMuted("  (none)"))
			} else {
				for _, edge := range outbound {
					fmt.Printf("  %s --%s--> %s\n", taskLabelForLink(task.ID, byID, ctx.showID), uiLinkType(edge.Type), taskLabelForLink(edge.To, byID, ctx.showID))
				}
			}
			fmt.Println(uiHeading("Inbound Links:"))
			if len(inbound) == 0 {
				fmt.Println(uiMuted("  (none)"))
			} else {
				for _, edge := range inbound {
					fmt.Printf("  %s --%s--> %s\n", taskLabelForLink(edge.From, byID, ctx.showID), uiLinkType(edge.Type), taskLabelForLink(task.ID, byID, ctx.showID))
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
		from            string
		maxDepth        int
		presetName      string
		view            string
		includeArchived bool
		onlyArchived    bool
		format          string
		kinds           []string
		statuses        []string
		tags            []string
		notKinds        []string
		notStats        []string
		notTags         []string
		plain           bool
		asJSON          bool
	)

	cmd := &cobra.Command{
		Use:     "tree",
		Aliases: []string{"tr"},
		Short:   "Show task tree",
		Example: "  shelf tree\n" +
			"  shelf tree --plain\n" +
			"  shelf tree --kind todo --not-status done --tag backend\n" +
			"  shelf tree --from root --max-depth 2 --json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			outputPreset, err := loadOutputPreset(ctx.rootDir, presetName, "tree")
			if err != nil {
				return err
			}
			view = applyPresetString(view, cmd.Flags().Changed("view"), outputPreset.View)
			format = applyPresetString(format, cmd.Flags().Changed("format"), outputPreset.Format)

			if err := validateFormat(format, []string{"compact", "detail"}); err != nil {
				return err
			}
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
			for _, tag := range parseTagFlagValues(tags) {
				if err := cfg.ValidateTag(tag); err != nil {
					return err
				}
			}
			for _, tag := range parseTagFlagValues(notTags) {
				if err := cfg.ValidateTag(tag); err != nil {
					return err
				}
			}

			fromID := ""
			if strings.TrimSpace(from) != "" && !strings.EqualFold(from, "root") {
				fromID = from
			}
			treeOpts := shelf.TreeOptions{
				FromID:          fromID,
				Kinds:           toKinds(kinds),
				Statuses:        toStatuses(statuses),
				Tags:            parseTagFlagValues(tags),
				NotKinds:        toKinds(notKinds),
				NotStatuses:     toStatuses(notStats),
				NotTags:         parseTagFlagValues(notTags),
				IncludeArchived: includeArchived,
				OnlyArchived:    onlyArchived,
				MaxDepth:        maxDepth,
			}
			preset, err := resolveTaskView(ctx.rootDir, view)
			if err != nil {
				return err
			}
			filter := shelf.TaskFilter{
				Kinds:           toKinds(kinds),
				Statuses:        toStatuses(statuses),
				Tags:            parseTagFlagValues(tags),
				NotKinds:        toKinds(notKinds),
				NotStatuses:     toStatuses(notStats),
				NotTags:         parseTagFlagValues(notTags),
				IncludeArchived: includeArchived,
				OnlyArchived:    onlyArchived,
				Limit:           0,
			}
			filter = mergeTaskFilterWithView(filter, preset, map[string]bool{
				"limit": true,
			})
			treeOpts, err = treeOptionsFromFilter(treeOpts, filter)
			if err != nil {
				return err
			}
			if resolveTreeOutputMode(dailyCockpitIsTTY(), asJSON, plain) == dailyCockpitOutputTUI {
				startDate, dayCount, err := resolveDailyCockpitRange(ctx.rootDir)
				if err != nil {
					return err
				}
				return runCalendarModeTUIFn(ctx.rootDir, startDate, dayCount, cfg.Statuses, calendarTUIOptions{
					Mode:   calendarModeTree,
					ShowID: ctx.showID,
					Filter: filter,
				})
			}
			nodes, err := shelf.BuildTree(ctx.rootDir, treeOpts)
			if err != nil {
				return err
			}

			if asJSON {
				type jsonTreeNode struct {
					ID          string         `json:"id"`
					Title       string         `json:"title"`
					Kind        string         `json:"kind"`
					Status      string         `json:"status"`
					Tags        []string       `json:"tags,omitempty"`
					GitHubURLs  []string       `json:"github_urls,omitempty"`
					DueOn       string         `json:"due_on,omitempty"`
					RepeatEvery string         `json:"repeat_every,omitempty"`
					ArchivedAt  string         `json:"archived_at,omitempty"`
					Parent      string         `json:"parent,omitempty"`
					Children    []jsonTreeNode `json:"children,omitempty"`
				}
				var convert func(node shelf.TreeNode) jsonTreeNode
				convert = func(node shelf.TreeNode) jsonTreeNode {
					children := make([]jsonTreeNode, 0, len(node.Children))
					for _, child := range node.Children {
						children = append(children, convert(child))
					}
					return jsonTreeNode{
						ID:          node.Task.ID,
						Title:       node.Task.Title,
						Kind:        string(node.Task.Kind),
						Status:      string(node.Task.Status),
						Tags:        node.Task.Tags,
						GitHubURLs:  node.Task.GitHubURLs,
						DueOn:       node.Task.DueOn,
						RepeatEvery: node.Task.RepeatEvery,
						ArchivedAt:  node.Task.ArchivedAt,
						Parent:      node.Task.Parent,
						Children:    children,
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
				printTreeNode(node, "", i == len(nodes)-1, ctx.showID, format)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "root", "Start from task ID or root")
	cmd.Flags().IntVar(&maxDepth, "max-depth", 0, "Maximum depth (0 means unlimited)")
	cmd.Flags().StringVar(&presetName, "preset", "", "Apply output preset for tree")
	cmd.Flags().StringVar(&view, "view", "", "Apply built-in view preset (active|ready|blocked|overdue)")
	cmd.Flags().BoolVar(&includeArchived, "include-archived", false, "Include archived tasks")
	cmd.Flags().BoolVar(&onlyArchived, "only-archived", false, "Include only archived tasks")
	cmd.Flags().StringVar(&format, "format", "compact", "Output format: compact|detail")
	cmd.Flags().StringArrayVar(&kinds, "kind", nil, "Include kind (repeatable)")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Include status (repeatable)")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Include tag (repeatable)")
	cmd.Flags().StringArrayVar(&notKinds, "not-kind", nil, "Exclude kind (repeatable)")
	cmd.Flags().StringArrayVar(&notStats, "not-status", nil, "Exclude status (repeatable)")
	cmd.Flags().StringArrayVar(&notTags, "not-tag", nil, "Exclude tag (repeatable)")
	cmd.Flags().BoolVar(&plain, "plain", false, "Force plain text output even on TTY")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func printTreeNode(node shelf.TreeNode, prefix string, isLast bool, showID bool, format string) {
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
	tagText := ""
	if len(node.Task.Tags) > 0 {
		tagText = fmt.Sprintf(" tags=%s", strings.Join(node.Task.Tags, ","))
	}
	if format == "detail" {
		repeatText := "-"
		if node.Task.RepeatEvery != "" {
			repeatText = node.Task.RepeatEvery
		}
		fmt.Printf("%s%s%s kind=%s status=%s tags=%s due=%s repeat=%s archived_at=%q\n", uiMuted(prefix), uiMuted(branch), label, uiKind(node.Task.Kind), uiStatus(node.Task.Status), formatTagSummary(node.Task.Tags), uiDue(node.Task.DueOn), repeatText, node.Task.ArchivedAt)
	} else {
		fmt.Printf("%s%s%s (%s/%s)%s%s\n", uiMuted(prefix), uiMuted(branch), label, uiKind(node.Task.Kind), uiStatus(node.Task.Status), dueText, tagText)
	}
	for i, child := range node.Children {
		printTreeNode(child, nextPrefix, i == len(node.Children)-1, showID, format)
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

func printContextTree(nodes []shelf.TreeNode, targetID string, prefix string, isRoot bool, showID bool) bool {
	printedAny := false
	matched := make([]shelf.TreeNode, 0, len(nodes))
	for _, node := range nodes {
		if containsTask(node, targetID) {
			matched = append(matched, node)
		}
	}
	for i, node := range matched {
		isLast := i == len(nodes)-1
		isLast = i == len(matched)-1
		printedAny = true
		branch := "├─ "
		nextPrefix := prefix + "│  "
		if isLast {
			branch = "└─ "
			nextPrefix = prefix + "   "
		}
		if isRoot {
			branch = ""
			if isLast {
				nextPrefix = "   "
			} else {
				nextPrefix = "│  "
			}
		}

		label := uiPrimary(node.Task.Title)
		if showID {
			label = fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(node.Task.ID)), uiPrimary(node.Task.Title))
		}
		if node.Task.ID == targetID {
			label = uiColor(label, "1;33")
		}
		dueText := ""
		if node.Task.DueOn != "" {
			dueText = fmt.Sprintf(" due=%s", uiDue(node.Task.DueOn))
		}
		fmt.Printf("%s%s%s (%s/%s)%s\n", uiMuted(prefix), uiMuted(branch), label, uiKind(node.Task.Kind), uiStatus(node.Task.Status), dueText)

		children := node.Children
		if node.Task.ID != targetID {
			filtered := make([]shelf.TreeNode, 0, len(node.Children))
			for _, child := range node.Children {
				if containsTask(child, targetID) {
					filtered = append(filtered, child)
				}
			}
			children = filtered
		}
		_ = printContextTree(children, targetID, nextPrefix, false, showID)
	}
	return printedAny
}

func containsTask(node shelf.TreeNode, targetID string) bool {
	if node.Task.ID == targetID {
		return true
	}
	for _, child := range node.Children {
		if containsTask(child, targetID) {
			return true
		}
	}
	return false
}

func taskLabelForLink(taskID string, byID map[string]shelf.Task, showID bool) string {
	task, ok := byID[taskID]
	if !ok {
		if showID {
			return uiShortID(shelf.ShortID(taskID))
		}
		return uiMuted("(missing)")
	}
	if showID {
		return fmt.Sprintf("%s %s", uiShortID(shelf.ShortID(task.ID)), uiPrimary(task.Title))
	}
	return uiPrimary(task.Title)
}
