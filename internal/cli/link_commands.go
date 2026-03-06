package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newLinkCommand(ctx *commandContext) *cobra.Command {
	var (
		from string
		to   string
		kind string
	)

	cmd := &cobra.Command{
		Use:   "link",
		Short: "Create outbound link",
		RunE: func(_ *cobra.Command, _ []string) error {
			if strings.TrimSpace(from) == "" || strings.TrimSpace(to) == "" || strings.TrimSpace(kind) == "" {
				var err error
				from, to, kind, err = resolveLinkInputInteractive(ctx)
				if err != nil {
					return err
				}
			}

			if err := shelf.LinkTasks(ctx.rootDir, from, to, shelf.LinkType(kind)); err != nil {
				return err
			}
			fmt.Printf("Linked: %s --%s--> %s\n", uiShortID(shelf.ShortID(from)), uiLinkType(shelf.LinkType(kind)), uiShortID(shelf.ShortID(to)))
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Source task ID")
	cmd.Flags().StringVar(&to, "to", "", "Destination task ID")
	cmd.Flags().StringVar(&kind, "type", "", "Link type")
	return cmd
}

func newUnlinkCommand(ctx *commandContext) *cobra.Command {
	var (
		from string
		to   string
		kind string
	)
	cmd := &cobra.Command{
		Use:   "unlink",
		Short: "Remove outbound link",
		RunE: func(_ *cobra.Command, _ []string) error {
			if strings.TrimSpace(from) == "" || strings.TrimSpace(to) == "" || strings.TrimSpace(kind) == "" {
				var err error
				from, to, kind, err = resolveUnlinkInputInteractive(ctx)
				if err != nil {
					return err
				}
			}

			removed, err := shelf.UnlinkTasks(ctx.rootDir, from, to, shelf.LinkType(kind))
			if err != nil {
				return err
			}
			if !removed {
				return errors.New("指定リンクは存在しません")
			}
			fmt.Printf("Unlinked: %s --%s--> %s\n", uiShortID(shelf.ShortID(from)), uiLinkType(shelf.LinkType(kind)), uiShortID(shelf.ShortID(to)))
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Source task ID")
	cmd.Flags().StringVar(&to, "to", "", "Destination task ID")
	cmd.Flags().StringVar(&kind, "type", "", "Link type")
	return cmd
}

func newLinksCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "links <id>",
		Short: "Show outbound and inbound links",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "リンクを表示するタスクを選択", nil, true)
			if err != nil {
				return err
			}
			outbound, inbound, err := shelf.ListLinks(ctx.rootDir, id)
			if err != nil {
				return err
			}

			fmt.Println(uiHeading("Outbound:"))
			if len(outbound) == 0 {
				fmt.Println(uiMuted("  (none)"))
			}
			for _, edge := range outbound {
				fmt.Printf("  %s --%s--> %s\n", uiShortID(shelf.ShortID(id)), uiLinkType(edge.Type), uiShortID(shelf.ShortID(edge.To)))
			}

			fmt.Println(uiHeading("Inbound:"))
			if len(inbound) == 0 {
				fmt.Println(uiMuted("  (none)"))
			}
			for _, edge := range inbound {
				fmt.Printf("  %s --%s--> %s\n", uiShortID(shelf.ShortID(edge.From)), uiLinkType(edge.Type), uiShortID(shelf.ShortID(id)))
			}

			fmt.Println("depends_on の向き: A depends_on B = AをやるにはBが先")
			return nil
		},
	}
	return cmd
}

