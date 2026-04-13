package main
 
import (
	"fmt"
	"strings"
)
 
// simulate runs the ant colony turn by turn and returns each turn's moves
// as a slice of strings like ["L1-3 L2-2", "L1-4 L2-5 L3-3", ...].
func simulate(colony *Colony, paths [][]string) []string {
	// Sort paths shortest-first so we fill the fastest lanes first.
	sorted := sortPaths(paths)
 
	// Decide how many ants travel each path (greedy by finish-turn).
	counts := distributeAnts(colony.numAnts, sorted)
 
	// ── Build ant list ────────────────────────────────────────
	// We number ants in round-robin release order across paths.
	// Example: path A gets 3 ants, path B gets 2 →
	//   ant 1 = A, ant 2 = B, ant 3 = A, ant 4 = B, ant 5 = A
	type ant struct {
		id   int
		path []string
		step int  // index into path; 0 = at start, len(path)-1 = at end
		done bool
	}
 
	// queues[pathIdx] holds the ants queued for that path (in release order).
	queues := make([][]*ant, len(sorted))
	antID := 1
	assigned := make([]int, len(sorted))
	total := 0
 
	for total < colony.numAnts {
		for pi := range sorted {
			if assigned[pi] < counts[pi] {
				queues[pi] = append(queues[pi], &ant{id: antID, path: sorted[pi]})
				antID++
				assigned[pi]++
				total++
				if total >= colony.numAnts {
					break
				}
			}
		}
	}
 
	// Flatten into a single slice with per-path start offsets for fast access.
	ants := make([]*ant, 0, colony.numAnts)
	pathStart := make([]int, len(sorted))
	for pi, q := range queues {
		pathStart[pi] = len(ants)
		ants = append(ants, q...)
	}
 
	// nextRelease[pi] = how many ants on path pi have been released so far.
	nextRelease := make([]int, len(sorted))
 
	var moves []string
	totalDone := 0
 
	// ── Main simulation loop ──────────────────────────────────
	for totalDone < colony.numAnts {
		// occupied tracks intermediate rooms currently holding an ant.
		// (start and end can hold any number of ants simultaneously.)
		occupied := map[string]bool{}
		// usedTunnel prevents two ants from crossing the same tunnel in one turn.
		usedTunnel := map[[2]string]bool{}
 
		// Pre-mark rooms occupied by released ants at their current positions.
		for _, a := range ants {
			if a.step > 0 && !a.done {
				room := a.path[a.step]
				if room != colony.start && room != colony.end {
					occupied[room] = true
				}
			}
		}
 
		var turnMoves []string
 
		// ── Step 1: move already-released ants (furthest first) ──
		for pi, path := range sorted {
			start := pathStart[pi]
			count := counts[pi]
 
			// Collect active ants on this path and sort by step descending
			// so the ant closest to the end moves first (avoids blocking).
			active := make([]*ant, 0, count)
			for i := start; i < start+count; i++ {
				a := ants[i]
				if a.step > 0 && !a.done {
					active = append(active, a)
				}
			}
			// Simple sort: largest step first.
			for i := 0; i < len(active)-1; i++ {
				for j := i + 1; j < len(active); j++ {
					if active[i].step < active[j].step {
						active[i], active[j] = active[j], active[i]
					}
				}
			}
 
			for _, a := range active {
				nextStep := a.step + 1
				if nextStep >= len(path) {
					continue
				}
				from := path[a.step]
				to := path[nextStep]
				tunnel := [2]string{from, to}
 
				roomFree := to == colony.end || !occupied[to]
				tunnelFree := !usedTunnel[tunnel]
 
				if roomFree && tunnelFree {
					// Mark the tunnel used in both directions.
					usedTunnel[tunnel] = true
					usedTunnel[[2]string{to, from}] = true
					// Free the room the ant just left, occupy the new one.
					if from != colony.start && from != colony.end {
						occupied[from] = false
					}
					if to != colony.end {
						occupied[to] = true
					}
					a.step = nextStep
					turnMoves = append(turnMoves, fmt.Sprintf("L%d-%s", a.id, to))
					if to == colony.end {
						a.done = true
						totalDone++
					}
				}
			}
		}
 
		// ── Step 2: release one new ant per path ─────────────────
		for pi, path := range sorted {
			if nextRelease[pi] >= counts[pi] || len(path) < 2 {
				continue
			}
			antIdx := pathStart[pi] + nextRelease[pi]
			a := ants[antIdx]
			firstRoom := path[1]
			tunnel := [2]string{colony.start, firstRoom}
 
			roomFree := firstRoom == colony.end || !occupied[firstRoom]
			tunnelFree := !usedTunnel[tunnel]
 
			if roomFree && tunnelFree {
				usedTunnel[tunnel] = true
				usedTunnel[[2]string{firstRoom, colony.start}] = true
				if firstRoom != colony.end {
					occupied[firstRoom] = true
				}
				a.step = 1
				nextRelease[pi]++
				turnMoves = append(turnMoves, fmt.Sprintf("L%d-%s", a.id, firstRoom))
				if firstRoom == colony.end {
					a.done = true
					totalDone++
				}
			}
		}
 
		if len(turnMoves) == 0 {
			break // no progress — should not happen on valid input
		}
		moves = append(moves, strings.Join(turnMoves, " "))
	}
 
	return moves
}
 
// sortPaths returns a copy of paths sorted shortest-first.
func sortPaths(paths [][]string) [][]string {
	sorted := make([][]string, len(paths))
	copy(sorted, paths)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if len(sorted[i]) > len(sorted[j]) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}
 