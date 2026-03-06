package cli

import (
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/interactive"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

func parseTagFlagValues(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			parts = append(parts, part)
		}
	}
	return shelf.NormalizeTags(parts)
}

func selectTagsInteractive(prompt string, catalog []string, current []string) ([]string, error) {
	selected := shelf.NormalizeTags(current)
	choices := shelf.NormalizeTags(catalog)
	for {
		options := make([]interactive.Option, 0, len(choices)+3)
		options = append(options, interactive.Option{
			Value:      "__done",
			Label:      fmt.Sprintf("Done (selected: %s)", formatTagSummary(selected)),
			SearchText: "done",
		})
		options = append(options, interactive.Option{
			Value:      "__new",
			Label:      "+ Add new tag",
			SearchText: "add new tag",
		})
		if len(selected) > 0 {
			options = append(options, interactive.Option{
				Value:      "__clear",
				Label:      "Clear selected tags",
				SearchText: "clear",
			})
		}
		for _, tag := range choices {
			marker := "[ ]"
			if containsTag(selected, tag) {
				marker = "[x]"
			}
			options = append(options, interactive.Option{
				Value:      tag,
				Label:      fmt.Sprintf("%s %s", marker, tag),
				SearchText: tag,
			})
		}

		chosen, err := selectEnumOption(prompt, options)
		if err != nil {
			return nil, err
		}
		switch chosen.Value {
		case "__done":
			return shelf.NormalizeTags(selected), nil
		case "__new":
			tag, err := interactive.PromptText("新しい tag を入力してください")
			if err != nil {
				return nil, err
			}
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			if !containsTag(choices, tag) {
				choices = append(choices, tag)
			}
			if !containsTag(selected, tag) {
				selected = append(selected, tag)
			}
		case "__clear":
			selected = []string{}
		default:
			tag := chosen.Value
			if containsTag(selected, tag) {
				selected = removeTag(selected, tag)
			} else {
				selected = append(selected, tag)
			}
			selected = shelf.NormalizeTags(selected)
		}
	}
}

func containsTag(tags []string, target string) bool {
	for _, tag := range tags {
		if tag == target {
			return true
		}
	}
	return false
}

func removeTag(tags []string, target string) []string {
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag == target {
			continue
		}
		result = append(result, tag)
	}
	return result
}

func formatTagSummary(tags []string) string {
	if len(tags) == 0 {
		return "(none)"
	}
	return strings.Join(tags, ", ")
}
