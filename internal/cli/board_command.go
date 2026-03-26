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
	var flags cockpitLaunchFlags
	cmd := &cobra.Command{
		Use:     "board",
		Aliases: []string{"kb"},
		Short:   "Open Cockpit in board mode",
		Example: "  shelf board\n" +
			"  shelf board --show-id",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !dailyCockpitIsTTY() {
				return errors.New("board requires a TTY")
			}
			return runCockpitLaunch(ctx, cmd, calendarModeBoard, flags)
		},
	}
	addCockpitLaunchFlags(cmd, &flags)
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
