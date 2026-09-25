package main

import (
	"fmt"
	"os"

	runcmd "polyglot-rosetta/internal/commands/run"
)

func main() {
	if err := runcmd.Execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
