package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestCLIMachineReadableOutputGoldens(t *testing.T) {
	root := createGoldenFixtureRoot(t)
	cases := []struct {
		name string
		args []string
	}{
		{
			name: "ls-json",
			args: []string{"ls", "--root", root, "--json", "--sort", "title"},
		},
		{
			name: "ls-jsonl",
			args: []string{"ls", "--root", root, "--format", "jsonl", "--sort", "title"},
		},
		{
			name: "ls-csv",
			args: []string{"ls", "--root", root, "--format", "csv", "--sort", "title"},
		},
		{
			name: "ls-tsv",
			args: []string{"ls", "--root", root, "--format", "tsv", "--fields", "id,title,parent_id,path,file", "--sort", "title"},
		},
		{
			name: "ls-group-by-parent-json",
			args: []string{"ls", "--root", root, "--json", "--group-by", "parent", "--sort", "title"},
		},
		{
			name: "next-json",
			args: []string{"next", "--root", root, "--json", "--sort", "title"},
		},
		{
			name: "next-jsonl",
			args: []string{"next", "--root", root, "--format", "jsonl", "--sort", "title"},
		},
		{
			name: "next-tsv",
			args: []string{"next", "--root", root, "--format", "tsv", "--fields", "id,title,parent_id,path,due_on,file", "--sort", "title"},
		},
		{
			name: "next-csv",
			args: []string{"next", "--root", root, "--format", "csv", "--fields", "id,title,parent_id,path,due_on,file", "--sort", "title"},
		},
		{
			name: "show-json",
			args: []string{"show", "--root", root, goldenTaskChildID, "--json"},
		},
		{
			name: "show-jsonl",
			args: []string{"show", "--root", root, goldenTaskChildID, "--format", "jsonl"},
		},
		{
			name: "show-csv",
			args: []string{"show", "--root", root, goldenTaskChildID, "--format", "csv", "--fields", "id,title,parent_id,path,file,body"},
		},
		{
			name: "links-json",
			args: []string{"links", "--root", root, goldenTaskChildID, "--json"},
		},
		{
			name: "links-jsonl",
			args: []string{"links", "--root", root, goldenTaskChildID, "--format", "jsonl"},
		},
		{
			name: "links-csv",
			args: []string{"links", "--root", root, goldenTaskChildID, "--format", "csv", "--fields", "direction,type,source_id,target_id,target_path,target_file"},
		},
		{
			name: "links-tsv",
			args: []string{"links", "--root", root, goldenTaskChildID, "--format", "tsv", "--fields", "direction,type,source_id,target_id,target_path,target_file"},
		},
		{
			name: "links-summary-json",
			args: []string{"links", "--root", root, goldenTaskChildID, "--json", "--summary"},
		},
		{
			name: "links-summary-csv",
			args: []string{"links", "--root", root, goldenTaskChildID, "--summary", "--format", "csv"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			out, err := executeCLI(t, tc.args...)
			if err != nil {
				t.Fatalf("%s failed: %v", tc.name, err)
			}
			assertGoldenOutput(t, tc.name, normalizeGoldenOutput(root, out))
		})
	}
}

const (
	goldenTaskParentID  = "01FIXPARENT0000000000000001"
	goldenTaskChildID   = "01FIXCHILD0000000000000002"
	goldenTaskBlockedID = "01FIXBLOCK0000000000000003"
	goldenTaskPeerID    = "01FIXPEER0000000000000004"
	goldenTaskArchiveID = "01FIXARCHI0000000000000005"
)

func createGoldenFixtureRoot(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	if _, err := executeCLI(t, "init", "--root", root); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	store := shelf.NewTaskStore(root)
	base := time.Date(2026, 3, 11, 9, 0, 0, 0, time.UTC)
	tasks := []shelf.Task{
		{
			ID:        goldenTaskParentID,
			Title:     "Project Alpha",
			Kind:      "todo",
			Status:    "open",
			Tags:      []string{"focus"},
			DueOn:     "2026-03-20",
			CreatedAt: base,
			UpdatedAt: base,
			Body:      "parent note",
		},
		{
			ID:          goldenTaskChildID,
			Title:       "Build CLI",
			Kind:        "todo",
			Status:      "in_progress",
			Tags:        []string{"cli", "release"},
			DueOn:       "2026-03-12",
			RepeatEvery: "1w",
			Parent:      goldenTaskParentID,
			CreatedAt:   base.Add(1 * time.Hour),
			UpdatedAt:   base.Add(2 * time.Hour),
			Body:        "line one\nline two",
		},
		{
			ID:        goldenTaskBlockedID,
			Title:     "Blocked Task",
			Kind:      "todo",
			Status:    "open",
			Tags:      []string{"blocked"},
			CreatedAt: base.Add(3 * time.Hour),
			UpdatedAt: base.Add(3 * time.Hour),
		},
		{
			ID:        goldenTaskPeerID,
			Title:     "Peer Task",
			Kind:      "memo",
			Status:    "open",
			Tags:      []string{"notes"},
			CreatedAt: base.Add(4 * time.Hour),
			UpdatedAt: base.Add(4 * time.Hour),
			Body:      "peer body",
		},
		{
			ID:         goldenTaskArchiveID,
			Title:      "Archived Task",
			Kind:       "todo",
			Status:     "done",
			ArchivedAt: "2026-03-10T00:00:00Z",
			CreatedAt:  base.Add(5 * time.Hour),
			UpdatedAt:  base.Add(5 * time.Hour),
		},
	}
	for _, task := range tasks {
		if err := store.Upsert(task); err != nil {
			t.Fatalf("upsert task %s failed: %v", task.ID, err)
		}
	}

	if err := shelf.LinkTasks(root, goldenTaskChildID, goldenTaskPeerID, "depends_on"); err != nil {
		t.Fatalf("link child -> peer failed: %v", err)
	}
	if err := shelf.LinkTasks(root, goldenTaskParentID, goldenTaskChildID, "related"); err != nil {
		t.Fatalf("link parent -> child failed: %v", err)
	}
	if err := shelf.LinkTasks(root, goldenTaskBlockedID, goldenTaskPeerID, "depends_on"); err != nil {
		t.Fatalf("link blocked -> peer failed: %v", err)
	}

	return root
}

func normalizeGoldenOutput(root string, text string) string {
	return strings.ReplaceAll(text, root, "{{ROOT}}")
}

func assertGoldenOutput(t *testing.T, name string, actual string) {
	t.Helper()

	path := filepath.Join("testdata", "outputs", name+".golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir golden dir failed: %v", err)
		}
		if err := os.WriteFile(path, []byte(actual), 0o644); err != nil {
			t.Fatalf("write golden failed: %v", err)
		}
	}
	wantBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s failed: %v", path, err)
	}
	if diff := compareGoldenStrings(string(wantBytes), actual); diff != "" {
		t.Fatalf("%s golden mismatch:\n%s", name, diff)
	}
}

func compareGoldenStrings(want string, got string) string {
	if want == got {
		return ""
	}
	return "want:\n" + want + "\n\ngot:\n" + got
}
