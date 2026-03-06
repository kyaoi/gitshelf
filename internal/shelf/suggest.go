package shelf

import (
	"slices"
	"strings"
	"time"
	"unicode"
)

type LinkSuggestion struct {
	TaskID   string   `json:"task_id"`
	Title    string   `json:"title"`
	Kind     Kind     `json:"kind"`
	Status   Status   `json:"status"`
	DueOn    string   `json:"due_on,omitempty"`
	Parent   string   `json:"parent,omitempty"`
	LinkType LinkType `json:"link_type"`
	Score    int      `json:"score"`
	Reasons  []string `json:"reasons,omitempty"`
}

func SuggestDependsOn(rootDir, taskID string, limit int) ([]LinkSuggestion, error) {
	task, tasks, outboundDependsOn, _, err := loadSuggestionContext(rootDir, taskID)
	if err != nil {
		return nil, err
	}

	suggestions := make([]LinkSuggestion, 0)
	for _, candidate := range tasks {
		if candidate.ID == task.ID || candidate.ArchivedAt != "" || candidate.Status == "cancelled" {
			continue
		}
		if _, ok := outboundDependsOn[candidate.ID]; ok {
			continue
		}
		cycle, err := wouldCreateDependsOnCycle(rootDir, task.ID, candidate.ID)
		if err != nil {
			return nil, err
		}
		if cycle {
			continue
		}
		score, reasons := scoreDependsOnSuggestion(task, candidate)
		if score < 3 {
			continue
		}
		suggestions = append(suggestions, LinkSuggestion{
			TaskID:   candidate.ID,
			Title:    candidate.Title,
			Kind:     candidate.Kind,
			Status:   candidate.Status,
			DueOn:    candidate.DueOn,
			Parent:   candidate.Parent,
			LinkType: "depends_on",
			Score:    score,
			Reasons:  reasons,
		})
	}
	sortSuggestions(suggestions)
	if limit > 0 && len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}
	return suggestions, nil
}

func SuggestRelated(rootDir, taskID string, limit int) ([]LinkSuggestion, error) {
	task, tasks, _, linkedAny, err := loadSuggestionContext(rootDir, taskID)
	if err != nil {
		return nil, err
	}

	suggestions := make([]LinkSuggestion, 0)
	for _, candidate := range tasks {
		if candidate.ID == task.ID || candidate.ArchivedAt != "" || candidate.Status == "cancelled" {
			continue
		}
		if _, ok := linkedAny[candidate.ID]; ok {
			continue
		}
		score, reasons := scoreRelatedSuggestion(task, candidate)
		if score < 3 {
			continue
		}
		suggestions = append(suggestions, LinkSuggestion{
			TaskID:   candidate.ID,
			Title:    candidate.Title,
			Kind:     candidate.Kind,
			Status:   candidate.Status,
			DueOn:    candidate.DueOn,
			Parent:   candidate.Parent,
			LinkType: "related",
			Score:    score,
			Reasons:  reasons,
		})
	}
	sortSuggestions(suggestions)
	if limit > 0 && len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}
	return suggestions, nil
}

func loadSuggestionContext(rootDir, taskID string) (Task, []Task, map[string]struct{}, map[string]struct{}, error) {
	taskStore := NewTaskStore(rootDir)
	task, err := taskStore.Get(taskID)
	if err != nil {
		return Task{}, nil, nil, nil, err
	}
	tasks, err := taskStore.List()
	if err != nil {
		return Task{}, nil, nil, nil, err
	}
	outbound, inbound, err := ListLinks(rootDir, taskID)
	if err != nil {
		return Task{}, nil, nil, nil, err
	}
	outboundDependsOn := map[string]struct{}{}
	linkedAny := map[string]struct{}{}
	for _, edge := range outbound {
		linkedAny[edge.To] = struct{}{}
		if edge.Type == "depends_on" {
			outboundDependsOn[edge.To] = struct{}{}
		}
	}
	for _, edge := range inbound {
		linkedAny[edge.From] = struct{}{}
	}
	return task, tasks, outboundDependsOn, linkedAny, nil
}

