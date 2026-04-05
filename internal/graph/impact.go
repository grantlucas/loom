package graph

import "sort"

// Impact holds the downstream impact metrics for a single ready node.
type Impact struct {
	NodeID       string
	UnblockCount int
	PrioritySum  int
	MaxDepth     int
	OwnPriority  int
	Downstream   []string // transitively blocked node IDs
}

// DownstreamImpact computes transitive downstream impact for each ready node.
// For each ready node, it finds all transitively reachable successors and
// computes: unblock count, blocked priority sum (4-priority per node),
// max chain depth, and own priority.
// Results are sorted by composite score: PrioritySum desc, UnblockCount desc,
// MaxDepth desc, OwnPriority asc.
func DownstreamImpact(g *DAG, readyIDs []string, priorities map[string]int) []Impact {
	results := make([]Impact, 0, len(readyIDs))
	for _, id := range readyIDs {
		imp := Impact{
			NodeID:      id,
			OwnPriority: priorities[id],
		}

		// BFS to collect all transitive successors
		visited := map[string]bool{id: true}
		queue := []string{id}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			for _, succ := range g.Successors(cur) {
				if visited[succ] {
					continue
				}
				visited[succ] = true
				imp.Downstream = append(imp.Downstream, succ)
				imp.UnblockCount++
				imp.PrioritySum += 4 - priorities[succ]
				queue = append(queue, succ)
			}
		}

		// Compute max depth using longest-path DP over the reachable subgraph.
		// dist[node] = longest distance from id to node.
		if imp.UnblockCount > 0 {
			dist := map[string]int{id: 0}
			// Use topological BFS (Kahn's) restricted to the reachable set.
			inDeg := make(map[string]int, len(visited))
			for node := range visited {
				inDeg[node] = 0
			}
			for node := range visited {
				for _, succ := range g.Successors(node) {
					if visited[succ] {
						inDeg[succ]++
					}
				}
			}
			var topo []string
			for node, deg := range inDeg {
				if deg == 0 {
					topo = append(topo, node)
				}
			}
			for len(topo) > 0 {
				cur := topo[0]
				topo = topo[1:]
				for _, succ := range g.Successors(cur) {
					nd := dist[cur] + 1
					if nd > dist[succ] {
						dist[succ] = nd
					}
					inDeg[succ]--
					if inDeg[succ] == 0 {
						topo = append(topo, succ)
					}
				}
			}
			for _, d := range dist {
				if d > imp.MaxDepth {
					imp.MaxDepth = d
				}
			}
		}

		results = append(results, imp)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].PrioritySum != results[j].PrioritySum {
			return results[i].PrioritySum > results[j].PrioritySum
		}
		if results[i].UnblockCount != results[j].UnblockCount {
			return results[i].UnblockCount > results[j].UnblockCount
		}
		if results[i].MaxDepth != results[j].MaxDepth {
			return results[i].MaxDepth > results[j].MaxDepth
		}
		return results[i].OwnPriority < results[j].OwnPriority
	})

	return results
}
