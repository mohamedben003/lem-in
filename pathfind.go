package main
 
// ─────────────────────────────────────────────────────────────
// HOW IT WORKS (node-split max-flow)
//
// Problem: find k vertex-disjoint paths (no shared intermediate room).
//
// Trick: split every room into two nodes — "in" and "out".
//   room i  →  in-node  = 2*i
//              out-node = 2*i+1
//
// Add edges:
//   in(i) ──cap 1──► out(i)   ← this enforces "one ant per room"
//   out(u) ──cap 1──► in(v)   ← for every tunnel u-v (both directions)
//
// Start and end rooms get cap = numAnts (unlimited flow through them).
//
// Run Edmonds-Karp (BFS max-flow) to push as many flows as possible.
// Each unit of flow = one path through the colony.
//
// Then try k = 1, 2, ..., maxK paths and pick whichever finishes in fewest turns.
// ─────────────────────────────────────────────────────────────
 
// edge represents a directed edge in the flow graph.
type edge struct {
	to      int // destination node
	rev     int // index of the reverse edge in adj[to]
	cap     int // remaining capacity
	origCap int // original capacity (used to detect where flow went)
}
 
// addEdge adds a directed edge u→v with the given capacity,
// plus its reverse edge (capacity 0) for the residual graph.
func addEdge(adj [][]edge, u, v, cap int) {
	adj[u] = append(adj[u], edge{v, len(adj[v]), cap, cap})
	adj[v] = append(adj[v], edge{u, len(adj[u]) - 1, 0, 0})
}
 
// bfsAugment finds one augmenting path from s to t and pushes one unit of flow.
// Returns true if a path was found (i.e. flow was pushed).
func bfsAugment(adj [][]edge, s, t int) bool {
	n := len(adj)
	prev := make([]int, n)
	prevEdge := make([]int, n)
	for i := range prev {
		prev[i] = -1
	}
	prev[s] = s
	queue := []int{s}
 
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for ei, e := range adj[u] {
			if e.cap > 0 && prev[e.to] < 0 {
				prev[e.to] = u
				prevEdge[e.to] = ei
				if e.to == t {
					// Trace back and decrement capacity along the path.
					for v := t; v != s; {
						u2 := prev[v]
						ei2 := prevEdge[v]
						adj[u2][ei2].cap--
						adj[v][adj[u2][ei2].rev].cap++
						v = u2
					}
					return true
				}
				queue = append(queue, e.to)
			}
		}
	}
	return false
}
 
// buildGraph constructs the node-split flow graph for the colony.
func buildGraph(colony *Colony, roomIndex map[string]int) [][]edge {
	n := len(roomIndex)
	adj := make([][]edge, 2*n)
	for i := range adj {
		adj[i] = []edge{}
	}
 
	// Internal edges: in(i) → out(i)
	for name, i := range roomIndex {
		cap := 1
		if name == colony.start || name == colony.end {
			cap = colony.numAnts + 1 // start/end have unlimited capacity
		}
		addEdge(adj, 2*i, 2*i+1, cap)
	}
 
	// Tunnel edges: out(u) → in(v) and out(v) → in(u), independently.
	// We track pairs to avoid adding duplicates.
	added := map[[2]int]bool{}
	for u, neighbors := range colony.links {
		ui := roomIndex[u]
		for _, v := range neighbors {
			vi := roomIndex[v]
			lo, hi := ui, vi
			if lo > hi {
				lo, hi = hi, lo
			}
			if added[[2]int{lo, hi}] {
				continue
			}
			added[[2]int{lo, hi}] = true
			addEdge(adj, 2*ui+1, 2*vi, 1)   // u → v
			addEdge(adj, 2*vi+1, 2*ui, 1)   // v → u
		}
	}
	return adj
}
 
// recoverPaths traces which edges carried flow and reconstructs the k paths.
// A flow-carrying forward edge has origCap > 0 and cap < origCap.
func recoverPaths(adj [][]edge, colony *Colony, roomIndex map[string]int, k int) [][]string {
	// Build a reverse lookup: index → room name.
	names := make([]string, len(roomIndex))
	for name, i := range roomIndex {
		names[i] = name
	}
 
	si := roomIndex[colony.start]
	ei := roomIndex[colony.end]
 
	// For each node, collect all out-neighbours that received flow.
	usedEdges := make([][]int, len(adj))
	for u, edges := range adj {
		for _, e := range edges {
			if e.origCap > 0 && e.cap < e.origCap {
				usedEdges[u] = append(usedEdges[u], e.to)
			}
		}
	}
 
	// src = out(start), snk = in(end)
	src := 2*si + 1
	snk := 2 * ei
 
	var paths [][]string
	for p := 0; p < k; p++ {
		path := traceSinglePath(usedEdges, names, src, snk, si, ei)
		if path == nil {
			break
		}
		paths = append(paths, path)
	}
	return paths
}
 
