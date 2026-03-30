package main

import (
	"fmt"
	"strings"
)

type Ant struct {
	id       int
	path     []string
	step     int
	released bool
	done     bool
}

func simulate(colony *Colony, paths [][]string) []string {
	sorted := sortPaths(paths)
	assignment := distributeAnts(colony.numAnts, sorted)
 
	// Build ant list per path, but number ants in interleaved release order.
	// e.g. with paths A(3 ants) and B(2 ants): release order is A,B,A,B,A
	// so ant IDs go: A[0]=1, B[0]=2, A[1]=3, B[1]=4, A[2]=5
	pathCount := make([]int, len(sorted))
	for i, c := range assignment {
		pathCount[i] = c
	}
 
	// queues[pathIdx] = slice of ants waiting to enter that path (in release order)
	queues := make([][]*Ant, len(sorted))
	for i := range queues {
		queues[i] = make([]*Ant, 0, pathCount[i])
	}
 
	antID := 1
	released := make([]int, len(sorted)) // how many assigned so far per path
	total := 0
	for total < colony.numAnts {
		for pathIdx := range sorted {
			if released[pathIdx] < pathCount[pathIdx] {
				queues[pathIdx] = append(queues[pathIdx], &Ant{
					id:   antID,
					path: sorted[pathIdx],
					step: 0,
				})
				antID++
				released[pathIdx]++
				total++
				if total >= colony.numAnts {
					break
				}
			}
		}
	}
 
	// Flatten into ants slice with pathStart offsets
	ants := make([]*Ant, 0, colony.numAnts)
	pathStart := make([]int, len(sorted))
	for pathIdx, q := range queues {
		pathStart[pathIdx] = len(ants)
		ants = append(ants, q...)
	}
 
	nextRelease := make([]int, len(sorted))
	var moves []string
	totalDone := 0

	for totalDone < colony.numAnts {
		// Track occupied rooms AND used tunnels this turn
		occupied := map[string]bool{}
		usedTunnel := map[[2]string]bool{} // (from, to) tunnel used

		// Pre-mark rooms occupied by released ants at current positions
		for _, ant := range ants {
			if ant.released && !ant.done {
				room := ant.path[ant.step]
				if room != colony.start && room != colony.end {
					occupied[room] = true
				}
			}
		}

		var turnMoves []string

		// Move released ants — furthest first per path
		for pathIdx := range sorted {
			path := sorted[pathIdx]
			start := pathStart[pathIdx]
			count := pathCount[pathIdx]

			// Collect active ants on this path, sorted by step desc
			var active []*Ant
			for i := start; i < start+count; i++ {
				a := ants[i]
				if a.released && !a.done {
					active = append(active, a)
				}
			}
			for i := 0; i < len(active)-1; i++ {
				for j := i + 1; j < len(active); j++ {
					if active[i].step < active[j].step {
						active[i], active[j] = active[j], active[i]
					}
				}
			}

			for _, ant := range active {
				nextStep := ant.step + 1
				if nextStep >= len(path) {
					continue
				}
				curRoom := path[ant.step]
				nextRoom := path[nextStep]
				tunnel := [2]string{curRoom, nextRoom}

				roomFree := nextRoom == colony.end || !occupied[nextRoom]
				tunnelFree := !usedTunnel[tunnel]

				if roomFree && tunnelFree {
					usedTunnel[tunnel] = true
					usedTunnel[[2]string{nextRoom, curRoom}] = true
					if curRoom != colony.start && curRoom != colony.end {
						occupied[curRoom] = false
					}
					if nextRoom != colony.end {
						occupied[nextRoom] = true
					}
					ant.step = nextStep
					turnMoves = append(turnMoves, fmt.Sprintf("L%d-%s", ant.id, nextRoom))
					if nextRoom == colony.end {
						ant.done = true
						totalDone++
					}
				}
			}
		}

		// Release one new ant per path per turn
		for pathIdx := range sorted {
			if nextRelease[pathIdx] >= pathCount[pathIdx] {
				continue
			}
			path := sorted[pathIdx]
			if len(path) < 2 {
				continue
			}
			antIdx := pathStart[pathIdx] + nextRelease[pathIdx]
			ant := ants[antIdx]
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
				ant.step = 1
				ant.released = true
				nextRelease[pathIdx]++
				turnMoves = append(turnMoves, fmt.Sprintf("L%d-%s", ant.id, firstRoom))
				if firstRoom == colony.end {
					ant.done = true
					totalDone++
				}
			}
		}

		if len(turnMoves) > 0 {
			moves = append(moves, strings.Join(turnMoves, " "))
		} else {
			break // safety: no progress possible
		}
	}

	return moves
}

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
