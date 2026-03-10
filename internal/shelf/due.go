package shelf

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const dueOnLayout = "2006-01-02"

var (
	nextWeekdayPattern = regexp.MustCompile(`^next\-(mon|tue|wed|thu|fri|sat|sun)$`)
	inDaysPattern      = regexp.MustCompile(`^in\s+(-?\d+)\s+days?$`)
)

func NormalizeDueOn(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	lower := strings.ToLower(trimmed)
	switch lower {
	case "today":
		return time.Now().Local().Format(dueOnLayout), nil
	case "tomorrow":
		return time.Now().Local().AddDate(0, 0, 1).Format(dueOnLayout), nil
	case "next-week":
		return time.Now().Local().AddDate(0, 0, 7).Format(dueOnLayout), nil
	case "this-week":
		return endOfWeekDate(time.Now().Local()).Format(dueOnLayout), nil
	case "mon", "tue", "wed", "thu", "fri", "sat", "sun":
		return nextWeekdayDate(lower, time.Now().Local()).Format(dueOnLayout), nil
	}
	if match := nextWeekdayPattern.FindStringSubmatch(lower); len(match) == 2 {
		return nextWeekdayDateStrict(match[1], time.Now().Local()).Format(dueOnLayout), nil
	}
	if match := inDaysPattern.FindStringSubmatch(lower); len(match) == 2 {
		days, err := strconv.Atoi(match[1])
		if err != nil {
			return "", fmt.Errorf("invalid due_on: %q", trimmed)
		}
		return time.Now().Local().AddDate(0, 0, days).Format(dueOnLayout), nil
	}
	if days, ok := parseRelativeDayToken(lower); ok {
		return time.Now().Local().AddDate(0, 0, days).Format(dueOnLayout), nil
	}
	if _, err := time.Parse(dueOnLayout, trimmed); err != nil {
		return "", fmt.Errorf("invalid due_on: %q (expected YYYY-MM-DD/today/tomorrow/+Nd/-Nd/next-week/this-week/mon..sun/next-mon..next-sun/in N days)", trimmed)
	}
	return trimmed, nil
}

func parseRelativeDayToken(value string) (int, bool) {
	if !strings.HasSuffix(value, "d") {
		return 0, false
	}
	num := strings.TrimSuffix(value, "d")
	if num == "" || num == "+" || num == "-" {
		return 0, false
	}
	parsed, err := strconv.Atoi(num)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func nextWeekdayDate(token string, now time.Time) time.Time {
	target := map[string]time.Weekday{
		"sun": time.Sunday,
		"mon": time.Monday,
		"tue": time.Tuesday,
		"wed": time.Wednesday,
		"thu": time.Thursday,
		"fri": time.Friday,
		"sat": time.Saturday,
	}[token]
	delta := int(target - now.Weekday())
	if delta < 0 {
		delta += 7
	}
	return now.AddDate(0, 0, delta)
}

func nextWeekdayDateStrict(token string, now time.Time) time.Time {
	target := map[string]time.Weekday{
		"sun": time.Sunday,
		"mon": time.Monday,
		"tue": time.Tuesday,
		"wed": time.Wednesday,
		"thu": time.Thursday,
		"fri": time.Friday,
		"sat": time.Saturday,
	}[token]
	delta := int(target - now.Weekday())
	if delta <= 0 {
		delta += 7
	}
	return now.AddDate(0, 0, delta)
}

func endOfWeekDate(now time.Time) time.Time {
	delta := int(time.Saturday - now.Weekday())
	if delta < 0 {
		delta += 7
	}
	return now.AddDate(0, 0, delta)
}
