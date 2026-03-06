package shelf

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Kind string
type Status string
type LinkType string

type Task struct {
	ID          string
	Title       string
	Kind        Kind
	Status      Status
	Tags        []string
	EstimateMin int
	SpentMin    int
	TimerStart  string
	DueOn       string
	RepeatEvery string
	ArchivedAt  string
	Parent      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Body        string
}

type taskFrontMatter struct {
	ID          string   `toml:"id"`
	Title       string   `toml:"title"`
	Kind        string   `toml:"kind"`
	Status      string   `toml:"status"`
	State       string   `toml:"state"`
	Tags        []string `toml:"tags"`
	EstimateMin int      `toml:"estimate_minutes"`
	SpentMin    int      `toml:"spent_minutes"`
	TimerStart  string   `toml:"timer_started_at"`
	DueOn       string   `toml:"due_on,omitempty"`
	RepeatEvery string   `toml:"repeat_every,omitempty"`
	ArchivedAt  string   `toml:"archived_at,omitempty"`
	Parent      string   `toml:"parent,omitempty"`
	CreatedAt   string   `toml:"created_at"`
	UpdatedAt   string   `toml:"updated_at"`
}

const frontMatterDelimiter = "+++"

func ParseTaskMarkdown(data []byte) (Task, error) {
	text := string(data)
	lines := strings.Split(text, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != frontMatterDelimiter {
		return Task{}, errors.New("front matter start delimiter `+++` is missing")
	}

	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == frontMatterDelimiter {
			end = i
			break
		}
	}
	if end == -1 {
		return Task{}, errors.New("front matter end delimiter `+++` is missing")
	}

	frontMatterRaw := strings.Join(lines[1:end], "\n")
	var fm taskFrontMatter
	if _, err := toml.Decode(frontMatterRaw, &fm); err != nil {
		return Task{}, fmt.Errorf("failed to parse front matter TOML: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, fm.CreatedAt)
	if err != nil {
		return Task{}, fmt.Errorf("invalid created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, fm.UpdatedAt)
	if err != nil {
		return Task{}, fmt.Errorf("invalid updated_at: %w", err)
	}
	status := strings.TrimSpace(fm.Status)
	if status == "" {
		status = strings.TrimSpace(fm.State)
	}
	dueOn, err := NormalizeDueOn(fm.DueOn)
	if err != nil {
		return Task{}, err
	}
	repeatEvery, err := NormalizeRepeatEvery(fm.RepeatEvery)
	if err != nil {
		return Task{}, err
	}
	archivedAt, err := normalizeArchivedAt(fm.ArchivedAt)
	if err != nil {
		return Task{}, err
	}
	timerStart, err := normalizeTimerStartedAt(fm.TimerStart)
	if err != nil {
		return Task{}, err
	}

	body := ""
	if end+1 < len(lines) {
		body = strings.Join(lines[end+1:], "\n")
		if strings.HasPrefix(body, "\n") {
			body = strings.TrimPrefix(body, "\n")
		}
		body = strings.TrimSuffix(body, "\n")
	}

	task := Task{
		ID:          fm.ID,
		Title:       fm.Title,
		Kind:        Kind(fm.Kind),
		Status:      Status(status),
		Tags:        NormalizeTags(fm.Tags),
		EstimateMin: fm.EstimateMin,
		SpentMin:    fm.SpentMin,
		TimerStart:  timerStart,
		DueOn:       dueOn,
		RepeatEvery: repeatEvery,
		ArchivedAt:  archivedAt,
		Parent:      fm.Parent,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Body:        body,
	}
	if err := validateTaskRequiredFields(task); err != nil {
		return Task{}, err
	}
	return task, nil
}

func FormatTaskMarkdown(task Task) ([]byte, error) {
	if err := validateTaskRequiredFields(task); err != nil {
		return nil, err
	}
	task.Tags = NormalizeTags(task.Tags)

	var buf bytes.Buffer
	buf.WriteString(frontMatterDelimiter + "\n")
	buf.WriteString(fmt.Sprintf("id = %q\n", task.ID))
	buf.WriteString(fmt.Sprintf("title = %q\n", task.Title))
	buf.WriteString(fmt.Sprintf("kind = %q\n", string(task.Kind)))
	buf.WriteString(fmt.Sprintf("status = %q\n", string(task.Status)))
	if len(task.Tags) > 0 {
		buf.WriteString("tags = [")
		for i, tag := range task.Tags {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(fmt.Sprintf("%q", tag))
		}
		buf.WriteString("]\n")
	}
	if task.EstimateMin > 0 {
		buf.WriteString(fmt.Sprintf("estimate_minutes = %d\n", task.EstimateMin))
	}
	if task.SpentMin > 0 {
		buf.WriteString(fmt.Sprintf("spent_minutes = %d\n", task.SpentMin))
	}
	if strings.TrimSpace(task.TimerStart) != "" {
		buf.WriteString(fmt.Sprintf("timer_started_at = %q\n", task.TimerStart))
	}
	if task.DueOn != "" {
		buf.WriteString(fmt.Sprintf("due_on = %q\n", task.DueOn))
	}
	if task.RepeatEvery != "" {
		buf.WriteString(fmt.Sprintf("repeat_every = %q\n", task.RepeatEvery))
	}
	if task.ArchivedAt != "" {
		buf.WriteString(fmt.Sprintf("archived_at = %q\n", task.ArchivedAt))
	}
	if task.Parent != "" {
		buf.WriteString(fmt.Sprintf("parent = %q\n", task.Parent))
	}
	buf.WriteString(fmt.Sprintf("created_at = %q\n", task.CreatedAt.Format(time.RFC3339)))
	buf.WriteString(fmt.Sprintf("updated_at = %q\n", task.UpdatedAt.Format(time.RFC3339)))
	buf.WriteString(frontMatterDelimiter + "\n\n")
	buf.WriteString(task.Body)
	if !strings.HasSuffix(task.Body, "\n") {
		buf.WriteString("\n")
	}
	return buf.Bytes(), nil
}

func validateTaskRequiredFields(task Task) error {
	switch {
	case task.ID == "":
		return errors.New("task id is required")
	case strings.TrimSpace(task.Title) == "":
		return errors.New("task title is required")
	case task.Kind == "":
		return errors.New("task kind is required")
	case task.Status == "":
		return errors.New("task status is required")
	case task.CreatedAt.IsZero():
		return errors.New("task created_at is required")
	case task.UpdatedAt.IsZero():
		return errors.New("task updated_at is required")
	default:
		task.Tags = NormalizeTags(task.Tags)
		if task.EstimateMin < 0 {
			return errors.New("estimate_minutes must be >= 0")
		}
		if task.SpentMin < 0 {
			return errors.New("spent_minutes must be >= 0")
		}
		if _, err := normalizeTimerStartedAt(task.TimerStart); err != nil {
			return err
		}
		if _, err := NormalizeDueOn(task.DueOn); err != nil {
			return err
		}
		if _, err := NormalizeRepeatEvery(task.RepeatEvery); err != nil {
			return err
		}
		if _, err := normalizeArchivedAt(task.ArchivedAt); err != nil {
			return err
		}
		return nil
	}
}

func normalizeArchivedAt(value string) (string, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return "", nil
	}
	parsed, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return "", fmt.Errorf("invalid archived_at: %w", err)
	}
	return parsed.Format(time.RFC3339), nil
}

func normalizeTimerStartedAt(value string) (string, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return "", nil
	}
	parsed, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return "", fmt.Errorf("invalid timer_started_at: %w", err)
	}
	return parsed.Format(time.RFC3339), nil
}
