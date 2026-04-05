package graph

import (
	"testing"
)

func TestDownstreamImpact_BasicChain(t *testing.T) {
	// a (ready) -> b -> c
	// a should unblock 2 issues, with priority sum and max depth
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")

	priorities := map[string]int{"a": 1, "b": 2, "c": 0}
	results := DownstreamImpact(g, []string{"a"}, priorities)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.NodeID != "a" {
		t.Errorf("expected node 'a', got %q", r.NodeID)
	}
	if r.UnblockCount != 2 {
		t.Errorf("expected unblock count 2, got %d", r.UnblockCount)
	}
	// Priority sum = (4-2) + (4-0) = 2 + 4 = 6
	if r.PrioritySum != 6 {
		t.Errorf("expected priority sum 6, got %d", r.PrioritySum)
	}
	// Max chain depth = 2 (a->b->c, 2 edges from a)
	if r.MaxDepth != 2 {
		t.Errorf("expected max depth 2, got %d", r.MaxDepth)
	}
}
