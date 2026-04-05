package graph

import (
	"testing"
)

func TestDownstreamImpact_SortedByCompositeScore(t *testing.T) {
	// Three ready nodes with different impacts:
	// x (ready) -> leaf (unblocks nothing)
	// a (ready) -> b (P0) — priority sum = 4, unblock = 1
	// c (ready) -> d (P2) -> e (P2) — priority sum = 4, unblock = 2 (tiebreak by unblock count)
	g := NewDAG()
	g.AddNode("x")
	g.AddEdge("a", "b")
	g.AddEdge("c", "d")
	g.AddEdge("d", "e")

	priorities := map[string]int{"x": 3, "a": 2, "b": 0, "c": 2, "d": 2, "e": 2}
	results := DownstreamImpact(g, []string{"x", "a", "c"}, priorities)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	// c should be first: priority sum 4, unblock 2
	// a should be second: priority sum 4, unblock 1
	// x should be last: priority sum 0, unblock 0
	if results[0].NodeID != "c" {
		t.Errorf("expected 'c' first (highest impact), got %q", results[0].NodeID)
	}
	if results[1].NodeID != "a" {
		t.Errorf("expected 'a' second, got %q", results[1].NodeID)
	}
	if results[2].NodeID != "x" {
		t.Errorf("expected 'x' last (leaf), got %q", results[2].NodeID)
	}
}

func TestDownstreamImpact_Empty(t *testing.T) {
	g := NewDAG()
	results := DownstreamImpact(g, nil, nil)
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty input, got %d", len(results))
	}
}

func TestDownstreamImpact_LeafNode(t *testing.T) {
	g := NewDAG()
	g.AddNode("a")
	priorities := map[string]int{"a": 2}
	results := DownstreamImpact(g, []string{"a"}, priorities)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].UnblockCount != 0 {
		t.Errorf("leaf should unblock nothing, got %d", results[0].UnblockCount)
	}
	if results[0].PrioritySum != 0 {
		t.Errorf("leaf should have priority sum 0, got %d", results[0].PrioritySum)
	}
	if results[0].MaxDepth != 0 {
		t.Errorf("leaf should have max depth 0, got %d", results[0].MaxDepth)
	}
	if results[0].OwnPriority != 2 {
		t.Errorf("expected own priority 2, got %d", results[0].OwnPriority)
	}
}

func TestDownstreamImpact_Downstream(t *testing.T) {
	// a (ready) -> b -> c
	// a (ready) -> d
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")
	g.AddEdge("a", "d")
	priorities := map[string]int{"a": 1, "b": 1, "c": 1, "d": 1}
	results := DownstreamImpact(g, []string{"a"}, priorities)
	if results[0].Downstream == nil {
		t.Fatal("expected non-nil Downstream slice")
	}
	if len(results[0].Downstream) != 3 {
		t.Errorf("expected 3 downstream nodes, got %d", len(results[0].Downstream))
	}
}

func TestDownstreamImpact_DiamondDoesNotDoubleCount(t *testing.T) {
	// a -> b -> d
	// a -> c -> d
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	g.AddEdge("b", "d")
	g.AddEdge("c", "d")
	priorities := map[string]int{"a": 1, "b": 1, "c": 1, "d": 1}
	results := DownstreamImpact(g, []string{"a"}, priorities)
	// d should only be counted once
	if results[0].UnblockCount != 3 {
		t.Errorf("expected unblock count 3 (b,c,d each once), got %d", results[0].UnblockCount)
	}
}

func TestDownstreamImpact_MaxDepthThroughLongestPath(t *testing.T) {
	// a -> b -> d (depth 2)
	// a -> c -> d -> e (depth 3 via c)
	// But since d is shared, max depth should be 3 (a->c->d->e)
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	g.AddEdge("b", "d")
	g.AddEdge("c", "d")
	g.AddEdge("d", "e")
	priorities := map[string]int{"a": 1, "b": 1, "c": 1, "d": 1, "e": 1}
	results := DownstreamImpact(g, []string{"a"}, priorities)
	if results[0].MaxDepth != 3 {
		t.Errorf("expected max depth 3, got %d", results[0].MaxDepth)
	}
}

func TestDownstreamImpact_MaxDepthWithShortcut(t *testing.T) {
	// a -> d directly (depth 1) AND a -> b -> c -> d (depth 3)
	// Max depth should be 3, not 1
	g := NewDAG()
	g.AddEdge("a", "d") // shortcut
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")
	g.AddEdge("c", "d")
	priorities := map[string]int{"a": 1, "b": 1, "c": 1, "d": 1}
	results := DownstreamImpact(g, []string{"a"}, priorities)
	if results[0].MaxDepth != 3 {
		t.Errorf("expected max depth 3 (longest path), got %d", results[0].MaxDepth)
	}
}

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
