package cli

import "fmt"

func validateFormat(value string, allowed []string) error {
	for _, item := range allowed {
		if value == item {
			return nil
		}
	}
	return fmt.Errorf("invalid --format: %s (allowed: %v)", value, allowed)
}
