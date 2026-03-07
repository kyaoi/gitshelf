package cli

import (
	"errors"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

type boardColumn struct {
	Status shelf.Status
	Tasks  []shelf.Task
}

func newBoardCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "board",
		Aliases: []string{"kb"},
		Short:   "Open the Daily Cockpit in board mode",
		Example: "  shelf board\n" +
			"  shelf board --show-id",
		RunE: func(_ *cobra.Command, _ []string) error {
			if !dailyCockpitIsTTY() {
				return errors.New("board はTTYが必要です")
			}
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			startDate, dayCount, err := resolveDailyCockpitRange(ctx.rootDir)
			if err != nil {
				return err
			}
			filter := shelf.TaskFilter{
				Statuses: append([]shelf.Status{}, cfg.Statuses...),
				Limit:    0,
			}
			return runCalendarModeTUIFn(ctx.rootDir, startDate, dayCount, filter.Statuses, calendarTUIOptions{
				Mode:   calendarModeBoard,
				ShowID: ctx.showID,
				Filter: filter,
			})
		},
	}
	return cmd
}

func buildBoardColumns(statuses []shelf.Status, tasks []shelf.Task) []boardColumn {
	grouped := map[shelf.Status][]shelf.Task{}
	for _, task := range tasks {
		grouped[task.Status] = append(grouped[task.Status], task)
	}
	columns := make([]boardColumn, 0, len(statuses))
	for _, status := range statuses {
		columns = append(columns, boardColumn{
			Status: status,
			Tasks:  grouped[status],
		})
	}
	return columns
}