func resolveLinkInputInteractive(ctx *commandContext) (string, string, string, error) {
	if !interactive.IsTTY() {
		return "", "", "", errors.New("非TTYでは対話入力できません。--from --to --type を指定してください")
	}

	taskStore := shelf.NewTaskStore(ctx.rootDir)
	tasks, err := taskStore.List()
	if err != nil {
		return "", "", "", err
	}
	if len(tasks) == 0 {
		return "", "", "", errors.New("タスクがありません")
	}

	taskOptions := buildTaskSelectionOptions(tasks, taskSelectionBuildOptions{
		Hierarchical:  true,
		ShowID:        ctx.showID,
		PreviewBody:   ctx.previewBody,
		IncludeOrphan: true,
	})

	src, err := interactive.Select("source を選択", taskOptions)
	if err != nil {
		return "", "", "", err
	}

	dstTasks := make([]shelf.Task, 0, len(tasks)-1)
	for _, task := range tasks {
		if task.ID == src.Value {
			continue
		}
		dstTasks = append(dstTasks, task)
	}
	dstOptions := buildTaskSelectionOptions(dstTasks, taskSelectionBuildOptions{
		Hierarchical:  true,
		ShowID:        ctx.showID,
		PreviewBody:   ctx.previewBody,
		IncludeOrphan: true,
	})
	dst, err := interactive.Select("destination を選択", dstOptions)
	if err != nil {
		return "", "", "", err
	}

	cfg, err := shelf.LoadConfig(ctx.rootDir)
	if err != nil {
		return "", "", "", err
	}
	typeOptions := make([]interactive.Option, 0, len(cfg.LinkTypes))
	for _, linkType := range cfg.LinkTypes {
		typeOptions = append(typeOptions, interactive.Option{
			Value:      string(linkType),
			Label:      string(linkType),
			SearchText: string(linkType),
		})
	}
	selectedType, err := interactive.Select("type を選択（A depends_on B = AをやるにはBが先）", typeOptions)
	if err != nil {
		return "", "", "", err
	}
	return src.Value, dst.Value, selectedType.Value, nil
}

func resolveUnlinkInputInteractive(ctx *commandContext) (string, string, string, error) {
	if !interactive.IsTTY() {
		return "", "", "", errors.New("非TTYでは対話入力できません。--from --to --type を指定してください")
	}

	taskStore := shelf.NewTaskStore(ctx.rootDir)
	tasks, err := taskStore.List()
	if err != nil {
		return "", "", "", err
	}
	if len(tasks) == 0 {
		return "", "", "", errors.New("タスクがありません")
	}

	taskOptions := buildTaskSelectionOptions(tasks, taskSelectionBuildOptions{
		Hierarchical:  true,
		ShowID:        ctx.showID,
		PreviewBody:   ctx.previewBody,
		IncludeOrphan: true,
	})
	byID := make(map[string]shelf.Task, len(tasks))
	for _, task := range tasks {
		byID[task.ID] = task
	}

	src, err := interactive.Select("source を選択", taskOptions)
	if err != nil {
		return "", "", "", err
	}

	edgeStore := shelf.NewEdgeStore(ctx.rootDir)
	edges, err := edgeStore.ListOutbound(src.Value)
	if err != nil {
		return "", "", "", err
	}
	if len(edges) == 0 {
		return "", "", "", errors.New("選択した source には outbound link がありません")
	}

	edgeOptions := make([]interactive.Option, 0, len(edges))
	for _, edge := range edges {
		dstLabel := edge.To
		if task, ok := byID[edge.To]; ok {
			dstLabel = task.Title
		}
		label := fmt.Sprintf("%s --%s--> %s", src.Label, edge.Type, dstLabel)
		if ctx.showID {
			label = fmt.Sprintf("[%s] --%s--> [%s]", shelf.ShortID(src.Value), edge.Type, shelf.ShortID(edge.To))
		}
		edgeOptions = append(edgeOptions, interactive.Option{
			Value:      fmt.Sprintf("%s\t%s", edge.To, edge.Type),
			Label:      label,
			SearchText: fmt.Sprintf("%s %s %s", edge.To, shelf.ShortID(edge.To), edge.Type),
		})
	}

	selectedEdge, err := interactive.Select("削除する edge を選択", edgeOptions)
	if err != nil {
		return "", "", "", err
	}
	parts := strings.SplitN(selectedEdge.Value, "\t", 2)
	if len(parts) != 2 {
		return "", "", "", errors.New("edge 値の解析に失敗しました")
	}

	return src.Value, parts[0], parts[1], nil
}
