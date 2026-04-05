package graph

import (
	"sort"
	"testing"
)

func TestNewDAG_ReturnsNonNil(t *testing.T) {
	g := NewDAG()
	if g == nil {
		t.Fatal("NewDAG should return a non-nil pointer")
	}
}

func TestDAG_AddNode(t *testing.T) {
	g := NewDAG()
	g.AddNode("a")
	if !g.HasNode("a") {
		t.Error("expected node 'a' to exist")
	}
}

func TestDAG_AddNode_Duplicate(t *testing.T) {
	g := NewDAG()
	g.AddNode("a")
	g.AddNode("a")
	if g.NodeCount() != 1 {
		t.Errorf("expected 1 node, got %d", g.NodeCount())
	}
}

func TestDAG_HasNode_Missing(t *testing.T) {
	g := NewDAG()
	if g.HasNode("x") {
		t.Error("expected HasNode to return false for missing node")
	}
}

func TestDAG_NodeCount_Empty(t *testing.T) {
	g := NewDAG()
	if g.NodeCount() != 0 {
		t.Errorf("expected 0 nodes, got %d", g.NodeCount())
	}
}

func TestDAG_AddEdge(t *testing.T) {
	g := NewDAG()
	g.AddNode("a")
	g.AddNode("b")
	err := g.AddEdge("a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDAG_AddEdge_AutoCreatesNodes(t *testing.T) {
	g := NewDAG()
	err := g.AddEdge("a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !g.HasNode("a") || !g.HasNode("b") {
		t.Error("AddEdge should auto-create missing nodes")
	}
}

func TestDAG_AddEdge_SelfLoop(t *testing.T) {
	g := NewDAG()
	g.AddNode("a")
	err := g.AddEdge("a", "a")
	if err == nil {
		t.Error("expected error for self-loop")
	}
}

func TestDAG_AddEdge_Duplicate(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	err := g.AddEdge("a", "b")
	if err != nil {
		t.Error("duplicate edge should be a no-op, not an error")
	}
}

func TestDAG_EdgeCount(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")
	if g.EdgeCount() != 2 {
		t.Errorf("expected 2 edges, got %d", g.EdgeCount())
	}
}

func TestDAG_EdgeCount_NoDuplicates(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("a", "b")
	if g.EdgeCount() != 1 {
		t.Errorf("expected 1 edge after duplicate add, got %d", g.EdgeCount())
	}
}

// --- Neighbors ---

func TestDAG_Successors(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	succ := g.Successors("a")
	sort.Strings(succ)
	if len(succ) != 2 || succ[0] != "b" || succ[1] != "c" {
		t.Errorf("expected [b c], got %v", succ)
	}
}

func TestDAG_Successors_Empty(t *testing.T) {
	g := NewDAG()
	g.AddNode("a")
	succ := g.Successors("a")
	if len(succ) != 0 {
		t.Errorf("expected no successors, got %v", succ)
	}
}

func TestDAG_Successors_MissingNode(t *testing.T) {
	g := NewDAG()
	succ := g.Successors("x")
	if succ != nil {
		t.Errorf("expected nil for missing node, got %v", succ)
	}
}

func TestDAG_Predecessors(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "c")
	g.AddEdge("b", "c")
	pred := g.Predecessors("c")
	sort.Strings(pred)
	if len(pred) != 2 || pred[0] != "a" || pred[1] != "b" {
		t.Errorf("expected [a b], got %v", pred)
	}
}

func TestDAG_Predecessors_Empty(t *testing.T) {
	g := NewDAG()
	g.AddNode("a")
	pred := g.Predecessors("a")
	if len(pred) != 0 {
		t.Errorf("expected no predecessors, got %v", pred)
	}
}

func TestDAG_Predecessors_MissingNode(t *testing.T) {
	g := NewDAG()
	pred := g.Predecessors("x")
	if pred != nil {
		t.Errorf("expected nil for missing node, got %v", pred)
	}
}

// --- Roots and Sinks ---

func TestDAG_Roots(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	g.AddEdge("b", "d")
	roots := g.Roots()
	if len(roots) != 1 || roots[0] != "a" {
		t.Errorf("expected [a], got %v", roots)
	}
}

func TestDAG_Roots_Multiple(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "c")
	g.AddEdge("b", "c")
	roots := g.Roots()
	sort.Strings(roots)
	if len(roots) != 2 || roots[0] != "a" || roots[1] != "b" {
		t.Errorf("expected [a b], got %v", roots)
	}
}

func TestDAG_Roots_Isolated(t *testing.T) {
	g := NewDAG()
	g.AddNode("a")
	roots := g.Roots()
	if len(roots) != 1 || roots[0] != "a" {
		t.Errorf("isolated node should be a root, got %v", roots)
	}
}

