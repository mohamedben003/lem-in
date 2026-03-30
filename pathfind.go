package main

// Optimal path finder using Edmonds-Karp max-flow on a directed node-split graph.
//
// Key insight for undirected graphs: each undirected tunnel (u,v) becomes
// TWO directed tunnel edges. In the node-split graph:
//   out(u) -> in(v)  and  out(v) -> in(u)
// We only add forward capacity, not backward, so the residual clearly shows
// which direction flow went.
//
// After pushing k flows, we extract paths by DFS following only edges where
// origCap > 0 AND cap < origCap (i.e., flow was consumed on this edge).

type edge struct {
	to, rev  int
	cap, orig int
}

func newAdj(n int) [][]edge { return make([][]edge, n) }

func addEdge(adj [][]edge, u, v, cap int) {
	adj[u] = append(adj[u], edge{v, len(adj[v]), cap, cap})
	adj[v] = append(adj[v], edge{u, len(adj[u]) - 1, 0, 0})
}

func bfsAugment(adj [][]edge, s, t int) bool {
	n := len(adj)
	prev := make([]int, n)
	prevE := make([]int, n)
	for i := range prev { prev[i] = -1 }
	prev[s] = s
	q := []int{s}
	for len(q) > 0 {
		u := q[0]; q = q[1:]
		for ei, e := range adj[u] {
			if e.cap > 0 && prev[e.to] < 0 {
				prev[e.to] = u
				prevE[e.to] = ei
				if e.to == t {
					for v := t; v != s; {
						u2 := prev[v]; ei2 := prevE[v]
						adj[u2][ei2].cap--
						adj[v][adj[u2][ei2].rev].cap++
						v = u2
					}
					return true
				}
				q = append(q, e.to)
			}
		}
	}
	return false
}

func buildAdj(colony *Colony, idx map[string]int, bigCap int) [][]edge {
	n := len(idx)
	adj := newAdj(2 * n)
	// Internal edge: in(i) -> out(i)
	for name, i := range idx {
		cap := 1
		if name == colony.start || name == colony.end {
			cap = bigCap
		}
		addEdge(adj, 2*i, 2*i+1, cap)
	}
	// Tunnel edges: directed BOTH ways separately (not as undirected pairs)
	// We add out(u)->in(v) AND out(v)->in(u) as separate forward edges each with
	// their own reverse (residual) edge.
	seen := map[[2]int]bool{}
	for u, neighbors := range colony.links {
		ui := idx[u]
		for _, v := range neighbors {
			vi := idx[v]
			lo, hi := ui, vi
			if lo > hi { lo, hi = hi, lo }
			key := [2]int{lo, hi}
			if seen[key] { continue }
			seen[key] = true
			// Add both directions as independent forward edges
			addEdge(adj, 2*ui+1, 2*vi,   1)
			addEdge(adj, 2*vi+1, 2*ui,   1)
		}
	}
	return adj
}

// extractPaths traces k paths by following consumed forward edges (orig>0, cap<orig).
func extractPaths(adj [][]edge, colony *Colony, idx map[string]int, k int) [][]string {
	names := make([]string, len(idx))
	for name, i := range idx { names[i] = name }
	si := idx[colony.start]
	ei := idx[colony.end]

	// used[u] = list of neighbours v where flow went u->v on a forward edge
	used := make([][]int, len(adj))
	for u, edges := range adj {
		for _, e := range edges {
			if e.orig > 0 && e.cap < e.orig {
				used[u] = append(used[u], e.to)
			}
		}
	}

	src := 2*si + 1 // out(start)
	snk := 2 * ei   // in(end)

	var paths [][]string
	for p := 0; p < k; p++ {
		path := tracePath(used, names, src, snk, si, ei)
		if path == nil { break }
		paths = append(paths, path)
	}
	return paths
}

// tracePath follows consumed edges from src to snk.
// Graph structure: src=out(start) → in(v) → out(v) → in(w) → ... → in(end)=snk
func tracePath(used [][]int, names []string, src, snk, startI, endI int) []string {
	path := []string{names[startI]}
	cur := src
	for cur != snk {
		if len(used[cur]) == 0 { return nil }
		next := used[cur][0]
		used[cur] = used[cur][1:]

		nodeI := next / 2
		if next%2 == 0 { // in(v) — arrived at a new room via tunnel
			path = append(path, names[nodeI])
			if next == snk { break }
			// Cross internal edge in(v)->out(v)
			if len(used[next]) == 0 { return nil }
			cur = used[next][0]
			used[next] = used[next][1:]
		} else { // out(v) — shouldn't normally occur mid-trace
			cur = next
		}
	}
	if len(path) < 2 || path[len(path)-1] != names[endI] { return nil }
	return path
}

// findPaths finds the k vertex-disjoint paths minimising total turns.
func findPaths(colony *Colony) [][]string {
	idx := make(map[string]int, len(colony.rooms))
	i := 0
	for name := range colony.rooms { idx[name] = i; i++ }

	si := idx[colony.start]
	ei := idx[colony.end]
	bigCap := colony.numAnts + 1

	// Use the main graph to know the max possible k
	adjMain := buildAdj(colony, idx, bigCap)
	s, t := 2*si+1, 2*ei
	maxK := 0
	for bfsAugment(adjMain, s, t) { maxK++ }
	if maxK == 0 { return nil }

	bestPaths := [][]string{}
	bestTurns := -1

	for k := 1; k <= maxK; k++ {
		adj := buildAdj(colony, idx, bigCap)
		for j := 0; j < k; j++ { bfsAugment(adj, 2*si+1, 2*ei) }
		paths := extractPaths(adj, colony, idx, k)
		if len(paths) != k { break }

		turns := calcTurns(colony.numAnts, paths)
		if bestTurns < 0 || turns < bestTurns {
			bestTurns = turns
			bestPaths = paths
		} else {
			break // more paths only makes it worse from here
		}
	}
	return bestPaths
}

// ---- Utilities ----

func calcTurns(numAnts int, paths [][]string) int {
	if len(paths) == 0 { return 0 }
	s := sortedByLength(paths)
	a := distributeAnts(numAnts, s)
	max := 0
	for i, c := range a {
		if c > 0 {
			if t := c + len(s[i]) - 2; t > max { max = t }
		}
	}
	return max
}

func sortedByLength(paths [][]string) [][]string {
	s := make([][]string, len(paths))
	copy(s, paths)
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if len(s[i]) > len(s[j]) { s[i], s[j] = s[j], s[i] }
		}
	}
	return s
}

func distributeAnts(numAnts int, paths [][]string) []int {
	a := make([]int, len(paths))
	for i := 0; i < numAnts; i++ {
		best, bt := 0, -1
		for j, p := range paths {
			if t := a[j] + len(p) - 1; bt < 0 || t < bt { best, bt = j, t }
		}
		a[best]++
	}
	return a
}

func bfs(colony *Colony, blocked map[string]bool) []string {
	prev := map[string]string{colony.start: ""}
	q := []string{colony.start}
	for len(q) > 0 {
		cur := q[0]; q = q[1:]
		for _, next := range colony.links[cur] {
			if _, seen := prev[next]; seen { continue }
			if blocked[next] && next != colony.end { continue }
			prev[next] = cur
			if next == colony.end { return reconstructPath(prev, colony.end) }
			q = append(q, next)
		}
	}
	return nil
}

func reconstructPath(prev map[string]string, end string) []string {
	var p []string
	for n := end; n != ""; n = prev[n] { p = append([]string{n}, p...) }
	return p
}
