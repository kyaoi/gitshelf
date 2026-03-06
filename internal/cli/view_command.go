package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newViewCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Manage saved task views",
	}
	cmd.AddCommand(newViewListCommand(ctx))
	cmd.AddCommand(newViewShowCommand(ctx))
	cmd.AddCommand(newViewSetCommand(ctx))
	cmd.AddCommand(newViewCopyCommand(ctx))
	cmd.AddCommand(newViewRenameCommand(ctx))
	cmd.AddCommand(newViewMergeCommand(ctx))
	cmd.AddCommand(newViewDeleteCommand(ctx))
	return cmd
}

func newViewListCommand(ctx *commandContext) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List built-in and custom views",
		Example: "  shelf view list\n  shelf view list --json",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			custom := make([]string, 0, len(cfg.Views))
			for name := range cfg.Views {
				custom = append(custom, name)
			}
			sort.Strings(custom)
			builtin := []string{"active", "ready", "blocked", "overdue"}

			if asJSON {
				payload := map[string]any{
					"builtin": builtin,
					"custom":  custom,
				}
				data, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Println(uiHeading("Built-in:"))
			for _, name := range builtin {
				fmt.Printf("  - %s\n", name)
			}
			fmt.Println(uiHeading("Custom:"))
			if len(custom) == 0 {
				fmt.Println(uiMuted("  (none)"))
				return nil
			}
			for _, name := range custom {
				fmt.Printf("  - %s\n", name)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newViewShowCommand(ctx *commandContext) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "show <name>",
		Short:   "Show view filter details",
		Example: "  shelf view show active\n  shelf view show myview --json",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			filter, err := resolveTaskView(ctx.rootDir, name)
			if err != nil {
				return err
			}

			if asJSON {
				payload := map[string]any{
					"name":   name,
					"filter": filter,
				}
				data, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("name: %s\n", name)
			fmt.Printf("kind: %v\n", filter.Kinds)
			fmt.Printf("status: %v\n", filter.Statuses)
			fmt.Printf("not-kind: %v\n", filter.NotKinds)
			fmt.Printf("not-status: %v\n", filter.NotStatuses)
			fmt.Printf("ready: %t\n", filter.ReadyOnly)
			fmt.Printf("blocked-by-deps: %t\n", filter.DepsBlocked)
			fmt.Printf("due-before: %s\n", filter.DueBefore)
			fmt.Printf("due-after: %s\n", filter.DueAfter)
			fmt.Printf("overdue: %t\n", filter.Overdue)
			fmt.Printf("no-due: %t\n", filter.NoDue)
			fmt.Printf("parent: %s\n", filter.Parent)
			fmt.Printf("search: %s\n", filter.Search)
			fmt.Printf("limit: %d\n", filter.Limit)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newViewSetCommand(ctx *commandContext) *cobra.Command {
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
		parent      string
		search      string
		limit       int
	)

	cmd := &cobra.Command{
		Use:   "set <name>",
		Short: "Create or replace a custom view",
		Example: "  shelf view set active_todos --kind todo --not-status done --not-status cancelled\n" +
			"  shelf view set focus --ready --limit 20",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return fmt.Errorf("view name is required")
			}
			if _, ok := builtinTaskView(name); ok {
				return fmt.Errorf("cannot overwrite built-in view: %s", name)
			}
			if !cmd.Flags().Changed("kind") &&
				!cmd.Flags().Changed("status") &&
				!cmd.Flags().Changed("not-kind") &&
				!cmd.Flags().Changed("not-status") &&
				!cmd.Flags().Changed("ready") &&
				!cmd.Flags().Changed("blocked-by-deps") &&
				!cmd.Flags().Changed("due-before") &&
				!cmd.Flags().Changed("due-after") &&
				!cmd.Flags().Changed("overdue") &&
				!cmd.Flags().Changed("no-due") &&
				!cmd.Flags().Changed("parent") &&
				!cmd.Flags().Changed("search") &&
				!cmd.Flags().Changed("limit") {
				return fmt.Errorf("at least one filter flag is required")
			}

			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			cfg.Views[name] = shelf.TaskView{
				Kinds:       toKinds(kinds),
				Statuses:    toStatuses(statuses),
				NotKinds:    toKinds(notKinds),
				NotStatuses: toStatuses(notStatuses),
				ReadyOnly:   ready,
				DepsBlocked: depsBlocked,
				DueBefore:   strings.TrimSpace(dueBefore),
				DueAfter:    strings.TrimSpace(dueAfter),
				Overdue:     overdue,
				NoDue:       noDue,
				Parent:      strings.TrimSpace(parent),
				Search:      strings.TrimSpace(search),
				Limit:       limit,
			}
			if err := prepareUndoSnapshot(ctx.rootDir, "view-set"); err != nil {
				return err
			}
			if err := shelf.SaveConfig(ctx.rootDir, cfg); err != nil {
				return err
			}
			fmt.Printf("Saved view: %s\n", name)
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
	cmd.Flags().StringVar(&parent, "parent", "", "Filter by parent task ID or root")
	cmd.Flags().StringVar(&search, "search", "", "Search by title/body")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of items")
	return cmd
}

func newViewDeleteCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm"},
		Short:   "Delete a custom view",
		Example: "  shelf view delete myview",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if _, ok := builtinTaskView(name); ok {
				return fmt.Errorf("cannot delete built-in view: %s", name)
			}
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			if _, ok := cfg.Views[name]; !ok {
				return fmt.Errorf("unknown custom view: %s", name)
			}
			delete(cfg.Views, name)
			if err := prepareUndoSnapshot(ctx.rootDir, "view-delete"); err != nil {
				return err
			}
			if err := shelf.SaveConfig(ctx.rootDir, cfg); err != nil {
				return err
			}
			fmt.Printf("Deleted view: %s\n", name)
			return nil
		},
	}
	return cmd
}

