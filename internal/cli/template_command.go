package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newTemplateCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage reusable task tree templates",
	}
	cmd.AddCommand(newTemplateListCommand(ctx))
	cmd.AddCommand(newTemplateSaveCommand(ctx))
	cmd.AddCommand(newTemplateShowCommand(ctx))
	cmd.AddCommand(newTemplateApplyCommand(ctx))
	cmd.AddCommand(newTemplateDeleteCommand(ctx))
	return cmd
}

func newTemplateListCommand(ctx *commandContext) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List template names",
		RunE: func(_ *cobra.Command, _ []string) error {
			names, err := shelf.ListTemplateNames(ctx.rootDir)
			if err != nil {
				return err
			}
			if asJSON {
				data, err := json.MarshalIndent(names, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}
			if len(names) == 0 {
				fmt.Println("(none)")
				return nil
			}
			for _, name := range names {
				fmt.Println(name)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newTemplateSaveCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save <name> <id>",
		Short: "Save a task subtree as a template",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			id := strings.TrimSpace(args[1])
			tpl, err := shelf.BuildTemplateFromTask(ctx.rootDir, name, id)
			if err != nil {
				return err
			}
			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "template-save"); err != nil {
					return err
				}
				if err := shelf.SaveTemplate(ctx.rootDir, tpl); err != nil {
					return err
				}
				fmt.Printf("Saved template: %s\n", name)
				return nil
			})
		},
	}
	return cmd
}

func newTemplateShowCommand(ctx *commandContext) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show template contents",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			tpl, err := shelf.LoadTemplate(ctx.rootDir, strings.TrimSpace(args[0]))
			if err != nil {
				return err
			}
			if asJSON {
				data, err := json.MarshalIndent(tpl, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}
			fmt.Printf("name: %s\n", tpl.Name)
			for _, item := range tpl.Tasks {
				parent := item.ParentKey
				if parent == "" {
					parent = "root"
				}
				fmt.Printf("- %s (%s/%s) parent=%s\n", item.Title, item.Kind, item.Status, parent)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newTemplateApplyCommand(ctx *commandContext) *cobra.Command {
	var parent string
	var titlePrefix string
	cmd := &cobra.Command{
		Use:   "apply <name>",
		Short: "Expand a template into real tasks",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			tpl, err := shelf.LoadTemplate(ctx.rootDir, strings.TrimSpace(args[0]))
			if err != nil {
				return err
			}
			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "template-apply"); err != nil {
					return err
				}
				created, err := shelf.ApplyTemplate(ctx.rootDir, tpl, parent, titlePrefix)
				if err != nil {
					return err
				}
				fmt.Printf("Applied template: %s (%d tasks)\n", tpl.Name, len(created))
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&parent, "parent", "", "Parent task ID or root")
	cmd.Flags().StringVar(&titlePrefix, "title-prefix", "", "Prefix added to created task titles")
	return cmd
}

func newTemplateDeleteCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm"},
		Short:   "Delete a template",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "template-delete"); err != nil {
					return err
				}
				if err := shelf.DeleteTemplate(ctx.rootDir, name); err != nil {
					return err
				}
				fmt.Printf("Deleted template: %s\n", name)
				return nil
			})
		},
	}
	return cmd
}
