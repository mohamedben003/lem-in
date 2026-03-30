package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run . <filename>")
		os.Exit(1)
	}

	colony, err := parseFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	paths := findPaths(colony)
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: invalid data format, no path found")
		os.Exit(1)
	}

	moves := simulate(colony, paths)

	fmt.Print(colony.raw)
	fmt.Println()
	for _, line := range moves {
		fmt.Println(line)
	}
}
