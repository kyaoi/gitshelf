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
