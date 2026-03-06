package shelf

import (
	"fmt"
	"strings"
	"time"
)

const dueOnLayout = "2006-01-02"

func NormalizeDueOn(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	switch strings.ToLower(trimmed) {
	case "today":
		return time.Now().Local().Format(dueOnLayout), nil
	case "tomorrow":
		return time.Now().Local().AddDate(0, 0, 1).Format(dueOnLayout), nil
	}
	if _, err := time.Parse(dueOnLayout, trimmed); err != nil {
		return "", fmt.Errorf("invalid due_on: %q (expected YYYY-MM-DD/today/tomorrow)", trimmed)
	}
	return trimmed, nil
}
