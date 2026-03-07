package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newCockpitCommand(ctx *commandContext) *cobra.Command {
	var (
		mode      string
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
	)

	cmd := &cobra.Command{
		Use:     "cockpit",
		Aliases: []string{"cp"},
		Short:   "Open the unified Daily Cockpit TUI",
		Example: "  shelf cockpit\n" +
			"  shelf cockpit --mode tree\n" +
			"  shelf cockpit --mode board --months 3\n" +
			"  shelf cockpit --mode review --status open --status blocked",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !dailyCockpitIsTTY() {
				return errors.New("cockpit はTTYが必要です")
			}
			cfg, err := shelf.LoadConfig(ctx.rootDir)
			if err != nil {
				return err
			}
			parsedMode, err := parseCockpitMode(mode)
			if err != nil {
				return err
			}
			for _, kind := range toKinds(kinds) {
				if err := cfg.ValidateKind(kind); err != nil {
					return err
				}
			}
			for _, kind := range toKinds(notKinds) {
				if err := cfg.ValidateKind(kind); err != nil {
					return err
				}
			}
			for _, status := range toStatuses(statuses) {
				if err := cfg.ValidateStatus(status); err != nil {
					return err
				}
			}
			for _, status := range toStatuses(notStatus) {
				if err := cfg.ValidateStatus(status); err != nil {
					return err
				}
			}
			for _, tag := range parseTagFlagValues(tags) {
				if err := cfg.ValidateTag(tag); err != nil {
					return err
				}
			}
			for _, tag := range parseTagFlagValues(notTags) {
				if err := cfg.ValidateTag(tag); err != nil {
					return err
				}
			}

			startDate, err := resolveCalendarStart(start)
			if err != nil {
				return err
			}
			rangeStart, dayCount, err := resolveCalendarRange(
				startDate,
				days,
				months,
				years,
				cfg.Commands.Calendar,
				cmd.Flags().Changed("days"),
				cmd.Flags().Changed("months"),
				cmd.Flags().Changed("years"),
			)
			if err != nil {
				return err
			}

			filter := shelf.TaskFilter{
				Kinds:       toKinds(kinds),
				Statuses:    toStatuses(statuses),
				Tags:        parseTagFlagValues(tags),
				NotKinds:    toKinds(notKinds),
				NotStatuses: toStatuses(notStatus),
				NotTags:     parseTagFlagValues(notTags),
				Limit:       0,
			}
			defaultStatuses := defaultCockpitStatuses(parsedMode, cfg)
			return runCalendarModeTUIFn(ctx.rootDir, rangeStart, dayCount, defaultStatuses, calendarTUIOptions{
				Mode:         parsedMode,
				ShowID:       ctx.showID,
				SectionLimit: limit,
				Filter:       filter,
			})
		},
	}

	cmd.Flags().StringVar(&mode, "mode", "calendar", "Starting mode: calendar|tree|board|review|now")
	cmd.Flags().StringVar(&start, "start", "", "Anchor date (YYYY-MM-DD|today|tomorrow). Defaults to current week Monday")
	cmd.Flags().IntVar(&days, "days", 0, "Render an explicit day range")
	cmd.Flags().IntVar(&months, "months", 0, "Render an explicit month range from the month containing --start")
	cmd.Flags().IntVar(&years, "years", 0, "Render an explicit year range from the year containing --start")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum items per non-focused section (0 means unlimited)")
	cmd.Flags().StringArrayVar(&kinds, "kind", nil, "Include kind (repeatable)")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "Include status (repeatable)")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Include tag (repeatable)")
	cmd.Flags().StringArrayVar(&notKinds, "not-kind", nil, "Exclude kind (repeatable)")
	cmd.Flags().StringArrayVar(&notStatus, "not-status", nil, "Exclude status (repeatable)")
	cmd.Flags().StringArrayVar(&notTags, "not-tag", nil, "Exclude tag (repeatable)")
	return cmd
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
	case "now", "today":
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
