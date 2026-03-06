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
	if _, err := time.Parse(dueOnLayout, trimmed); err != nil {
		return "", fmt.Errorf("invalid due_on: %q (expected YYYY-MM-DD)", trimmed)
	}
	return trimmed, nil
}
