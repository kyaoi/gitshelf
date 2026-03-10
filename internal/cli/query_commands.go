package cli

import (
	"encoding/json"
	"fmt"
	"sort"
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
		header          bool
		noHeader        bool
		sortBy          string
		reverse         bool
		groupBy         string
		countOnly       bool
		limit           int
		search          string
		schemaValue     string
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
			if err := validateFormat(format, []string{"compact", "detail", "kanban", "tree", "tsv", "csv", "jsonl"}); err != nil {
				return err
			}
			if err := validateLsGrouping(format, groupBy, countOnly); err != nil {
				return err
			}
			schema, err := parseOutputSchema(schemaValue)
			if err != nil {
				return err
			}
			if err := validateCountModeFlags(cmd, countOnly, fields, header, noHeader, sortBy, reverse, limit); err != nil {
				return err
			}
			if strings.TrimSpace(fields) != "" && format != "tsv" && format != "csv" {
				return fmt.Errorf("--fields requires --format tsv or csv")
			}
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			if err := applyLsPreset(cmd, preset, cfg, &format, &ready, &statuses, &notStatuses); err != nil {
				return err
			}

			filterLimit := limit
			if countOnly {
				filterLimit = 0
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
				Limit:           filterLimit,
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
			if err := sortTaskQueryResults(tasks, byID, sortBy, reverse); err != nil {
				return err
			}
			if countOnly {
				return printCountResult(len(tasks), asJSON)
			}
			groupedRecords := buildGroupedTaskQueryRecords(ctx.rootDir, tasks, byID, groupBy)
			groupedRecordsV2 := buildGroupedTaskQueryRecordsV2(ctx.rootDir, tasks, byID, groupBy)

			if asJSON {
				items := groupedTaskRecordsToAny(schema, groupedRecords, groupedRecordsV2, groupBy)
				data, err := json.MarshalIndent(items, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			if format == "jsonl" {
				text, err := renderGroupedTaskJSONL(schema, groupedRecords, groupedRecordsV2, groupBy)
				if err != nil {
					return err
				}
				fmt.Print(text)
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
				selectedFields, err := resolveTSVFields(fields, defaultLsTabularFields(schema, groupBy), allowedLsTabularFields(schema, groupBy))
				if err != nil {
					return err
				}
				switch schema {
				case outputSchemaV2:
					for _, record := range groupedRecordsV2 {
						fmt.Println(joinTSVFields(selectedFields, record.TSVFields()))
					}
				default:
					for _, record := range groupedRecords {
						fmt.Println(joinTSVFields(selectedFields, record.TSVFields()))
					}
				}
				return nil
			}

			if format == "csv" {
				selectedFields, err := resolveTSVFields(fields, defaultLsTabularFields(schema, groupBy), allowedLsTabularFields(schema, groupBy))
				if err != nil {
					return err
				}
				includeHeader, err := resolveTabularHeader(format, header, noHeader)
				if err != nil {
					return err
				}
				switch schema {
				case outputSchemaV2:
					text, err := renderCSV(groupedRecordsV2, selectedFields, includeHeader)
					if err != nil {
						return err
					}
					fmt.Print(text)
				default:
					text, err := renderCSV(groupedRecords, selectedFields, includeHeader)
					if err != nil {
						return err
					}
					fmt.Print(text)
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

			if strings.TrimSpace(groupBy) != "" {
				printGroupedLsRecords(groupedRecords, format, ctx.showID)
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
	cmd.Flags().StringVar(&format, "format", "compact", "Output format: compact|detail|kanban|tree|tsv|csv|jsonl")
	cmd.Flags().BoolVar(&ready, "ready", false, "Include only actionable tasks")
	cmd.Flags().BoolVar(&depsBlocked, "blocked-by-deps", false, "Include only tasks blocked by unresolved dependencies")
	cmd.Flags().StringVar(&dueBefore, "due-before", "", "Include only tasks due before this date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&dueAfter, "due-after", "", "Include only tasks due after this date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&overdue, "overdue", false, "Include only overdue tasks")
	cmd.Flags().BoolVar(&noDue, "no-due", false, "Include only tasks without due date")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&parent, "parent", "", "Filter by parent task ID or root")
	cmd.Flags().StringVar(&preset, "preset", "", "Apply read-only defaults similar to a Cockpit view: now|review|board")
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated field names for --format tsv or csv")
	cmd.Flags().BoolVar(&header, "header", false, "Include a header row for tabular output")
	cmd.Flags().BoolVar(&noHeader, "no-header", false, "Omit the header row for tabular output")
	cmd.Flags().StringVar(&sortBy, "sort", "", "Sort by: id|title|path|kind|status|due_on|created_at|updated_at")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "Reverse the sort order")
	cmd.Flags().StringVar(&groupBy, "group-by", "", "Group tasks by: status|kind|parent")
	cmd.Flags().BoolVar(&countOnly, "count", false, "Print only the total number of matching tasks")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of items")
	cmd.Flags().StringVar(&search, "search", "", "Search by title/body")
	cmd.Flags().StringVar(&schemaValue, "schema", "v1", "Machine-readable schema: v1|v2")
	return cmd
}

func newNextCommand(ctx *commandContext) *cobra.Command {
	var (
		limit       int
		asJSON      bool
		format      string
		fields      string
		header      bool
		noHeader    bool
		sortBy      string
		reverse     bool
		countOnly   bool
		schemaValue string
	)

	cmd := &cobra.Command{
		Use:   "next",
		Short: "List actionable tasks (ready to work on)",
		Example: "  shelf next\n" +
			"  shelf next --limit 20\n" +
			"  shelf next --format tsv",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateFormat(format, []string{"compact", "tsv", "csv", "jsonl"}); err != nil {
				return err
			}
			schema, err := parseOutputSchema(schemaValue)
			if err != nil {
				return err
			}
			if err := validateCountModeFlags(cmd, countOnly, fields, header, noHeader, sortBy, reverse, limit); err != nil {
				return err
			}
			if strings.TrimSpace(fields) != "" && format != "tsv" && format != "csv" {
				return fmt.Errorf("--fields requires --format tsv or csv")
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
			if err := sortTaskQueryResults(tasks, byID, sortBy, reverse); err != nil {
				return err
			}
			if countOnly {
				count := 0
				for _, task := range tasks {
					info, ok := readiness[task.ID]
					if ok && info.Ready {
						count++
					}
				}
				return printCountResult(count, asJSON)
			}

			if asJSON {
				items := buildNextJSONItems(schema, ctx.rootDir, tasks, readiness, byID, limit)
				data, err := json.MarshalIndent(items, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			if format == "jsonl" {
				text, err := renderNextJSONL(schema, ctx.rootDir, tasks, readiness, byID, limit)
				if err != nil {
					return err
				}
				fmt.Print(text)
				return nil
			}

			if format == "tsv" {
				selectedFields, err := resolveTSVFields(fields, defaultNextTSVFields(schema), allowedNextTSVFields(schema))
				if err != nil {
					return err
				}
				printNextTSV(schema, selectedFields, ctx.rootDir, tasks, readiness, byID, limit)
				return nil
			}

			if format == "csv" {
				selectedFields, err := resolveTSVFields(fields, defaultNextTSVFields(schema), allowedNextTSVFields(schema))
				if err != nil {
					return err
				}
				includeHeader, err := resolveTabularHeader(format, header, noHeader)
				if err != nil {
					return err
				}
				text, err := renderNextCSV(schema, selectedFields, includeHeader, ctx.rootDir, tasks, readiness, byID, limit)
				if err != nil {
					return err
				}
				fmt.Print(text)
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
	cmd.Flags().StringVar(&format, "format", "compact", "Output format: compact|tsv|csv|jsonl")
	cmd.Flags().StringVar(&fields, "fields", "", "Comma-separated field names for --format tsv or csv")
	cmd.Flags().BoolVar(&header, "header", false, "Include a header row for tabular output")
	cmd.Flags().BoolVar(&noHeader, "no-header", false, "Omit the header row for tabular output")
	cmd.Flags().StringVar(&sortBy, "sort", "", "Sort by: id|title|path|kind|status|due_on|created_at|updated_at")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "Reverse the sort order")
	cmd.Flags().BoolVar(&countOnly, "count", false, "Print only the total number of ready tasks")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&schemaValue, "schema", "v1", "Machine-readable schema: v1|v2")
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

func defaultLsTSVFields(schema outputSchema) []string {
	parentField := "parent"
	if schema == outputSchemaV2 {
		parentField = "parent_id"
	}
	return []string{"id", "title", "path", "kind", "status", "due_on", "repeat_every", "archived_at", parentField, "parent_path", "tags", "file"}
}

func allowedLsTSVFields(schema outputSchema) map[string]struct{} {
	allowed := map[string]struct{}{
		"id": {}, "title": {}, "path": {}, "kind": {}, "status": {}, "due_on": {}, "repeat_every": {},
		"archived_at": {}, "parent_id": {}, "parent_path": {}, "tags": {}, "file": {},
	}
	if schema == outputSchemaV1 {
		allowed["parent"] = struct{}{}
	}
	return allowed
}

func defaultLsTabularFields(schema outputSchema, groupBy string) []string {
	fields := append([]string{}, defaultLsTSVFields(schema)...)
	if strings.TrimSpace(groupBy) == "" {
		return fields
	}
	return append([]string{"group"}, fields...)
}

func allowedLsTabularFields(schema outputSchema, groupBy string) map[string]struct{} {
	allowed := allowedLsTSVFields(schema)
	if strings.TrimSpace(groupBy) != "" {
		allowed["group"] = struct{}{}
	}
	return allowed
}

func defaultNextTSVFields(schema outputSchema) []string {
	parentField := "parent"
	if schema == outputSchemaV2 {
		parentField = "parent_id"
	}
	return []string{"id", "title", "path", "kind", "status", "due_on", "repeat_every", parentField, "parent_path", "tags", "file"}
}

func allowedNextTSVFields(schema outputSchema) map[string]struct{} {
	allowed := map[string]struct{}{
		"id": {}, "title": {}, "path": {}, "kind": {}, "status": {}, "due_on": {}, "repeat_every": {},
		"parent_id": {}, "parent_path": {}, "tags": {}, "file": {},
	}
	if schema == outputSchemaV1 {
		allowed["parent"] = struct{}{}
	}
	return allowed
}

func sortTaskQueryResults(tasks []shelf.Task, byID map[string]shelf.Task, sortBy string, reverse bool) error {
	field := strings.TrimSpace(sortBy)
	if field == "" {
		return nil
	}
	if _, ok := allowedTaskSortFields()[field]; !ok {
		return fmt.Errorf("unknown --sort field: %s", field)
	}
	sort.SliceStable(tasks, func(i, j int) bool {
		order := compareTaskQueryField(tasks[i], tasks[j], byID, field)
		if reverse {
			return order > 0
		}
		return order < 0
	})
	return nil
}

func allowedTaskSortFields() map[string]struct{} {
	return map[string]struct{}{
		"id": {}, "title": {}, "path": {}, "kind": {}, "status": {}, "due_on": {}, "created_at": {}, "updated_at": {},
	}
}

func compareTaskQueryField(a shelf.Task, b shelf.Task, byID map[string]shelf.Task, field string) int {
	switch field {
	case "id":
		return compareTaskStringField(a.ID, b.ID)
	case "title":
		return compareTaskStringFieldWithID(a.Title, b.Title, a.ID, b.ID)
	case "path":
		return compareTaskStringFieldWithID(buildTaskPath(a, byID), buildTaskPath(b, byID), a.ID, b.ID)
	case "kind":
		return compareTaskStringFieldWithID(string(a.Kind), string(b.Kind), a.ID, b.ID)
	case "status":
		return compareTaskStringFieldWithID(string(a.Status), string(b.Status), a.ID, b.ID)
	case "due_on":
		return compareOptionalTaskStringField(a.DueOn, b.DueOn, a.ID, b.ID)
	case "created_at":
		if !a.CreatedAt.Equal(b.CreatedAt) {
			if a.CreatedAt.Before(b.CreatedAt) {
				return -1
			}
			return 1
		}
		return compareTaskStringField(a.ID, b.ID)
	case "updated_at":
		if !a.UpdatedAt.Equal(b.UpdatedAt) {
			if a.UpdatedAt.Before(b.UpdatedAt) {
				return -1
			}
			return 1
		}
		return compareTaskStringField(a.ID, b.ID)
	default:
		return compareTaskStringField(a.ID, b.ID)
	}
}

func compareTaskStringField(a string, b string) int {
	if a != b {
		if a < b {
			return -1
		}
		return 1
	}
	return 0
}

func compareTaskStringFieldWithID(a string, b string, aID string, bID string) int {
	if cmp := compareTaskStringField(a, b); cmp != 0 {
		return cmp
	}
	return compareTaskStringField(aID, bID)
}

func compareOptionalTaskStringField(a string, b string, aID string, bID string) int {
	aBlank := strings.TrimSpace(a) == ""
	bBlank := strings.TrimSpace(b) == ""
	if aBlank != bBlank {
		if aBlank {
			return 1
		}
		return -1
	}
	return compareTaskStringFieldWithID(a, b, aID, bID)
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

func validateCountModeFlags(cmd *cobra.Command, countOnly bool, fields string, header bool, noHeader bool, sortBy string, reverse bool, limit int) error {
	if !countOnly {
		return nil
	}
	if cmd.Flags().Changed("format") {
		return fmt.Errorf("--count cannot be combined with --format")
	}
	if strings.TrimSpace(fields) != "" {
		return fmt.Errorf("--count cannot be combined with --fields")
	}
	if header || noHeader {
		return fmt.Errorf("--count cannot be combined with --header or --no-header")
	}
	if strings.TrimSpace(sortBy) != "" || reverse {
		return fmt.Errorf("--count cannot be combined with --sort or --reverse")
	}
	if cmd.Flags().Changed("limit") && limit != 50 {
		return fmt.Errorf("--count cannot be combined with --limit")
	}
	return nil
}

func validateLsGrouping(format string, groupBy string, countOnly bool) error {
	field := strings.TrimSpace(groupBy)
	if field == "" {
		return nil
	}
	if countOnly {
		return fmt.Errorf("--group-by cannot be combined with --count")
	}
	if format == "kanban" || format == "tree" {
		return fmt.Errorf("--group-by cannot be combined with --format %s", format)
	}
	if _, ok := allowedLsGroupFields()[field]; !ok {
		return fmt.Errorf("unknown --group-by field: %s", field)
	}
	return nil
}

func allowedLsGroupFields() map[string]struct{} {
	return map[string]struct{}{
		"status": {}, "kind": {}, "parent": {},
	}
}

func buildGroupedTaskQueryRecords(rootDir string, tasks []shelf.Task, byID map[string]shelf.Task, groupBy string) []groupedTaskQueryRecord {
	field := strings.TrimSpace(groupBy)
	records := make([]groupedTaskQueryRecord, 0, len(tasks))
	for _, task := range tasks {
		records = append(records, groupedTaskQueryRecord{
			Group:           groupTaskLabel(task, byID, field),
			taskQueryRecord: buildTaskQueryRecord(rootDir, task, byID),
		})
	}
	if field == "" {
		return records
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Group != records[j].Group {
			return records[i].Group < records[j].Group
		}
		return records[i].taskQueryRecord.ID < records[j].taskQueryRecord.ID
	})
	return records
}

func buildGroupedTaskQueryRecordsV2(rootDir string, tasks []shelf.Task, byID map[string]shelf.Task, groupBy string) []groupedTaskQueryRecordV2 {
	field := strings.TrimSpace(groupBy)
	records := make([]groupedTaskQueryRecordV2, 0, len(tasks))
	for _, task := range tasks {
		records = append(records, groupedTaskQueryRecordV2{
			Group:             groupTaskLabel(task, byID, field),
			taskQueryRecordV2: buildTaskQueryRecordV2(rootDir, task, byID),
		})
	}
	if field == "" {
		return records
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Group != records[j].Group {
			return records[i].Group < records[j].Group
		}
		return records[i].taskQueryRecordV2.ID < records[j].taskQueryRecordV2.ID
	})
	return records
}

func groupTaskLabel(task shelf.Task, byID map[string]shelf.Task, field string) string {
	switch field {
	case "status":
		return string(task.Status)
	case "kind":
		return string(task.Kind)
	case "parent":
		if task.Parent == "" {
			return "root"
		}
		parent, ok := byID[task.Parent]
		if !ok {
			return "(missing)"
		}
		return buildTaskPath(parent, byID)
	default:
		return ""
	}
}

func groupedRecordsToAny(records []groupedTaskQueryRecord, groupBy string) any {
	if strings.TrimSpace(groupBy) == "" {
		items := make([]taskQueryRecord, 0, len(records))
		for _, record := range records {
			items = append(items, record.taskQueryRecord)
		}
		return items
	}
	return records
}

func groupedTaskRecordsToAny(schema outputSchema, recordsV1 []groupedTaskQueryRecord, recordsV2 []groupedTaskQueryRecordV2, groupBy string) any {
	if schema == outputSchemaV2 {
		if strings.TrimSpace(groupBy) == "" {
			items := make([]taskQueryRecordV2, 0, len(recordsV2))
			for _, record := range recordsV2 {
				items = append(items, record.taskQueryRecordV2)
			}
			return items
		}
		return recordsV2
	}
	return groupedRecordsToAny(recordsV1, groupBy)
}

func renderGroupedTaskJSONL(schema outputSchema, recordsV1 []groupedTaskQueryRecord, recordsV2 []groupedTaskQueryRecordV2, groupBy string) (string, error) {
	if schema == outputSchemaV2 {
		if strings.TrimSpace(groupBy) == "" {
			items := make([]taskQueryRecordV2, 0, len(recordsV2))
			for _, record := range recordsV2 {
				items = append(items, record.taskQueryRecordV2)
			}
			return renderJSONL(items)
		}
		return renderJSONL(recordsV2)
	}
	if strings.TrimSpace(groupBy) == "" {
		items := make([]taskQueryRecord, 0, len(recordsV1))
		for _, record := range recordsV1 {
			items = append(items, record.taskQueryRecord)
		}
		return renderJSONL(items)
	}
	return renderJSONL(recordsV1)
}

func buildNextJSONItems(schema outputSchema, rootDir string, tasks []shelf.Task, readiness map[string]shelf.TaskReadiness, byID map[string]shelf.Task, limit int) any {
	switch schema {
	case outputSchemaV2:
		items := make([]taskQueryRecordV2, 0)
		for _, task := range tasks {
			info, ok := readiness[task.ID]
			if !ok || !info.Ready {
				continue
			}
			items = append(items, buildTaskQueryRecordV2(rootDir, task, byID))
			if limit > 0 && len(items) >= limit {
				break
			}
		}
		return items
	default:
		items := make([]taskQueryRecord, 0)
		for _, task := range tasks {
			info, ok := readiness[task.ID]
			if !ok || !info.Ready {
				continue
			}
			items = append(items, buildTaskQueryRecord(rootDir, task, byID))
			if limit > 0 && len(items) >= limit {
				break
			}
		}
		return items
	}
}

func renderNextJSONL(schema outputSchema, rootDir string, tasks []shelf.Task, readiness map[string]shelf.TaskReadiness, byID map[string]shelf.Task, limit int) (string, error) {
	switch items := buildNextJSONItems(schema, rootDir, tasks, readiness, byID, limit).(type) {
	case []taskQueryRecordV2:
		return renderJSONL(items)
	default:
		return renderJSONL(items.([]taskQueryRecord))
	}
}

func printNextTSV(schema outputSchema, selectedFields []string, rootDir string, tasks []shelf.Task, readiness map[string]shelf.TaskReadiness, byID map[string]shelf.Task, limit int) {
	count := 0
	for _, task := range tasks {
		info, ok := readiness[task.ID]
		if !ok || !info.Ready {
			continue
		}
		switch schema {
		case outputSchemaV2:
			fmt.Println(joinTSVFields(selectedFields, buildTaskQueryRecordV2(rootDir, task, byID).TSVFields()))
		default:
			fmt.Println(joinTSVFields(selectedFields, buildTaskQueryRecord(rootDir, task, byID).TSVFields()))
		}
		count++
		if limit > 0 && count >= limit {
			break
		}
	}
}

func renderNextCSV(schema outputSchema, selectedFields []string, includeHeader bool, rootDir string, tasks []shelf.Task, readiness map[string]shelf.TaskReadiness, byID map[string]shelf.Task, limit int) (string, error) {
	switch items := buildNextJSONItems(schema, rootDir, tasks, readiness, byID, limit).(type) {
	case []taskQueryRecordV2:
		return renderCSV(items, selectedFields, includeHeader)
	default:
		return renderCSV(items.([]taskQueryRecord), selectedFields, includeHeader)
	}
}

func printGroupedLsRecords(records []groupedTaskQueryRecord, format string, showID bool) {
	if len(records) == 0 {
		fmt.Println(uiMuted("(none)"))
		return
	}
	currentGroup := ""
	for _, record := range records {
		if record.Group != currentGroup {
			currentGroup = record.Group
			fmt.Println(uiHeading(currentGroup + ":"))
		}
		task := shelf.Task{
			ID:          record.taskQueryRecord.ID,
			Title:       record.taskQueryRecord.Title,
			Kind:        shelf.Kind(record.taskQueryRecord.Kind),
			Status:      shelf.Status(record.taskQueryRecord.Status),
			Tags:        append([]string{}, record.taskQueryRecord.Tags...),
			DueOn:       record.taskQueryRecord.DueOn,
			RepeatEvery: record.taskQueryRecord.RepeatEvery,
			ArchivedAt:  record.taskQueryRecord.ArchivedAt,
			Parent:      record.taskQueryRecord.Parent,
		}
		parentLabel := uiMuted("root")
		if record.taskQueryRecord.Parent != "" {
			if strings.TrimSpace(record.taskQueryRecord.ParentPath) != "" {
				parentLabel = record.taskQueryRecord.ParentPath
			} else {
				parentLabel = uiMuted("(missing)")
			}
		}
		label := uiPrimary(task.Title)
		if showID {
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
			fmt.Printf("  %s kind=%s status=%s tags=%s due=%s repeat=%s archived_at=%q parent=%s\n", label, uiKind(task.Kind), uiStatus(task.Status), formatTagSummary(task.Tags), uiDue(task.DueOn), repeatText, task.ArchivedAt, parentLabel)
			continue
		}
		fmt.Printf("  %s  (%s/%s)%s%s%s parent=%s\n", label, uiKind(task.Kind), uiStatus(task.Status), dueText, tagText, archivedText, parentLabel)
	}
}

func printCountResult(count int, asJSON bool) error {
	if asJSON {
		data, err := json.MarshalIndent(map[string]int{"count": count}, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	fmt.Println(count)
	return nil
}
