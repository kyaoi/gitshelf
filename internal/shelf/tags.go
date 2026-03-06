package shelf

import "strings"

func NormalizeTags(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	normalized := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		tag := strings.TrimSpace(value)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}
	return normalized
}

func containsAnyTag(taskTags []string, filterTags []string) bool {
	if len(taskTags) == 0 || len(filterTags) == 0 {
		return false
	}
	set := map[string]struct{}{}
	for _, tag := range taskTags {
		set[tag] = struct{}{}
	}
	for _, filterTag := range filterTags {
		if _, ok := set[filterTag]; ok {
			return true
		}
	}
	return false
}
