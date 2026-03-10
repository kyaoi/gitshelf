package cli

import (
	"encoding/json"
	"fmt"
	"strings"

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
		preset          string
		fields          string
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
			if err := validateFormat(format, []string{"compact", "detail", "kanban", "tree", "tsv"}); err != nil {
				return err
			}
			if strings.TrimSpace(fields) != "" && format != "tsv" {
				return fmt.Errorf("--fields requires --format tsv")
			}
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			if err := applyLsPreset(cmd, preset, cfg, &format, &ready, &statuses, &notStatuses); err != nil {
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

			tasks, err := shelf.ListTasks(ctx.rootDir, filter)
			if err != nil {
				return err
			}
			allTasks, err := shelf.NewTaskStore(ctx.rootDir).List()
			if err != nil {
				return err
			}
			byID := make(map[string]shelf.Task, len(allTasks))
			for _, task := range allTasks {
				byID[task.ID] = task
			}

			if asJSON {
				items := make([]taskQueryRecord, 0, len(tasks))
				for _, task := range tasks {
					items = append(items, buildTaskQueryRecord(ctx.rootDir, task, byID))
				}
				data, err := json.MarshalIndent(items, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			if format == "tree" {
				fromID := filter.Parent
				if fromID == "root" {
					fromID = ""
				}
				nodes, err := shelf.BuildTree(ctx.rootDir, shelf.TreeOptions{
					Kinds:           filter.Kinds,
					Statuses:        filter.Statuses,
					Tags:            filter.Tags,
					NotKinds:        filter.NotKinds,
					NotStatuses:     filter.NotStatuses,
					NotTags:         filter.NotTags,
					IncludeArchived: filter.IncludeArchived,
					OnlyArchived:    filter.OnlyArchived,
					FromID:          fromID,
				})
				if err != nil {
					return err
				}
				if filter.Parent == "root" {
					rootNodes := make([]shelf.TreeNode, 0, len(nodes))
					for _, node := range nodes {
						if node.Task.Parent == "" {
							rootNodes = append(rootNodes, node)
						}
					}
					nodes = rootNodes
				}
				printed := 0
				for i, node := range nodes {
					printTreeNode(node, "", i == len(nodes)-1, ctx.showID, "compact")
					printed++
					if filter.Limit > 0 && printed >= filter.Limit {
						break
					}
				}
				if printed == 0 {
					fmt.Println(uiMuted("(none)"))
				}
				return nil
			}

			if format == "tsv" {
				selectedFields, err := resolveTSVFields(fields, defaultLsTSVFields(), allowedLsTSVFields())
				if err != nil {
					return err
				}
				for _, task := range tasks {
					fmt.Println(joinTSVFields(selectedFields, buildTaskQueryRecord(ctx.rootDir, task, byID).TSVFields()))
				}
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
					if parent, ok := byID[task.Parent]; ok {
						parentLabel = formatTaskPathLabel(parent, byID, ctx.showID)
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
					fmt.Printf("%s kind=%s status=%s tags=%s due=%s repeat=%s archived_at=%q parent=%s\n", label, uiKind(task.Kind), uiStatus(task.Status), formatTagSummary(task.Tags), uiDue(task.DueOn), repeatText, task.ArchivedAt, parentLabel)
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
	cmd.Flags().StringVar(&preset, "preset", "", "Apply read-only defaults similar to a Cockpit view: now|review|board")
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated field names for --format tsv")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of items")
	cmd.Flags().StringVar(&search, "search", "", "Search by title/body")
	return cmd
}

func newNextCommand(ctx *commandContext) *cobra.Command {
	var (
		limit  int
		asJSON bool
		format string
		fields string
	)

	cmd := &cobra.Command{
		Use:   "next",
		Short: "List actionable tasks (ready to work on)",
		Example: "  shelf next\n" +
			"  shelf next --limit 20\n" +
			"  shelf next --format tsv",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := validateFormat(format, []string{"compact", "tsv"}); err != nil {
				return err
			}
			if strings.TrimSpace(fields) != "" && format != "tsv" {
				return fmt.Errorf("--fields requires --format tsv")
			}
			filter := shelf.TaskFilter{Limit: 0}
			tasks, err := shelf.ListTasks(ctx.rootDir, filter)
			if err != nil {
				return err
			}
			readiness, err := shelf.BuildTaskReadiness(ctx.rootDir)
			if err != nil {
				return err
			}

			byID := make(map[string]shelf.Task, len(tasks))
			for _, task := range tasks {
				byID[task.ID] = task
			}

			if asJSON {
				items := make([]taskQueryRecord, 0)
				for _, task := range tasks {
					info, ok := readiness[task.ID]
					if !ok || !info.Ready {
						continue
					}
					items = append(items, buildTaskQueryRecord(ctx.rootDir, task, byID))
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

			if format == "tsv" {
				selectedFields, err := resolveTSVFields(fields, defaultNextTSVFields(), allowedNextTSVFields())
				if err != nil {
					return err
				}
				count := 0
				for _, task := range tasks {
					info, ok := readiness[task.ID]
					if !ok || !info.Ready {
						continue
					}
					fmt.Println(joinTSVFields(selectedFields, buildTaskQueryRecord(ctx.rootDir, task, byID).TSVFields()))
					count++
					if limit > 0 && count >= limit {
						break
					}
				}
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
					if parent, ok := byID[task.Parent]; ok {
						parentLabel = formatTaskPathLabel(parent, byID, ctx.showID)
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
	cmd.Flags().StringVar(&format, "format", "compact", "Output format: compact|tsv")
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated field names for --format tsv")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newTreeCommand(ctx *commandContext) *cobra.Command {
	var flags cockpitLaunchFlags

	cmd := &cobra.Command{
		Use:     "tree",
		Aliases: []string{"tr"},
		Short:   "Open Cockpit in tree mode",
		Example: "  shelf tree\n" +
			"  shelf tree --kind todo --not-status done --tag backend\n" +
			"  shelf tree --months 3 --status open",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !dailyCockpitIsTTY() {
				return fmt.Errorf("tree はTTYが必要です")
			}
			return runCockpitLaunch(ctx, cmd, calendarModeTree, flags)
		},
	}

	addCockpitLaunchFlags(cmd, &flags)
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

func formatTaskPathLabel(task shelf.Task, byID map[string]shelf.Task, showID bool) string {
	label := buildTaskPath(task, byID)
	if showID {
		return fmt.Sprintf("%s [%s]", label, shelf.ShortID(task.ID))
	}
	return label
}

func sanitizeTSVField(value string) string {
	value = strings.ReplaceAll(value, "\t", " ")
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}

func resolveTSVFields(raw string, defaults []string, allowed map[string]struct{}) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return append([]string{}, defaults...), nil
	}
	parts := strings.Split(raw, ",")
	fields := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		field := strings.TrimSpace(part)
		if field == "" {
			continue
		}
		if _, ok := allowed[field]; !ok {
			return nil, fmt.Errorf("unknown --fields entry: %s", field)
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		fields = append(fields, field)
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("--fields must not be empty")
	}
	return fields, nil
}

func joinTSVFields(fields []string, row map[string]string) string {
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		values = append(values, sanitizeTSVField(row[field]))
	}
	return strings.Join(values, "\t")
}

func defaultLsTSVFields() []string {
	return []string{"id", "title", "path", "kind", "status", "due_on", "repeat_every", "archived_at", "parent", "parent_path", "tags", "file"}
}

func allowedLsTSVFields() map[string]struct{} {
	return map[string]struct{}{
		"id": {}, "title": {}, "path": {}, "kind": {}, "status": {}, "due_on": {}, "repeat_every": {},
		"archived_at": {}, "parent": {}, "parent_path": {}, "tags": {}, "file": {},
	}
}

func defaultNextTSVFields() []string {
	return []string{"id", "title", "path", "kind", "status", "due_on", "repeat_every", "parent", "parent_path", "tags", "file"}
}

func allowedNextTSVFields() map[string]struct{} {
	return map[string]struct{}{
		"id": {}, "title": {}, "path": {}, "kind": {}, "status": {}, "due_on": {}, "repeat_every": {},
		"parent": {}, "parent_path": {}, "tags": {}, "file": {},
	}
}

func applyLsPreset(cmd *cobra.Command, preset string, cfg shelf.Config, format *string, ready *bool, statuses *[]string, notStatuses *[]string) error {
	switch strings.TrimSpace(preset) {
	case "":
		return nil
	case "now":
		if !cmd.Flags().Changed("status") && !cmd.Flags().Changed("not-status") {
			*statuses = statusStrings(defaultCockpitStatuses(calendarModeNow, cfg))
		}
		if !cmd.Flags().Changed("ready") && !cmd.Flags().Changed("blocked-by-deps") && !cmd.Flags().Changed("status") && !cmd.Flags().Changed("not-status") {
			*ready = true
		}
		return nil
	case "review":
		if !cmd.Flags().Changed("status") && !cmd.Flags().Changed("not-status") {
			*statuses = statusStrings(defaultCockpitStatuses(calendarModeReview, cfg))
		}
		if !cmd.Flags().Changed("format") {
			*format = "detail"
		}
		return nil
	case "board":
		if !cmd.Flags().Changed("format") {
			*format = "kanban"
		}
		if !cmd.Flags().Changed("status") && !cmd.Flags().Changed("not-status") {
			*statuses = statusStrings(defaultCockpitStatuses(calendarModeBoard, cfg))
			*notStatuses = nil
		}
		return nil
	default:
		return fmt.Errorf("unknown --preset: %s (allowed: now|review|board)", preset)
	}
}
