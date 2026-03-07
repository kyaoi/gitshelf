package shelf

import "testing"

func TestLinkAndUnlinkTasks(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	a, err := AddTask(root, AddTaskInput{Title: "A"})
	if err != nil {
		t.Fatalf("add A failed: %v", err)
	}
	b, err := AddTask(root, AddTaskInput{Title: "B"})
	if err != nil {
		t.Fatalf("add B failed: %v", err)
	}

	if err := LinkTasks(root, a.ID, b.ID, "depends_on"); err != nil {
		t.Fatalf("link failed: %v", err)
	}
	outbound, inbound, err := ListLinks(root, b.ID)
	if err != nil {
		t.Fatalf("list links failed: %v", err)
	}
	if len(outbound) != 0 {
		t.Fatalf("expected no outbound from B, got %d", len(outbound))
	}
	if len(inbound) != 1 || inbound[0].From != a.ID {
		t.Fatalf("unexpected inbound: %+v", inbound)
	}

	removed, err := UnlinkTasks(root, a.ID, b.ID, "depends_on")
	if err != nil {
		t.Fatalf("unlink failed: %v", err)
	}
	if !removed {
		t.Fatal("expected link removal")
	}
}

func TestBuildTaskReadinessByDependsOn(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	a, err := AddTask(root, AddTaskInput{Title: "A"})
	if err != nil {
		t.Fatalf("add A failed: %v", err)
	}
	b, err := AddTask(root, AddTaskInput{Title: "B"})
	if err != nil {
		t.Fatalf("add B failed: %v", err)
	}
	if err := LinkTasks(root, a.ID, b.ID, "depends_on"); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	readiness, err := BuildTaskReadiness(root)
	if err != nil {
		t.Fatalf("build readiness failed: %v", err)
	}
	if readiness[a.ID].Ready {
		t.Fatalf("A should be blocked by dependency: %+v", readiness[a.ID])
	}
	if !readiness[a.ID].BlockedByDeps {
		t.Fatalf("A should be marked as blocked by deps: %+v", readiness[a.ID])
	}
	if !readiness[b.ID].Ready {
		t.Fatalf("B should be ready: %+v", readiness[b.ID])
	}

	done := Status("done")
	if _, err := SetTask(root, b.ID, SetTaskInput{Status: &done}); err != nil {
		t.Fatalf("set B done failed: %v", err)
	}
	readiness, err = BuildTaskReadiness(root)
	if err != nil {
		t.Fatalf("build readiness after done failed: %v", err)
	}
	if !readiness[a.ID].Ready {
		t.Fatalf("A should become ready when dependency is done: %+v", readiness[a.ID])
	}
}

func TestLinkTasksRejectsDependsOnCycle(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	a, err := AddTask(root, AddTaskInput{Title: "A"})
	if err != nil {
		t.Fatalf("add A failed: %v", err)
	}
	b, err := AddTask(root, AddTaskInput{Title: "B"})
	if err != nil {
		t.Fatalf("add B failed: %v", err)
	}
	if err := LinkTasks(root, a.ID, b.ID, "depends_on"); err != nil {
		t.Fatalf("link A->B failed: %v", err)
	}
	if err := LinkTasks(root, b.ID, a.ID, "depends_on"); err == nil {
		t.Fatal("expected depends_on cycle error")
	}
}

func TestListTransitiveDependencies(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	a, _ := AddTask(root, AddTaskInput{Title: "A"})
	b, _ := AddTask(root, AddTaskInput{Title: "B"})
	c, _ := AddTask(root, AddTaskInput{Title: "C"})
	d, _ := AddTask(root, AddTaskInput{Title: "D"})
	if err := LinkTasks(root, a.ID, b.ID, "depends_on"); err != nil {
		t.Fatalf("link A->B failed: %v", err)
	}
	if err := LinkTasks(root, b.ID, c.ID, "depends_on"); err != nil {
		t.Fatalf("link B->C failed: %v", err)
	}
	if err := LinkTasks(root, a.ID, d.ID, "related"); err != nil {
		t.Fatalf("link A->D related failed: %v", err)
	}

	got, err := ListTransitiveDependencies(root, a.ID)
	if err != nil {
		t.Fatalf("list transitive dependencies failed: %v", err)
	}
	if len(got) != 2 || got[0] != b.ID || got[1] != c.ID {
		t.Fatalf("unexpected transitive dependencies: %+v", got)
	}
}
