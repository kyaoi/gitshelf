package shelf

import (
	"fmt"
	"regexp"
	"strings"
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