func newViewCopyCommand(ctx *commandContext) *cobra.Command {
	return &cobra.Command{
		Use:     "copy <src> <dst>",
		Short:   "Copy a custom/built-in view to a custom view name",
		Example: "  shelf view copy active active_copy",
		Args:    cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			src := strings.TrimSpace(args[0])
			dst := strings.TrimSpace(args[1])
			if dst == "" {
				return fmt.Errorf("destination view name is required")
			}
			if _, ok := builtinTaskView(dst); ok {
				return fmt.Errorf("cannot overwrite built-in view: %s", dst)
			}
			srcView, err := resolveViewAsTaskView(ctx.rootDir, src)
			if err != nil {
				return err
			}
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			cfg.Views[dst] = srcView
			if err := prepareUndoSnapshot(ctx.rootDir, "view-copy"); err != nil {
				return err
			}
			if err := shelf.SaveConfig(ctx.rootDir, cfg); err != nil {
				return err
			}
			fmt.Printf("Copied view: %s -> %s\n", src, dst)
			return nil
		},
	}
}

func newViewRenameCommand(ctx *commandContext) *cobra.Command {
	return &cobra.Command{
		Use:     "rename <src> <dst>",
		Short:   "Rename a custom view",
		Example: "  shelf view rename focus today_focus",
		Args:    cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			src := strings.TrimSpace(args[0])
			dst := strings.TrimSpace(args[1])
			if _, ok := builtinTaskView(src); ok {
				return fmt.Errorf("cannot rename built-in view: %s", src)
			}
			if _, ok := builtinTaskView(dst); ok {
				return fmt.Errorf("cannot overwrite built-in view: %s", dst)
			}
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			view, ok := cfg.Views[src]
			if !ok {
				return fmt.Errorf("unknown custom view: %s", src)
			}
			delete(cfg.Views, src)
			cfg.Views[dst] = view
			if err := prepareUndoSnapshot(ctx.rootDir, "view-rename"); err != nil {
				return err
			}
			if err := shelf.SaveConfig(ctx.rootDir, cfg); err != nil {
				return err
			}
			fmt.Printf("Renamed view: %s -> %s\n", src, dst)
			return nil
		},
	}
}

func newViewMergeCommand(ctx *commandContext) *cobra.Command {
	var (
		from     []string
		strategy string
	)
	cmd := &cobra.Command{
		Use:   "merge <dst>",
		Short: "Merge multiple views into one custom view",
		Example: "  shelf view merge focus --from active --from ready\n" +
			"  shelf view merge active_ready --from active --from ready --strategy union",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			dst := strings.TrimSpace(args[0])
			if dst == "" {
				return fmt.Errorf("destination view name is required")
			}
			if _, ok := builtinTaskView(dst); ok {
				return fmt.Errorf("cannot overwrite built-in view: %s", dst)
			}
			if len(from) < 2 {
				return fmt.Errorf("--from must be specified at least 2 times")
			}
			switch strategy {
			case "overlay", "union":
			default:
				return fmt.Errorf("invalid --strategy: %s (allowed: overlay|union)", strategy)
			}

			views := make([]shelf.TaskView, 0, len(from))
			for _, src := range from {
				view, err := resolveViewAsTaskView(ctx.rootDir, src)
				if err != nil {
					return err
				}
				views = append(views, view)
			}

			merged := views[0]
			for i := 1; i < len(views); i++ {
				if strategy == "overlay" {
					merged = overlayTaskView(merged, views[i])
				} else {
					merged = unionTaskView(merged, views[i])
				}
			}

			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			cfg.Views[dst] = merged
			if err := prepareUndoSnapshot(ctx.rootDir, "view-merge"); err != nil {
				return err
			}
			if err := shelf.SaveConfig(ctx.rootDir, cfg); err != nil {
				return err
			}
			fmt.Printf("Merged view: %s (%s)\n", dst, strategy)
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&from, "from", nil, "Source view name (repeatable)")
	cmd.Flags().StringVar(&strategy, "strategy", "overlay", "Merge strategy: overlay|union")
	return cmd
}

