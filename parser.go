package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Room struct {
	name string
	x, y int
}

type Colony struct {
	numAnts int
	rooms   map[string]*Room
	links   map[string][]string
	start   string
	end     string
	raw     string
}

func parseFile(filename string) (*Colony, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("ERROR: invalid data format, cannot read file")
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	colony := &Colony{
		rooms: make(map[string]*Room),
		links: make(map[string][]string),
	}

	// Build raw output (exclude comments but keep structure)
	var rawLines []string

	if len(lines) == 0 {
		return nil, fmt.Errorf("ERROR: invalid data format")
	}

	// Parse number of ants
	numAnts, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil || numAnts <= 0 {
		return nil, fmt.Errorf("ERROR: invalid data format, invalid number of ants")
	}
	colony.numAnts = numAnts
	rawLines = append(rawLines, lines[0])

	nextIsStart := false
	nextIsEnd := false

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			continue
		}

		if line == "##start" {
			nextIsStart = true
			rawLines = append(rawLines, line)
			continue
		}
		if line == "##end" {
			nextIsEnd = true
			rawLines = append(rawLines, line)
			continue
		}
		// Skip comments but don't add to raw
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Try to parse as a link
		if strings.Contains(line, "-") && !strings.Contains(line, " ") {
			parts := strings.Split(line, "-")
			if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
				a, b := parts[0], parts[1]
				if _, ok := colony.rooms[a]; !ok {
					rawLines = append(rawLines, line)
					continue
				}
				if _, ok := colony.rooms[b]; !ok {
					rawLines = append(rawLines, line)
					continue
				}
				// Check duplicate links
				for _, existing := range colony.links[a] {
					if existing == b {
						return nil, fmt.Errorf("ERROR: invalid data format, duplicate tunnel %s-%s", a, b)
					}
				}
				colony.links[a] = append(colony.links[a], b)
				colony.links[b] = append(colony.links[b], a)
				rawLines = append(rawLines, line)
				continue
			}
		}

		// Try to parse as a room: name x y
		parts := strings.Fields(line)
		if len(parts) == 3 {
			name := parts[0]
			if strings.HasPrefix(name, "L") || strings.HasPrefix(name, "#") {
				return nil, fmt.Errorf("ERROR: invalid data format, invalid room name: %s", name)
			}
			x, err1 := strconv.Atoi(parts[1])
			y, err2 := strconv.Atoi(parts[2])
			if err1 != nil || err2 != nil {
				return nil, fmt.Errorf("ERROR: invalid data format, invalid room coordinates")
			}
			if _, exists := colony.rooms[name]; exists {
				return nil, fmt.Errorf("ERROR: invalid data format, duplicate room: %s", name)
			}
			room := &Room{name: name, x: x, y: y}
			colony.rooms[name] = room

			if nextIsStart {
				colony.start = name
				nextIsStart = false
			} else if nextIsEnd {
				colony.end = name
				nextIsEnd = false
			}
			rawLines = append(rawLines, line)
			continue
		}

		// Unknown line — ignore as per spec
	}

	if colony.start == "" {
		return nil, fmt.Errorf("ERROR: invalid data format, no start room found")
	}
	if colony.end == "" {
		return nil, fmt.Errorf("ERROR: invalid data format, no end room found")
	}

	colony.raw = strings.Join(rawLines, "\n")
	return colony, nil
}
