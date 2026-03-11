package shelf

import (
	"os"
	"testing"
)

func TestEdgesAreSortedAndDeduplicated(t *testing.T) {
	edges := []Edge{
		{To: "B", Type: LinkType("related")},
		{To: "A", Type: LinkType("depends_on")},
		{To: "A", Type: LinkType("depends_on")},
	}

	data := FormatEdgesTOML(edges)
	got, err := ParseEdgesTOML(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(got))
	}
	if got[0].To != "A" || got[0].Type != "depends_on" {
		t.Fatalf("unexpected first edge: %+v", got[0])
	}
	if got[1].To != "B" || got[1].Type != "related" {
		t.Fatalf("unexpected second edge: %+v", got[1])
	}
}

func TestEdgeStoreInboundLookup(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(EdgesDir(root), 0o755); err != nil {
		t.Fatal(err)
	}

	store := NewEdgeStore(root)
	allowed := []LinkType{"depends_on", "related"}

	if err := store.AddOutbound("SRC1", Edge{To: "DST", Type: "depends_on"}, allowed); err != nil {
		t.Fatalf("add outbound failed: %v", err)
	}
	if err := store.AddOutbound("SRC2", Edge{To: "DST", Type: "related"}, allowed); err != nil {
		t.Fatalf("add outbound failed: %v", err)
	}

	inbound, err := store.FindInbound("DST")
	if err != nil {
		t.Fatalf("find inbound failed: %v", err)
	}
	if len(inbound) != 2 {
		t.Fatalf("expected 2 inbound edges, got %d", len(inbound))
	}
	if inbound[0].From != "SRC1" || inbound[0].Type != "depends_on" {
		t.Fatalf("unexpected first inbound edge: %+v", inbound[0])
	}
	if inbound[1].From != "SRC2" || inbound[1].Type != "related" {
		t.Fatalf("unexpected second inbound edge: %+v", inbound[1])
	}
}

func TestEdgeStoreRemoveOutboundKeepsRemainingEdges(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(EdgesDir(root), 0o755); err != nil {
		t.Fatal(err)
	}

	store := NewEdgeStore(root)
	allowed := []LinkType{"depends_on", "related"}
	if err := store.AddOutbound("SRC", Edge{To: "DST1", Type: "depends_on"}, allowed); err != nil {
		t.Fatalf("add outbound failed: %v", err)
	}
	if err := store.AddOutbound("SRC", Edge{To: "DST2", Type: "related"}, allowed); err != nil {
		t.Fatalf("add outbound failed: %v", err)
	}

	removed, err := store.RemoveOutbound("SRC", Edge{To: "DST1", Type: "depends_on"})
	if err != nil {
		t.Fatalf("remove outbound failed: %v", err)
	}
	if !removed {
		t.Fatal("expected one edge to be removed")
	}

	outbound, err := store.ListOutbound("SRC")
	if err != nil {
		t.Fatalf("list outbound failed: %v", err)
	}
	if len(outbound) != 1 {
		t.Fatalf("expected 1 remaining edge, got %d", len(outbound))
	}
	if outbound[0].To != "DST2" || outbound[0].Type != "related" {
		t.Fatalf("unexpected remaining edge: %+v", outbound[0])
	}
}

func TestValidateLinkType(t *testing.T) {
	allowed := []LinkType{"depends_on", "related"}
	if err := ValidateLinkType("depends_on", allowed); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := ValidateLinkType("unknown", allowed); err == nil {
		t.Fatal("expected validation error")
	}
}
