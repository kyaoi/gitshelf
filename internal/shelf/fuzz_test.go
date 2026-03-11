package shelf

import (
	"reflect"
	"testing"
	"time"
)

func FuzzNormalizeTags(f *testing.F) {
	f.Add(" backend ", "urgent", "backend", "")
	f.Add("", " ", "\t", "\n")
	f.Add("alpha", "beta", "alpha", " beta ")

	f.Fuzz(func(t *testing.T, a string, b string, c string, d string) {
		got := NormalizeTags([]string{a, b, c, d})
		seen := map[string]struct{}{}
		for _, tag := range got {
			if tag == "" {
				t.Fatalf("NormalizeTags returned empty tag: %+v", got)
			}
			if _, ok := seen[tag]; ok {
				t.Fatalf("NormalizeTags returned duplicate tag %q from %+v", tag, got)
			}
			seen[tag] = struct{}{}
		}
		if !reflect.DeepEqual(got, NormalizeTags(got)) {
			t.Fatalf("NormalizeTags should be idempotent: got=%+v", got)
		}
	})
}

func FuzzTaskMarkdownRoundTrip(f *testing.F) {
	f.Add("01JABCDEF0123456789XYZ", "weekly review", "todo", "open", "backend", "urgent", "2026-03-10", "1w", "", "01JPARENT00000000000000", "body line 1\nbody line 2")
	f.Add("01JEMPTYBODY000000000000", "trimmed tags", "todo", "in_progress", " alpha ", "alpha", "", "", "", "", "")

	f.Fuzz(func(t *testing.T, id string, title string, kind string, status string, tagA string, tagB string, dueOn string, repeatEvery string, archivedAt string, parent string, body string) {
		now := time.Date(2026, 3, 5, 12, 34, 56, 0, time.UTC)
		task := Task{
			ID:          id,
			Title:       title,
			Kind:        Kind(kind),
			Status:      Status(status),
			Tags:        []string{tagA, tagB},
			DueOn:       dueOn,
			RepeatEvery: repeatEvery,
			ArchivedAt:  archivedAt,
			Parent:      parent,
			CreatedAt:   now,
			UpdatedAt:   now,
			Body:        body,
		}

		data, err := FormatTaskMarkdown(task)
		if err != nil {
			return
		}

		parsed, err := ParseTaskMarkdown(data)
		if err != nil {
			t.Fatalf("ParseTaskMarkdown should succeed after FormatTaskMarkdown: %v\n%s", err, string(data))
		}
		wantDueOn, err := NormalizeDueOn(task.DueOn)
		if err != nil {
			t.Fatalf("NormalizeDueOn failed: %v", err)
		}
		wantRepeatEvery, err := NormalizeRepeatEvery(task.RepeatEvery)
		if err != nil {
			t.Fatalf("NormalizeRepeatEvery failed: %v", err)
		}
		wantArchivedAt, err := normalizeArchivedAt(task.ArchivedAt)
		if err != nil {
			t.Fatalf("normalizeArchivedAt failed: %v", err)
		}
		if parsed.ID != task.ID || parsed.Title != task.Title || parsed.Kind != task.Kind || parsed.Status != task.Status {
			t.Fatalf("task identity mismatch after round-trip: got=%+v want=%+v", parsed, task)
		}
		if parsed.DueOn != wantDueOn || parsed.RepeatEvery != wantRepeatEvery || parsed.ArchivedAt != wantArchivedAt {
			t.Fatalf("normalized date fields changed unexpectedly: got=%+v want=%+v", parsed, task)
		}
		if parsed.Parent != task.Parent || parsed.Body != task.Body {
			t.Fatalf("optional fields mismatch after round-trip: got=%+v want=%+v", parsed, task)
		}
		if !reflect.DeepEqual(parsed.Tags, NormalizeTags(task.Tags)) {
			t.Fatalf("tag normalization mismatch: got=%+v want=%+v", parsed.Tags, NormalizeTags(task.Tags))
		}
	})
}

func FuzzEdgesTOMLRoundTrip(f *testing.F) {
	f.Add("01A", "depends_on", "01B", "related")
	f.Add(" 01C ", " related ", "", "")

	f.Fuzz(func(t *testing.T, toA string, typeA string, toB string, typeB string) {
		edges := []Edge{
			{To: toA, Type: LinkType(typeA)},
			{To: toB, Type: LinkType(typeB)},
		}
		data := FormatEdgesTOML(edges)
		parsed, err := ParseEdgesTOML(data)
		if err != nil {
			t.Fatalf("ParseEdgesTOML should succeed after FormatEdgesTOML: %v\n%s", err, string(data))
		}
		want := normalizeEdges(edges)
		if !reflect.DeepEqual(parsed, want) {
			t.Fatalf("edge round-trip mismatch: got=%+v want=%+v", parsed, want)
		}
	})
}

func FuzzParseTaskMarkdown(f *testing.F) {
	f.Add([]byte("+++\nid = \"01J\"\ntitle = \"demo\"\nkind = \"todo\"\nstatus = \"open\"\ncreated_at = \"2026-03-05T12:34:56Z\"\nupdated_at = \"2026-03-05T12:34:56Z\"\n+++\n\nbody\n"))
	f.Add([]byte("not-front-matter"))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseTaskMarkdown(data)
	})
}
