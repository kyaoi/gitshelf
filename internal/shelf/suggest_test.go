package shelf

import "testing"

func TestSuggestDependsOnSkipsExistingEdgesAndCycles(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	parent, err := AddTask(root, AddTaskInput{Title: "Parent"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	target, err := AddTask(root, AddTaskInput{
		Title:  "Implement API",
		Kind:   "todo",
		Status: "open",
		Tags:   []string{"backend"},
		DueOn:  "2026-03-10",
		Parent: parent.ID,
	})
	if err != nil {
		t.Fatalf("add target failed: %v", err)
	}
	good, err := AddTask(root, AddTaskInput{
		Title:  "Design API",
		Kind:   "todo",
		Status: "done",
		Tags:   []string{"backend"},
		DueOn:  "2026-03-08",
		Parent: parent.ID,
	})
	if err != nil {
		t.Fatalf("add good failed: %v", err)
	}
	existing, err := AddTask(root, AddTaskInput{
		Title:  "Existing dep",
		Kind:   "todo",
		Status: "open",
		Tags:   []string{"backend"},
		DueOn:  "2026-03-07",
		Parent: parent.ID,
	})
	if err != nil {
		t.Fatalf("add existing failed: %v", err)
	}
	cycle, err := AddTask(root, AddTaskInput{
		Title:  "Cycle dep",
		Kind:   "todo",
		Status: "open",
		Tags:   []string{"backend"},
		DueOn:  "2026-03-09",
		Parent: parent.ID,
	})
	if err != nil {
		t.Fatalf("add cycle failed: %v", err)
	}
	if err := LinkTasks(root, target.ID, existing.ID, "depends_on"); err != nil {
		t.Fatalf("link existing failed: %v", err)
	}
	if err := LinkTasks(root, cycle.ID, target.ID, "depends_on"); err != nil {
		t.Fatalf("link cycle failed: %v", err)
	}

	suggestions, err := SuggestDependsOn(root, target.ID, 10)
	if err != nil {
		t.Fatalf("suggest depends_on failed: %v", err)
	}
	if len(suggestions) == 0 || suggestions[0].TaskID != good.ID {
		t.Fatalf("expected best suggestion to be %s, got: %+v", good.ID, suggestions)
	}
	for _, item := range suggestions {
		if item.TaskID == existing.ID {
			t.Fatalf("existing dependency should not be suggested: %+v", suggestions)
		}
		if item.TaskID == cycle.ID {
			t.Fatalf("cycle candidate should not be suggested: %+v", suggestions)
		}
	}
}

func TestSuggestRelatedPrefersSharedContextAndSkipsLinkedTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	parent, err := AddTask(root, AddTaskInput{Title: "Parent"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	target, err := AddTask(root, AddTaskInput{
		Title:  "UI polish",
		Kind:   "todo",
		Status: "open",
		Tags:   []string{"frontend", "ux"},
		Parent: parent.ID,
	})
	if err != nil {
		t.Fatalf("add target failed: %v", err)
	}
	target.GitHubURLs = []string{"https://github.com/acme/app/issues/10"}
	if _, err := SetTask(root, target.ID, SetTaskInput{GitHubURLs: &target.GitHubURLs}); err != nil {
		t.Fatalf("set target github urls failed: %v", err)
	}
	good, err := AddTask(root, AddTaskInput{
		Title:  "UX checklist",
		Kind:   "memo",
		Status: "open",
		Tags:   []string{"ux"},
		Parent: parent.ID,
	})
	if err != nil {
		t.Fatalf("add good failed: %v", err)
	}
	goodURLs := []string{"https://github.com/acme/app/pull/25"}
	if _, err := SetTask(root, good.ID, SetTaskInput{GitHubURLs: &goodURLs}); err != nil {
		t.Fatalf("set good github urls failed: %v", err)
	}
	linked, err := AddTask(root, AddTaskInput{
		Title:  "Already linked",
		Kind:   "memo",
		Status: "open",
		Tags:   []string{"ux"},
		Parent: parent.ID,
	})
	if err != nil {
		t.Fatalf("add linked failed: %v", err)
	}
	if err := LinkTasks(root, target.ID, linked.ID, "related"); err != nil {
		t.Fatalf("link related failed: %v", err)
	}

	suggestions, err := SuggestRelated(root, target.ID, 10)
	if err != nil {
		t.Fatalf("suggest related failed: %v", err)
	}
	if len(suggestions) == 0 || suggestions[0].TaskID != good.ID {
		t.Fatalf("expected best related suggestion to be %s, got: %+v", good.ID, suggestions)
	}
	for _, item := range suggestions {
		if item.TaskID == linked.ID {
			t.Fatalf("already linked task should not be suggested: %+v", suggestions)
		}
	}
}
