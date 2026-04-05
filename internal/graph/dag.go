package graph

import "fmt"

// DAG is a directed acyclic graph with string node IDs.
type DAG struct {
	successors   map[string]map[string]bool
	predecessors map[string]map[string]bool
}

// NewDAG creates an empty DAG.
func NewDAG() *DAG {
	return &DAG{
		successors:   make(map[string]map[string]bool),
		predecessors: make(map[string]map[string]bool),
	}
}

// AddNode adds a node if it doesn't already exist.
func (g *DAG) AddNode(id string) {
	if _, ok := g.successors[id]; !ok {
		g.successors[id] = make(map[string]bool)
		g.predecessors[id] = make(map[string]bool)
	}
}

// HasNode returns true if the node exists.
func (g *DAG) HasNode(id string) bool {
	_, ok := g.successors[id]
	return ok
}

// NodeCount returns the number of nodes.
func (g *DAG) NodeCount() int {
	return len(g.successors)
}

// Nodes returns all node IDs.
func (g *DAG) Nodes() []string {
	nodes := make([]string, 0, len(g.successors))
	for id := range g.successors {
		nodes = append(nodes, id)
	}
	return nodes
}

// AddEdge adds a directed edge from -> to. Creates nodes if needed.
// Returns an error for self-loops.
func (g *DAG) AddEdge(from, to string) error {
	if from == to {
		return fmt.Errorf("self-loop not allowed: %s", from)
	}
	g.AddNode(from)
	g.AddNode(to)
	g.successors[from][to] = true
	g.predecessors[to][from] = true
	return nil
}

// addEdgeUnchecked adds an edge without self-loop validation.
// Used for testing cycle detection.
func (g *DAG) addEdgeUnchecked(from, to string) {
	g.AddNode(from)
	g.AddNode(to)
	g.successors[from][to] = true
	g.predecessors[to][from] = true
}

// EdgeCount returns the total number of edges.
func (g *DAG) EdgeCount() int {
	count := 0
	for _, succs := range g.successors {
		count += len(succs)
	}
	return count
}

// Successors returns the direct successors of a node.
// Returns nil if the node doesn't exist.
func (g *DAG) Successors(id string) []string {
	succs, ok := g.successors[id]
	if !ok {
		return nil
	}
	result := make([]string, 0, len(succs))
	for s := range succs {
		result = append(result, s)
	}
	return result
}

// Predecessors returns the direct predecessors of a node.
// Returns nil if the node doesn't exist.
func (g *DAG) Predecessors(id string) []string {
	preds, ok := g.predecessors[id]
	if !ok {
		return nil
	}
	result := make([]string, 0, len(preds))
	for p := range preds {
		result = append(result, p)
	}
	return result
}

// Roots returns all nodes with no predecessors.
func (g *DAG) Roots() []string {
	var roots []string
	for id, preds := range g.predecessors {
		if len(preds) == 0 {
			roots = append(roots, id)
		}
	}
	return roots
}

// Sinks returns all nodes with no successors.
func (g *DAG) Sinks() []string {
	var sinks []string
	for id, succs := range g.successors {
		if len(succs) == 0 {
			sinks = append(sinks, id)
		}
	}
	return sinks
}

// TopoSort returns a topological ordering of all nodes using Kahn's algorithm.
// Returns an error if the graph contains a cycle.
func (g *DAG) TopoSort() ([]string, error) {
	// Compute in-degrees
	inDegree := make(map[string]int, len(g.successors))
	for id := range g.successors {
		inDegree[id] = len(g.predecessors[id])
	}

	// Start with all roots (in-degree 0)
	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	var order []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)
		for succ := range g.successors[node] {
			inDegree[succ]--
			if inDegree[succ] == 0 {
				queue = append(queue, succ)
			}
		}
	}

	if len(order) != len(g.successors) {
		return nil, fmt.Errorf("graph contains a cycle")
	}
	return order, nil
}

// DetectCycle returns a cycle as a slice of node IDs, or nil if acyclic.
// Uses DFS with three-color marking.
func (g *DAG) DetectCycle() []string {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[string]int, len(g.successors))
	parent := make(map[string]string, len(g.successors))

	var dfs func(node string) []string
	dfs = func(node string) []string {
		color[node] = gray
		for succ := range g.successors[node] {
			if color[succ] == gray {
				// Found cycle, reconstruct
				cycle := []string{succ, node}
				for cur := node; cur != succ; {
					cur = parent[cur]
					cycle = append(cycle, cur)
				}
				// Reverse to get forward order
				for i, j := 0, len(cycle)-1; i < j; i, j = i+1, j-1 {
					cycle[i], cycle[j] = cycle[j], cycle[i]
				}
				return cycle
			}
			if color[succ] == white {
				parent[succ] = node
				if c := dfs(succ); c != nil {
					return c
				}
			}
		}
		color[node] = black
		return nil
	}

	for node := range g.successors {
		if color[node] == white {
			if c := dfs(node); c != nil {
				return c
			}
		}
	}
	return nil
}
