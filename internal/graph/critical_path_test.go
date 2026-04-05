package graph

import (
	"testing"
)

func TestCriticalPaths_Linear(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")
	priorities := map[string]int{"a": 1, "b": 2, "c": 1}

	chains := CriticalPaths(g, priorities)
	if len(chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(chains))
	}
	if len(chains[0].Nodes) != 3 {
		t.Errorf("expected chain length 3, got %d", len(chains[0].Nodes))
	}
	if chains[0].Nodes[0] != "a" || chains[0].Nodes[2] != "c" {
		t.Errorf("expected [a b c], got %v", chains[0].Nodes)
	}
}

func TestCriticalPaths_Diamond(t *testing.T) {
	// a -> b -> d
	// a -> c -> d
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	g.AddEdge("b", "d")
	g.AddEdge("c", "d")
	priorities := map[string]int{"a": 1, "b": 1, "c": 1, "d": 1}

	chains := CriticalPaths(g, priorities)
	// One sink (d), longest path is 3 (a->b->d or a->c->d)
	if len(chains) != 1 {
		t.Fatalf("expected 1 chain (one sink), got %d", len(chains))
	}
	if len(chains[0].Nodes) != 3 {
		t.Errorf("expected chain length 3, got %d", len(chains[0].Nodes))
	}
	if chains[0].Nodes[0] != "a" || chains[0].Nodes[2] != "d" {
		t.Errorf("expected chain starting at a ending at d, got %v", chains[0].Nodes)
	}
}

func TestCriticalPaths_MultipleSinks(t *testing.T) {
	// a -> b -> c (length 3)
	// a -> d     (length 2)
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")
	g.AddEdge("a", "d")
	priorities := map[string]int{"a": 1, "b": 1, "c": 1, "d": 1}

	chains := CriticalPaths(g, priorities)
	if len(chains) != 2 {
		t.Fatalf("expected 2 chains (two sinks), got %d", len(chains))
	}
	// Sorted by length desc
	if len(chains[0].Nodes) != 3 {
		t.Errorf("expected first chain length 3, got %d", len(chains[0].Nodes))
	}
	if len(chains[1].Nodes) != 2 {
		t.Errorf("expected second chain length 2, got %d", len(chains[1].Nodes))
	}
}

func TestCriticalPaths_SortByPriority(t *testing.T) {
	// Two independent chains of equal length
	// x -> y (max priority = min(3,3) = 3)
	// a -> b (max priority = min(1,1) = 1)
	g := NewDAG()
	g.AddEdge("x", "y")
	g.AddEdge("a", "b")
	priorities := map[string]int{"x": 3, "y": 3, "a": 1, "b": 1}

	chains := CriticalPaths(g, priorities)
	if len(chains) != 2 {
		t.Fatalf("expected 2 chains, got %d", len(chains))
	}
	// Same length, so sort by max priority (lower number = higher priority = first)
	if chains[0].MaxPriority != 1 {
		t.Errorf("expected first chain max priority 1 (highest), got %d", chains[0].MaxPriority)
	}
	if chains[1].MaxPriority != 3 {
		t.Errorf("expected second chain max priority 3, got %d", chains[1].MaxPriority)
	}
}

func TestCriticalPaths_Empty(t *testing.T) {
	g := NewDAG()
	chains := CriticalPaths(g, nil)
	if len(chains) != 0 {
		t.Errorf("expected no chains in empty graph, got %d", len(chains))
	}
}

func TestCriticalPaths_SingleNode(t *testing.T) {
	g := NewDAG()
	g.AddNode("a")
	priorities := map[string]int{"a": 1}

	chains := CriticalPaths(g, priorities)
	if len(chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(chains))
	}
	if len(chains[0].Nodes) != 1 || chains[0].Nodes[0] != "a" {
		t.Errorf("expected [a], got %v", chains[0].Nodes)
	}
}

func TestCriticalPaths_NilPriorities(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	chains := CriticalPaths(g, nil)
	if len(chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(chains))
	}
	// With nil priorities, MaxPriority should default to 0
	if chains[0].MaxPriority != 0 {
		t.Errorf("expected max priority 0 with nil map, got %d", chains[0].MaxPriority)
	}
}

func TestChain_Length(t *testing.T) {
	c := Chain{Nodes: []string{"a", "b", "c"}}
	if c.Length() != 3 {
		t.Errorf("expected length 3, got %d", c.Length())
	}
}

func TestCriticalPaths_LongestPathChosen(t *testing.T) {
	// Two paths to same sink:
	// a -> b -> c -> d (length 4)
	// a -> d          (length 2)
	// Should report the longest path to sink d
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")
	g.AddEdge("c", "d")
	g.AddEdge("a", "d")
	priorities := map[string]int{"a": 1, "b": 1, "c": 1, "d": 1}

	chains := CriticalPaths(g, priorities)
	if len(chains) != 1 {
		t.Fatalf("expected 1 chain (one sink), got %d", len(chains))
	}
	if len(chains[0].Nodes) != 4 {
		t.Errorf("expected longest path length 4, got %d: %v", len(chains[0].Nodes), chains[0].Nodes)
	}
}

func TestCriticalPaths_MaxPriorityOnChain(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")
	priorities := map[string]int{"a": 2, "b": 0, "c": 3}

	chains := CriticalPaths(g, priorities)
	if len(chains) != 1 {
		t.Fatalf("expected 1 chain, got %d", len(chains))
	}
	// Max priority is the minimum value (0 = P0 = highest priority)
	if chains[0].MaxPriority != 0 {
		t.Errorf("expected max priority 0, got %d", chains[0].MaxPriority)
	}
}

func TestCriticalPaths_CyclicGraph(t *testing.T) {
	g := NewDAG()
	g.addEdgeUnchecked("a", "b")
	g.addEdgeUnchecked("b", "a")
	chains := CriticalPaths(g, nil)
	if len(chains) != 0 {
		t.Errorf("expected no chains for cyclic graph, got %d", len(chains))
	}
}

func TestCriticalPaths_DisconnectedComponents(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("x", "y")
	g.AddEdge("y", "z")
	priorities := map[string]int{"a": 1, "b": 1, "x": 1, "y": 1, "z": 1}

	chains := CriticalPaths(g, priorities)
	// Two sinks: b and z
	if len(chains) != 2 {
		t.Fatalf("expected 2 chains, got %d", len(chains))
	}
	// x->y->z is longer (3) than a->b (2)
	if len(chains[0].Nodes) != 3 {
		t.Errorf("expected first chain length 3, got %d", len(chains[0].Nodes))
	}
	if len(chains[1].Nodes) != 2 {
		t.Errorf("expected second chain length 2, got %d", len(chains[1].Nodes))
	}
}
