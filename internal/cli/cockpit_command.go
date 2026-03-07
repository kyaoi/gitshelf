package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

type cockpitLaunchFlags struct {
	start     string
	days      int
	months    int
	years     int
	limit     int
	kinds     []string
	statuses  []string
	tags      []string
	notKinds  []string
	notStatus []string
	notTags   []string
}

func newCockpitCommand(ctx *commandContext) *cobra.Command {
	var (
		mode  string
		flags cockpitLaunchFlags
	)

	cmd := &cobra.Command{
		Use:     "cockpit",
		Aliases: []string{"cp"},
		Short:   "Open the main Cockpit workspace TUI",
		Example: "  shelf cockpit\n" +
			"  shelf cockpit --mode tree\n" +
			"  shelf cockpit --mode board --months 3\n" +
			"  shelf cockpit --mode review --status open --status blocked",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !dailyCockpitIsTTY() {
				return errors.New("cockpit はTTYが必要です")
			}
			parsedMode, err := parseCockpitMode(mode)
			if err != nil {
				return err
			}
			return runCockpitLaunch(ctx, cmd, parsedMode, flags)
		},
	}

	cmd.Flags().StringVar(&mode, "mode", "calendar", "Starting mode: calendar|tree|board|review|now")
	addCockpitLaunchFlags(cmd, &flags)
	return cmd
}

func addCockpitLaunchFlags(cmd *cobra.Command, flags *cockpitLaunchFlags) {
	cmd.Flags().StringVar(&flags.start, "start", "", "Anchor date (YYYY-MM-DD|today|tomorrow). Defaults to current week Monday")
	cmd.Flags().IntVar(&flags.days, "days", 0, "Render an explicit day range")
	cmd.Flags().IntVar(&flags.months, "months", 0, "Render an explicit month range from the month containing --start")
	cmd.Flags().IntVar(&flags.years, "years", 0, "Render an explicit year range from the year containing --start")
	cmd.Flags().IntVar(&flags.limit, "limit", 0, "Maximum items per non-focused section (0 means unlimited)")
	cmd.Flags().StringArrayVar(&flags.kinds, "kind", nil, "Include kind (repeatable)")
	cmd.Flags().StringArrayVar(&flags.statuses, "status", nil, "Include status (repeatable)")
	cmd.Flags().StringArrayVar(&flags.tags, "tag", nil, "Include tag (repeatable)")
	cmd.Flags().StringArrayVar(&flags.notKinds, "not-kind", nil, "Exclude kind (repeatable)")
	cmd.Flags().StringArrayVar(&flags.notStatus, "not-status", nil, "Exclude status (repeatable)")
	cmd.Flags().StringArrayVar(&flags.notTags, "not-tag", nil, "Exclude tag (repeatable)")
}

func runCockpitLaunch(ctx *commandContext, cmd *cobra.Command, mode calendarMode, flags cockpitLaunchFlags) error {
	cfg, err := shelf.LoadConfig(ctx.rootDir)
	if err != nil {
		return err
	}
	for _, kind := range toKinds(flags.kinds) {
		if err := cfg.ValidateKind(kind); err != nil {
			return err
		}
	}
	for _, kind := range toKinds(flags.notKinds) {
		if err := cfg.ValidateKind(kind); err != nil {
			return err
		}
	}
	for _, status := range toStatuses(flags.statuses) {
		if err := cfg.ValidateStatus(status); err != nil {
			return err
		}
	}
	for _, status := range toStatuses(flags.notStatus) {
		if err := cfg.ValidateStatus(status); err != nil {
			return err
		}
	}
	for _, tag := range parseTagFlagValues(flags.tags) {
		if err := cfg.ValidateTag(tag); err != nil {
			return err
		}
	}
	for _, tag := range parseTagFlagValues(flags.notTags) {
		if err := cfg.ValidateTag(tag); err != nil {
			return err
		}
	}

	startDate, err := resolveCalendarStart(flags.start)
	if err != nil {
		return err
	}
	rangeStart, dayCount, err := resolveCalendarRange(
		startDate,
		flags.days,
		flags.months,
		flags.years,
		cfg.Commands.Calendar,
		cmd.Flags().Changed("days"),
		cmd.Flags().Changed("months"),
		cmd.Flags().Changed("years"),
	)
	if err != nil {
		return err
	}

	filter := shelf.TaskFilter{
		Kinds:       toKinds(flags.kinds),
		Statuses:    toStatuses(flags.statuses),
		Tags:        parseTagFlagValues(flags.tags),
		NotKinds:    toKinds(flags.notKinds),
		NotStatuses: toStatuses(flags.notStatus),
		NotTags:     parseTagFlagValues(flags.notTags),
		Limit:       0,
	}
	defaultStatuses := defaultCockpitStatuses(mode, cfg)
	return runCalendarModeTUIFn(ctx.rootDir, rangeStart, dayCount, defaultStatuses, calendarTUIOptions{
		Mode:         mode,
		ShowID:       ctx.showID,
		SectionLimit: flags.limit,
		Filter:       filter,
	})
}

func parseCockpitMode(value string) (calendarMode, error) {
	switch strings.TrimSpace(value) {
	case "", "calendar":
		return calendarModeCalendar, nil
	case "tree":
		return calendarModeTree, nil
	case "board":
		return calendarModeBoard, nil
	case "review":
		return calendarModeReview, nil
	case "now":
		return calendarModeNow, nil
	default:
		return "", fmt.Errorf("unknown cockpit mode: %s", value)
	}
}

func defaultCockpitStatuses(mode calendarMode, cfg shelf.Config) []shelf.Status {
	switch mode {
	case calendarModeCalendar, calendarModeReview, calendarModeNow:
		return activeStatusFilter()
	case calendarModeTree, calendarModeBoard:
		return append([]shelf.Status{}, cfg.Statuses...)
	default:
		return append([]shelf.Status{}, cfg.Statuses...)
	}
}
