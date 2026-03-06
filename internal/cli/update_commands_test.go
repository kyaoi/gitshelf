package cli

import (
	"strings"
	"testing"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestBuildSetChangePreview(t *testing.T) {
	orig := shelf.Task{
		Title:  "old",
		Kind:   "todo",
		Status: "open",
		DueOn:  "",
		Parent: "",
	}
	title := "new"
	kind := shelf.Kind("memo")
	status := shelf.Status("done")
	due := "today"
	parent := "root"
	body := "replace"
	appendix := "append"

	preview := buildSetChangePreview(orig, shelf.SetTaskInput{
		Title:      &title,
		Kind:       &kind,
		Status:     &status,
		DueOn:      &due,
		Parent:     &parent,
		Body:       &body,
		AppendBody: &appendix,
	})

	for _, want := range []string{
		`Title: "old" -> "new"`,
		`Kind: "todo" -> "memo"`,
		`Status: "open" -> "done"`,
		`Due: "(none)" -> "today"`,
		`Parent: "root" -> "root"`,
		`Body: replace`,
		`Body: append`,
	} {
		if !strings.Contains(preview, want) {
			t.Fatalf("preview should contain %q, got: %s", want, preview)
		}
	}
}
