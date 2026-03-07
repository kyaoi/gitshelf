package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTodayCommand(ctx *commandContext) *cobra.Command {
	var flags cockpitLaunchFlags

	cmd := &cobra.Command{
		Use:     "now",
		Aliases: []string{"today", "td", "nw"},
		Short:   "Open Cockpit in the Now view",
		Example: "  shelf now\n" +
			"  shelf now --limit 10\n" +
			"  shelf now --status open --status blocked",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !dailyCockpitIsTTY() {
				return fmt.Errorf("now はTTYが必要です")
			}
			return runCockpitLaunch(ctx, cmd, calendarModeNow, flags)
		},
	}
	addCockpitLaunchFlags(cmd, &flags)
	return cmd
}