func scoreDependsOnSuggestion(task, candidate Task) (int, []string) {
	score := 0
	reasons := make([]string, 0, 6)
	if task.Parent != "" && task.Parent == candidate.Parent {
		score += 2
		reasons = append(reasons, "same parent")
	}
	sharedTags := sharedSuggestionValues(task.Tags, candidate.Tags)
	if len(sharedTags) > 0 {
		score += len(sharedTags) * 3
		reasons = append(reasons, "shared tags: "+strings.Join(sharedTags, ", "))
	}
	if task.DueOn != "" && candidate.DueOn != "" && candidate.DueOn <= task.DueOn {
		score += 2
		reasons = append(reasons, "earlier or same due date")
	}
	if candidate.ID < task.ID {
		score++
		reasons = append(reasons, "older task")
	}
	if candidate.Status == "done" {
		score++
		reasons = append(reasons, "already done")
	}
	if repo := sharedGitHubRepo(task.GitHubURLs, candidate.GitHubURLs); repo != "" {
		score += 2
		reasons = append(reasons, "same GitHub repo: "+repo)
	}
	sharedTokens := sharedSuggestionValues(suggestionTokens(task.Title), suggestionTokens(candidate.Title))
	if len(sharedTokens) > 0 {
		tokenCount := len(sharedTokens)
		if tokenCount > 2 {
			tokenCount = 2
		}
		score += tokenCount
		reasons = append(reasons, "shared title tokens: "+strings.Join(sharedTokens, ", "))
	}
	return score, reasons
}

func scoreRelatedSuggestion(task, candidate Task) (int, []string) {
	score := 0
	reasons := make([]string, 0, 6)
	if task.Parent != "" && task.Parent == candidate.Parent {
		score += 2
		reasons = append(reasons, "same parent")
	}
	if task.Kind == candidate.Kind {
		score++
		reasons = append(reasons, "same kind")
	}
	sharedTags := sharedSuggestionValues(task.Tags, candidate.Tags)
	if len(sharedTags) > 0 {
		score += len(sharedTags) * 3
		reasons = append(reasons, "shared tags: "+strings.Join(sharedTags, ", "))
	}
	if repo := sharedGitHubRepo(task.GitHubURLs, candidate.GitHubURLs); repo != "" {
		score += 3
		reasons = append(reasons, "same GitHub repo: "+repo)
	}
	sharedTokens := sharedSuggestionValues(suggestionTokens(task.Title), suggestionTokens(candidate.Title))
	if len(sharedTokens) > 0 {
		tokenCount := len(sharedTokens)
		if tokenCount > 2 {
			tokenCount = 2
		}
		score += tokenCount
		reasons = append(reasons, "shared title tokens: "+strings.Join(sharedTokens, ", "))
	}
	if dueDatesClose(task.DueOn, candidate.DueOn) {
		score++
		reasons = append(reasons, "near due dates")
	}
	return score, reasons
}

func sharedSuggestionValues(left, right []string) []string {
	if len(left) == 0 || len(right) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(left))
	for _, value := range left {
		set[value] = struct{}{}
	}
	shared := make([]string, 0)
	for _, value := range right {
		if _, ok := set[value]; !ok {
			continue
		}
		shared = append(shared, value)
	}
	slices.Sort(shared)
	return slices.Compact(shared)
}

func suggestionTokens(value string) []string {
	fields := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(value)), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		if utf8Len := len([]rune(field)); utf8Len < 2 {
			continue
		}
		tokens = append(tokens, field)
	}
	slices.Sort(tokens)
	return slices.Compact(tokens)
}

func sharedGitHubRepo(left, right []string) string {
	if len(left) == 0 || len(right) == 0 {
		return ""
	}
	leftRepos := make(map[string]struct{}, len(left))
	for _, value := range left {
		if repo := githubRepoKey(value); repo != "" {
			leftRepos[repo] = struct{}{}
		}
	}
	for _, value := range right {
		repo := githubRepoKey(value)
		if repo == "" {
			continue
		}
		if _, ok := leftRepos[repo]; ok {
			return repo
		}
	}
	return ""
}

func githubRepoKey(value string) string {
	trimmed := strings.TrimPrefix(strings.TrimSpace(value), "https://github.com/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[0] + "/" + parts[1]
}

func dueDatesClose(left, right string) bool {
	if left == "" || right == "" {
		return false
	}
	leftDate, err := time.Parse(dueOnLayout, left)
	if err != nil {
		return false
	}
	rightDate, err := time.Parse(dueOnLayout, right)
	if err != nil {
		return false
	}
	delta := leftDate.Sub(rightDate)
	if delta < 0 {
		delta = -delta
	}
	return delta <= 7*24*time.Hour
}

func sortSuggestions(items []LinkSuggestion) {
	slices.SortFunc(items, func(a, b LinkSuggestion) int {
		if a.Score != b.Score {
			if a.Score > b.Score {
				return -1
			}
			return 1
		}
		if a.DueOn != b.DueOn {
			switch {
			case a.DueOn == "":
				return 1
			case b.DueOn == "":
				return -1
			case a.DueOn < b.DueOn:
				return -1
			case a.DueOn > b.DueOn:
				return 1
			}
		}
		if a.TaskID < b.TaskID {
			return -1
		}
		if a.TaskID > b.TaskID {
			return 1
		}
		return 0
	})
}
