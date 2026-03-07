package cli

import (
	"testing"

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