func resolveViewAsTaskView(rootDir string, name string) (shelf.TaskView, error) {
	filter, err := resolveTaskView(rootDir, name)
	if err != nil {
		return shelf.TaskView{}, err
	}
	return shelf.TaskView{
		Kinds:       filter.Kinds,
		Statuses:    filter.Statuses,
		NotKinds:    filter.NotKinds,
		NotStatuses: filter.NotStatuses,
		ReadyOnly:   filter.ReadyOnly,
		DepsBlocked: filter.DepsBlocked,
		DueBefore:   filter.DueBefore,
		DueAfter:    filter.DueAfter,
		Overdue:     filter.Overdue,
		NoDue:       filter.NoDue,
		Parent:      filter.Parent,
		Search:      filter.Search,
		Limit:       filter.Limit,
	}, nil
}

func overlayTaskView(base shelf.TaskView, next shelf.TaskView) shelf.TaskView {
	out := base
	if len(next.Kinds) > 0 {
		out.Kinds = next.Kinds
	}
	if len(next.Statuses) > 0 {
		out.Statuses = next.Statuses
	}
	if len(next.NotKinds) > 0 {
		out.NotKinds = next.NotKinds
	}
	if len(next.NotStatuses) > 0 {
		out.NotStatuses = next.NotStatuses
	}
	out.ReadyOnly = out.ReadyOnly || next.ReadyOnly
	out.DepsBlocked = out.DepsBlocked || next.DepsBlocked
	if strings.TrimSpace(next.DueBefore) != "" {
		out.DueBefore = next.DueBefore
	}
	if strings.TrimSpace(next.DueAfter) != "" {
		out.DueAfter = next.DueAfter
	}
	out.Overdue = out.Overdue || next.Overdue
	out.NoDue = out.NoDue || next.NoDue
	if strings.TrimSpace(next.Parent) != "" {
		out.Parent = next.Parent
	}
	if strings.TrimSpace(next.Search) != "" {
		out.Search = next.Search
	}
	if next.Limit > 0 {
		out.Limit = next.Limit
	}
	return out
}

func unionTaskView(base shelf.TaskView, next shelf.TaskView) shelf.TaskView {
	out := shelf.TaskView{
		Kinds:       uniqueKinds(append(slicesCloneKinds(base.Kinds), next.Kinds...)),
		Statuses:    uniqueStatuses(append(slicesCloneStatuses(base.Statuses), next.Statuses...)),
		NotKinds:    uniqueKinds(append(slicesCloneKinds(base.NotKinds), next.NotKinds...)),
		NotStatuses: uniqueStatuses(append(slicesCloneStatuses(base.NotStatuses), next.NotStatuses...)),
		ReadyOnly:   base.ReadyOnly || next.ReadyOnly,
		DepsBlocked: base.DepsBlocked || next.DepsBlocked,
		Overdue:     base.Overdue || next.Overdue,
		NoDue:       base.NoDue || next.NoDue,
		Limit:       max(base.Limit, next.Limit),
	}
	if strings.TrimSpace(base.DueBefore) == "" {
		out.DueBefore = next.DueBefore
	} else if strings.TrimSpace(next.DueBefore) == "" {
		out.DueBefore = base.DueBefore
	} else if next.DueBefore > base.DueBefore {
		out.DueBefore = next.DueBefore
	} else {
		out.DueBefore = base.DueBefore
	}
	if strings.TrimSpace(base.DueAfter) == "" {
		out.DueAfter = next.DueAfter
	} else if strings.TrimSpace(next.DueAfter) == "" {
		out.DueAfter = base.DueAfter
	} else if next.DueAfter < base.DueAfter {
		out.DueAfter = next.DueAfter
	} else {
		out.DueAfter = base.DueAfter
	}
	out.Parent = strings.TrimSpace(base.Parent)
	if strings.TrimSpace(next.Parent) != "" {
		out.Parent = strings.TrimSpace(next.Parent)
	}
	out.Search = strings.TrimSpace(base.Search)
	if strings.TrimSpace(next.Search) != "" {
		out.Search = strings.TrimSpace(next.Search)
	}
	return out
}

func slicesCloneKinds(values []shelf.Kind) []shelf.Kind {
	out := make([]shelf.Kind, len(values))
	copy(out, values)
	return out
}

func slicesCloneStatuses(values []shelf.Status) []shelf.Status {
	out := make([]shelf.Status, len(values))
	copy(out, values)
	return out
}

func uniqueKinds(values []shelf.Kind) []shelf.Kind {
	seen := map[shelf.Kind]struct{}{}
	out := make([]shelf.Kind, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func uniqueStatuses(values []shelf.Status) []shelf.Status {
	seen := map[shelf.Status]struct{}{}
	out := make([]shelf.Status, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
