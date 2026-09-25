package main

import (
	"fmt"
	"os"

	scaffoldcmd "polyglot-rosetta/internal/commands/scaffold"
)

func main() {
	if err := scaffoldcmd.Execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
