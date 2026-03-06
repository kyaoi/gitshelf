package shelf

import (
	"fmt"
	"strings"
	"time"
)

func ParseWorkDurationMinutes(value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("duration is required")
	}
	d, err := time.ParseDuration(trimmed)
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %w", err)
	}
	minutes := int(d / time.Minute)
	if d < 0 || minutes < 0 {
		return 0, fmt.Errorf("duration must be >= 0")
	}
	return minutes, nil
}

func FormatWorkMinutes(minutes int) string {
	if minutes <= 0 {
		return "0m"
	}
	hours := minutes / 60
	remain := minutes % 60
	if hours == 0 {
		return fmt.Sprintf("%dm", remain)
	}
	if remain == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, remain)
}

func ElapsedMinutesSince(startedAt string, now time.Time) (int, error) {
	normalized, err := normalizeTimerStartedAt(startedAt)
	if err != nil {
		return 0, err
	}
	if normalized == "" {
		return 0, fmt.Errorf("timer is not running")
	}
	start, err := time.Parse(time.RFC3339, normalized)
	if err != nil {
		return 0, err
	}
	if now.Before(start) {
		return 0, fmt.Errorf("timer_started_at is in the future")
	}
	elapsed := now.Sub(start)
	minutes := int(elapsed / time.Minute)
	if elapsed > 0 && minutes == 0 {
		minutes = 1
	}
	return minutes, nil
}
