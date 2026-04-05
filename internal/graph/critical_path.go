package graph

import "sort"

// Chain represents a longest path from a source to a sink in the DAG.
type Chain struct {
	Nodes       []string
	MaxPriority int // lowest priority value (0=P0=highest) among nodes
}

// Length returns the number of nodes in the chain.
func (c Chain) Length() int {
	return len(c.Nodes)
}

// CriticalPaths finds the longest path from any source to each sink in the DAG.
// Results are sorted by length descending, then by max priority ascending
// (lower number = higher priority = sorted first).
// The priorities map gives each node's priority value; missing entries default to 0.
func CriticalPaths(g *DAG, priorities map[string]int) []Chain {
	if g.NodeCount() == 0 {
		return nil
	}

	// Compute longest distance and predecessor on longest path for each node
	// using topological order (dynamic programming).
	order, err := g.TopoSort()
	if err != nil {
		return nil // cyclic graph, no valid paths
	}

	dist := make(map[string]int, len(order))
	prev := make(map[string]string, len(order))
	for _, node := range order {
		dist[node] = 0
	}

	for _, node := range order {
		for succ := range g.successors[node] {
			newDist := dist[node] + 1
			if newDist > dist[succ] {
				dist[succ] = newDist
				prev[succ] = node
			}
		}
	}

	// For each sink, reconstruct the longest path back to a source
	sinks := g.Sinks()
	chains := make([]Chain, 0, len(sinks))
	for _, sink := range sinks {
		// Trace back from sink to source
		var path []string
		for cur := sink; cur != ""; cur = prev[cur] {
			path = append(path, cur)
		}
		// Reverse to get source -> sink order
		for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
			path[i], path[j] = path[j], path[i]
		}

		maxPri := maxPriorityOf(path, priorities)
		chains = append(chains, Chain{Nodes: path, MaxPriority: maxPri})
	}

	sort.Slice(chains, func(i, j int) bool {
		if chains[i].Length() != chains[j].Length() {
			return chains[i].Length() > chains[j].Length()
		}
		return chains[i].MaxPriority < chains[j].MaxPriority
	})

	return chains
}

// maxPriorityOf returns the minimum priority value across nodes
// (lower number = higher priority).
func maxPriorityOf(nodes []string, priorities map[string]int) int {
	minVal := priorities[nodes[0]]
	for _, n := range nodes[1:] {
		if p := priorities[n]; p < minVal {
			minVal = p
		}
	}
	return minVal
}
