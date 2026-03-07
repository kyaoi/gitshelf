package shelf

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var repeatEveryPattern = regexp.MustCompile(`^[1-9][0-9]*[dwmy]$`)

func NormalizeRepeatEvery(value string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return "", nil
	}
	if !repeatEveryPattern.MatchString(v) {
		return "", fmt.Errorf("invalid repeat_every: %q (expected <N>d|<N>w|<N>m|<N>y)", value)
	}
	return v, nil
}

func AdvanceDueByRepeat(currentDue string, repeatEvery string, now time.Time) (string, error) {
	repeat, err := NormalizeRepeatEvery(repeatEvery)
	if err != nil {
		return "", err
	}
	if repeat == "" {
		return "", fmt.Errorf("repeat_every is empty")
	}

	amount, err := strconv.Atoi(repeat[:len(repeat)-1])
	if err != nil {
		return "", fmt.Errorf("invalid repeat_every: %q", repeatEvery)
	}
	unit := repeat[len(repeat)-1]

	base := now.Local()
	if strings.TrimSpace(currentDue) != "" {
		base, err = time.ParseInLocation(dueOnLayout, currentDue, time.Local)
		if err != nil {
			return "", fmt.Errorf("invalid current due_on: %q", currentDue)
		}
	}

	switch unit {
	case 'd':
		base = base.AddDate(0, 0, amount)
	case 'w':
		base = base.AddDate(0, 0, 7*amount)
	case 'm':
		base = base.AddDate(0, amount, 0)
	case 'y':
		base = base.AddDate(amount, 0, 0)
	default:
		return "", fmt.Errorf("invalid repeat_every unit: %q", unit)
	}
	return base.Format(dueOnLayout), nil
}
