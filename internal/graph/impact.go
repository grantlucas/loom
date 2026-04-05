package graph

// Impact holds the downstream impact metrics for a single ready node.
type Impact struct {
	NodeID      string
	UnblockCount int
	PrioritySum  int
	MaxDepth     int
	OwnPriority  int
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

		// BFS to find all transitive successors and max depth
		visited := map[string]bool{id: true}
		type entry struct {
			node  string
			depth int
		}
		queue := []entry{{id, 0}}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			for _, succ := range g.Successors(cur.node) {
				if visited[succ] {
					continue
				}
				visited[succ] = true
				d := cur.depth + 1
				imp.UnblockCount++
				imp.PrioritySum += 4 - priorities[succ]
				if d > imp.MaxDepth {
					imp.MaxDepth = d
				}
				queue = append(queue, entry{succ, d})
			}
		}

		results = append(results, imp)
	}
	return results
}