// traceSinglePath follows consumed edges from src to snk once,
// consuming them so the next call to traceSinglePath finds a different path.
//
// Graph traversal pattern:
//   out(start) → in(v) → out(v) → in(w) → ... → in(end)
// Even node index = in-node, odd = out-node.
func traceSinglePath(usedEdges [][]int, names []string, src, snk, startI, endI int) []string {
	path := []string{names[startI]}
	cur := src
 
	for cur != snk {
		if len(usedEdges[cur]) == 0 {
			return nil // dead end — shouldn't happen if flow is consistent
		}
		next := usedEdges[cur][0]
		usedEdges[cur] = usedEdges[cur][1:]
 
		roomI := next / 2
		if next%2 == 0 {
			// Arrived at in(v) via a tunnel edge — we've entered a new room.
			path = append(path, names[roomI])
			if next == snk {
				break
			}
			// Cross the internal edge in(v) → out(v).
			if len(usedEdges[next]) == 0 {
				return nil
			}
			cur = usedEdges[next][0]
			usedEdges[next] = usedEdges[next][1:]
		} else {
			// Landed on out(v) directly — just move.
			cur = next
		}
	}
 
	if len(path) < 2 || path[len(path)-1] != names[endI] {
		return nil
	}
	return path
}
 
// findPaths returns the set of vertex-disjoint paths that minimises total turns.
func findPaths(colony *Colony) [][]string {
	// Assign a numeric index to every room.
	roomIndex := make(map[string]int, len(colony.rooms))
	i := 0
	for name := range colony.rooms {
		roomIndex[name] = i
		i++
	}
 
	si := roomIndex[colony.start]
	ei := roomIndex[colony.end]
	src := 2*si + 1
	snk := 2 * ei
 
	// Find the maximum number of vertex-disjoint paths (max flow).
	adj := buildGraph(colony, roomIndex)
	maxK := 0
	for bfsAugment(adj, src, snk) {
		maxK++
	}
	if maxK == 0 {
		return nil
	}
 
	// Try k = 1..maxK and keep whichever yields fewest turns.
	bestPaths := [][]string{}
	bestTurns := -1
 
	for k := 1; k <= maxK; k++ {
		adj = buildGraph(colony, roomIndex)
		for j := 0; j < k; j++ {
			bfsAugment(adj, src, snk)
		}
		paths := recoverPaths(adj, colony, roomIndex, k)
		if len(paths) != k {
			break
		}
		turns := calcTurns(colony.numAnts, paths)
		if bestTurns < 0 || turns < bestTurns {
			bestTurns = turns
			bestPaths = paths
		} else {
			break // adding more paths only makes things worse from here
		}
	}
	return bestPaths
}
 
// ── Helpers ──────────────────────────────────────────────────
 
// calcTurns simulates ant distribution and returns total turns needed.
func calcTurns(numAnts int, paths [][]string) int {
	sorted := sortedByLength(paths)
	counts := distributeAnts(numAnts, sorted)
	max := 0
	for i, c := range counts {
		if c > 0 {
			if t := c + len(sorted[i]) - 2; t > max {
				max = t
			}
		}
	}
	return max
}
 
// sortedByLength returns a copy of paths sorted shortest-first.
func sortedByLength(paths [][]string) [][]string {
	s := make([][]string, len(paths))
	copy(s, paths)
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if len(s[i]) > len(s[j]) {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
	return s
}
 
// distributeAnts greedily assigns each ant to the path that minimises its finish turn.
func distributeAnts(numAnts int, paths [][]string) []int {
	counts := make([]int, len(paths))
	for i := 0; i < numAnts; i++ {
		best, bestTurn := 0, -1
		for j, p := range paths {
			// finish turn for the next ant on this path
			t := counts[j] + len(p) - 1
			if bestTurn < 0 || t < bestTurn {
				best, bestTurn = j, t
			}
		}
		counts[best]++
	}
	return counts
}