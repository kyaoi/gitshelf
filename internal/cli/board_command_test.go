package cli

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kyaoi/gitshelf/internal/shelf"
)

func TestBuildBoardColumns(t *testing.T) {
	statuses := []shelf.Status{"open", "done"}
	tasks := []shelf.Task{
		{ID: "01A", Title: "A", Status: "open"},
		{ID: "01B", Title: "B", Status: "done"},
		{ID: "01C", Title: "C", Status: "open"},
	}
	columns := buildBoardColumns(statuses, tasks)
	if len(columns) != 2 {
		t.Fatalf("unexpected column count: %d", len(columns))
	}
	if len(columns[0].Tasks) != 2 || len(columns[1].Tasks) != 1 {
		t.Fatalf("unexpected grouped columns: %+v", columns)
	}
}

func TestBoardSelectedTask(t *testing.T) {
	model := boardModel{
		columns: []boardColumn{
			{Status: "open", Tasks: []shelf.Task{{ID: "01A", Title: "A"}}},
		},
		rowIndex: map[int]int{0: 0},
	}
	task, ok := model.selectedTask()
	if !ok || task.ID != "01A" {
		t.Fatalf("unexpected selected task: %+v ok=%t", task, ok)
	}
}

func TestBoardUpdateAcceptsIShortcut(t *testing.T) {
	root := t.TempDir()
	if _, err := shelf.Initialize(root, false); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	task, err := shelf.AddTask(root, shelf.AddTaskInput{
		Title:  "Task",
		Kind:   "todo",
		Status: "open",
	})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}
	model := boardModel{
		rootDir: root,
		columns: []boardColumn{
			{Status: "open", Tasks: []shelf.Task{task}},
		},
		rowIndex: map[int]int{0: 0},
	}

	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	board := updatedModel.(boardModel)
	updated, err := shelf.EnsureTaskExists(root, task.ID)
	if err != nil {
		t.Fatalf("EnsureTaskExists failed: %v", err)
	}
	if updated.Status != "in_progress" {
		t.Fatalf("unexpected status: %s", updated.Status)
	}
	if board.message == "" {
		t.Fatal("expected status change message")
	}
}