func TestDAG_Roots_Empty(t *testing.T) {
	g := NewDAG()
	roots := g.Roots()
	if len(roots) != 0 {
		t.Errorf("expected no roots in empty graph, got %v", roots)
	}
}

func TestDAG_Sinks(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	sinks := g.Sinks()
	sort.Strings(sinks)
	if len(sinks) != 2 || sinks[0] != "b" || sinks[1] != "c" {
		t.Errorf("expected [b c], got %v", sinks)
	}
}

func TestDAG_Sinks_Isolated(t *testing.T) {
	g := NewDAG()
	g.AddNode("a")
	sinks := g.Sinks()
	if len(sinks) != 1 || sinks[0] != "a" {
		t.Errorf("isolated node should be a sink, got %v", sinks)
	}
}

func TestDAG_Sinks_Empty(t *testing.T) {
	g := NewDAG()
	sinks := g.Sinks()
	if len(sinks) != 0 {
		t.Errorf("expected no sinks in empty graph, got %v", sinks)
	}
}

// --- Topological Sort ---

func TestDAG_TopoSort_Linear(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")
	order, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(order))
	}
	// a must come before b, b before c
	pos := indexMap(order)
	if pos["a"] >= pos["b"] || pos["b"] >= pos["c"] {
		t.Errorf("invalid order: %v", order)
	}
}

func TestDAG_TopoSort_Diamond(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	g.AddEdge("b", "d")
	g.AddEdge("c", "d")
	order, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pos := indexMap(order)
	if pos["a"] >= pos["b"] || pos["a"] >= pos["c"] {
		t.Errorf("a should come before b and c: %v", order)
	}
	if pos["b"] >= pos["d"] || pos["c"] >= pos["d"] {
		t.Errorf("b and c should come before d: %v", order)
	}
}

func TestDAG_TopoSort_Empty(t *testing.T) {
	g := NewDAG()
	order, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 0 {
		t.Errorf("expected empty order, got %v", order)
	}
}

func TestDAG_TopoSort_SingleNode(t *testing.T) {
	g := NewDAG()
	g.AddNode("a")
	order, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 1 || order[0] != "a" {
		t.Errorf("expected [a], got %v", order)
	}
}

func TestDAG_TopoSort_Disconnected(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddNode("c")
	order, err := g.TopoSort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(order))
	}
	pos := indexMap(order)
	if pos["a"] >= pos["b"] {
		t.Errorf("a should come before b: %v", order)
	}
}

// --- Cycle Detection ---

func TestDAG_DetectCycle_NoCycle(t *testing.T) {
	g := NewDAG()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")
	cycle := g.DetectCycle()
	if cycle != nil {
		t.Errorf("expected no cycle, got %v", cycle)
	}
}

func TestDAG_DetectCycle_DirectCycle(t *testing.T) {
	g := NewDAG()
	g.AddNode("a")
	g.AddNode("b")
	// Force a cycle by directly manipulating (AddEdge prevents self-loop,
	// but we need to test cycle detection with mutual edges)
	g.addEdgeUnchecked("a", "b")
	g.addEdgeUnchecked("b", "a")
	cycle := g.DetectCycle()
	if cycle == nil {
		t.Fatal("expected cycle to be detected")
	}
	if len(cycle) < 2 {
		t.Errorf("cycle should have at least 2 nodes, got %v", cycle)
	}
}

func TestDAG_DetectCycle_LongCycle(t *testing.T) {
	g := NewDAG()
	g.addEdgeUnchecked("a", "b")
	g.addEdgeUnchecked("b", "c")
	g.addEdgeUnchecked("c", "a")
	cycle := g.DetectCycle()
	if cycle == nil {
		t.Fatal("expected cycle to be detected")
	}
}

func TestDAG_DetectCycle_Empty(t *testing.T) {
	g := NewDAG()
	cycle := g.DetectCycle()
	if cycle != nil {
		t.Errorf("expected no cycle in empty graph, got %v", cycle)
	}
}

func TestDAG_TopoSort_WithCycle_ReturnsError(t *testing.T) {
	g := NewDAG()
	g.addEdgeUnchecked("a", "b")
	g.addEdgeUnchecked("b", "a")
	_, err := g.TopoSort()
	if err == nil {
		t.Error("expected error for cyclic graph")
	}
}

// --- Nodes ---

func TestDAG_Nodes(t *testing.T) {
	g := NewDAG()
	g.AddNode("c")
	g.AddNode("a")
	g.AddNode("b")
	nodes := g.Nodes()
	sort.Strings(nodes)
	if len(nodes) != 3 || nodes[0] != "a" || nodes[1] != "b" || nodes[2] != "c" {
		t.Errorf("expected [a b c], got %v", nodes)
	}
}

// helper
func indexMap(order []string) map[string]int {
	m := make(map[string]int, len(order))
	for i, v := range order {
		m[v] = i
	}
	return m
}
