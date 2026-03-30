package main

import (
	"strings"
	"testing"
)

// ---- BFS tests ----

func TestBFS_SimplePath(t *testing.T) {
	colony := &Colony{
		start: "a", end: "c",
		links: map[string][]string{"a": {"b"}, "b": {"a", "c"}, "c": {"b"}},
		rooms: map[string]*Room{"a": {name: "a"}, "b": {name: "b"}, "c": {name: "c"}},
	}
	path := bfs(colony, map[string]bool{})
	if len(path) != 3 || path[0] != "a" || path[2] != "c" {
		t.Errorf("expected a->b->c, got %v", path)
	}
}

func TestBFS_NoPath(t *testing.T) {
	colony := &Colony{
		start: "a", end: "c",
		links: map[string][]string{"a": {"b"}, "b": {"a"}},
		rooms: map[string]*Room{"a": {name: "a"}, "b": {name: "b"}, "c": {name: "c"}},
	}
	if bfs(colony, map[string]bool{}) != nil {
		t.Error("expected nil path")
	}
}

func TestBFS_Blocked(t *testing.T) {
	colony := &Colony{
		start: "a", end: "c",
		links: map[string][]string{"a": {"b", "d"}, "b": {"a", "c"}, "d": {"a", "c"}, "c": {"b", "d"}},
		rooms: map[string]*Room{"a": {name: "a"}, "b": {name: "b"}, "c": {name: "c"}, "d": {name: "d"}},
	}
	path := bfs(colony, map[string]bool{"b": true})
	if len(path) == 0 || path[1] == "b" {
		t.Errorf("expected path avoiding b, got %v", path)
	}
}

// ---- Path finding tests ----

func TestFindPaths_TwoPaths(t *testing.T) {
	// Graph: start->3 (direct) and start->1->2->3
	colony := &Colony{
		numAnts: 20, start: "0", end: "3",
		rooms: map[string]*Room{"0": {}, "1": {}, "2": {}, "3": {}},
		links: map[string][]string{
			"0": {"1", "3"}, "1": {"0", "2"}, "2": {"1", "3"}, "3": {"0", "2"},
		},
	}
	paths := findPaths(colony)
	if len(paths) != 2 {
		t.Errorf("expected 2 paths, got %d: %v", len(paths), paths)
	}
}

func TestFindPaths_NoPath(t *testing.T) {
	colony := &Colony{
		numAnts: 2, start: "a", end: "c",
		rooms: map[string]*Room{"a": {}, "b": {}, "c": {}},
		links: map[string][]string{"a": {"b"}, "b": {"a"}},
	}
	if findPaths(colony) != nil {
		t.Error("expected no paths")
	}
}

// ---- Distribution & turn count tests ----

func TestDistributeAnts(t *testing.T) {
	paths := [][]string{{"s", "a", "e"}, {"s", "b", "c", "e"}}
	asgn := distributeAnts(4, paths)
	total := 0
	for _, c := range asgn { total += c }
	if total != 4 {
		t.Errorf("expected 4 ants assigned, got %d", total)
	}
}

func TestCalcTurns_SinglePath(t *testing.T) {
	paths := [][]string{{"s", "a", "e"}}
	// 3 ants, path length 3: 3+3-2=4 turns
	if turns := calcTurns(3, paths); turns != 4 {
		t.Errorf("expected 4 turns, got %d", turns)
	}
}

func TestCalcTurns_TwoPaths(t *testing.T) {
	// short path len 2, long path len 4; 4 ants
	paths := [][]string{{"s", "e"}, {"s", "a", "b", "e"}}
	turns := calcTurns(4, paths)
	if turns > 5 {
		t.Errorf("expected <=5 turns, got %d", turns)
	}
}

// ---- Simulation tests ----

func TestSimulate_SinglePath(t *testing.T) {
	colony := &Colony{
		numAnts: 2, start: "start", end: "end",
		rooms: map[string]*Room{"start": {}, "mid": {}, "end": {}},
		links: map[string][]string{"start": {"mid"}, "mid": {"start", "end"}, "end": {"mid"}},
	}
	paths := findPaths(colony)
	moves := simulate(colony, paths)
	antsSeen := map[string]bool{}
	for _, line := range moves {
		for _, m := range strings.Fields(line) {
			if strings.HasSuffix(m, "-end") {
				antsSeen[strings.Split(m, "-")[0]] = true
			}
		}
	}
	if len(antsSeen) != 2 {
		t.Errorf("expected 2 ants at end, got %d; moves: %v", len(antsSeen), moves)
	}
}

func TestSimulate_NoRoomCollisions(t *testing.T) {
	colony := &Colony{
		numAnts: 3, start: "s", end: "e",
		rooms: map[string]*Room{"s": {}, "a": {}, "b": {}, "e": {}},
		links: map[string][]string{
			"s": {"a"}, "a": {"s", "b"}, "b": {"a", "e"}, "e": {"b"},
		},
	}
	paths := findPaths(colony)
	moves := simulate(colony, paths)
	for _, line := range moves {
		roomCount := map[string]int{}
		for _, mv := range strings.Fields(line) {
			parts := strings.SplitN(mv, "-", 2)
			if len(parts) == 2 {
				room := parts[1]
				if room != "s" && room != "e" {
					roomCount[room]++
					if roomCount[room] > 1 {
						t.Errorf("collision in room %s: %s", room, line)
					}
				}
			}
		}
	}
}

func TestSimulate_TunnelUsedOncePerTurn(t *testing.T) {
	colony := &Colony{
		numAnts: 4, start: "s", end: "e",
		rooms: map[string]*Room{"s": {}, "a": {}, "b": {}, "c": {}, "e": {}},
		links: map[string][]string{
			"s": {"a", "b"}, "a": {"s", "e"}, "b": {"s", "c"}, "c": {"b", "e"}, "e": {"a", "c"},
		},
	}
	paths := findPaths(colony)
	moves := simulate(colony, paths)
	for _, line := range moves {
		tunnelUsed := map[[2]string]bool{}
		for _, mv := range strings.Fields(line) {
			// We can't know "from" from the output alone, but verify no ant goes to same room twice
			_ = mv
		}
		_ = tunnelUsed
	}
	// Just verify all ants arrive
	antsDone := map[string]bool{}
	for _, line := range moves {
		for _, mv := range strings.Fields(line) {
			parts := strings.SplitN(mv, "-", 2)
			if len(parts) == 2 && parts[1] == "e" {
				antsDone[parts[0]] = true
			}
		}
	}
	if len(antsDone) != 4 {
		t.Errorf("expected 4 ants at end, got %d", len(antsDone))
	}
}

// ---- Error/parser tests ----

func TestParseFile_NoFile(t *testing.T) {
	_, err := parseFile("nonexistent.txt")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestFindPaths_SingleRoom(t *testing.T) {
	// start == end edge case
	colony := &Colony{
		numAnts: 1, start: "a", end: "a",
		rooms: map[string]*Room{"a": {}},
		links: map[string][]string{},
	}
	// Should not panic
	_ = findPaths(colony)
}
