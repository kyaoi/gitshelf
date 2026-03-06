package shelf

import "testing"

func TestBuildAndApplyTemplateFromTask(t *testing.T) {
	root := t.TempDir()
	if _, err := Initialize(root, false); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	parent, err := AddTask(root, AddTaskInput{Title: "Parent", Kind: "todo", Status: "open"})
	if err != nil {
		t.Fatalf("add parent failed: %v", err)
	}
	if _, err := AddTask(root, AddTaskInput{Title: "Child", Kind: "memo", Status: "open", Parent: parent.ID}); err != nil {
		t.Fatalf("add child failed: %v", err)
	}

	tpl, err := BuildTemplateFromTask(root, "weekly", parent.ID)
	if err != nil {
		t.Fatalf("build template failed: %v", err)
	}
	if len(tpl.Tasks) != 2 {
		t.Fatalf("unexpected template size: %d", len(tpl.Tasks))
	}
	if tpl.Tasks[0].ParentKey != "" {
		t.Fatalf("root task should have empty parent key: %+v", tpl.Tasks[0])
	}
	if tpl.Tasks[1].ParentKey != tpl.Tasks[0].Key {
		t.Fatalf("child should point to template root: %+v", tpl.Tasks[1])
	}

	if err := SaveTemplate(root, tpl); err != nil {
		t.Fatalf("save template failed: %v", err)
	}
	loaded, err := LoadTemplate(root, "weekly")
	if err != nil {
		t.Fatalf("load template failed: %v", err)
	}
	created, err := ApplyTemplate(root, loaded, "", "Copy: ")
	if err != nil {
		t.Fatalf("apply template failed: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("unexpected created task count: %d", len(created))
	}
	if created[0].Title != "Copy: Parent" {
		t.Fatalf("unexpected root title: %q", created[0].Title)
	}
	if created[1].Parent != created[0].ID {
		t.Fatalf("child should be attached to created root: %+v", created[1])
	}
}
