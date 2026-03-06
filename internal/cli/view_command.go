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
			if err := shelf.SaveConfig(ctx.rootDir, cfg); err != nil {
				return err
			}
			fmt.Printf("Deleted view: %s\n", name)
			return nil
		},
	}
	return cmd
}
