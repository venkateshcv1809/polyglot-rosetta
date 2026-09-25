package main

import (
	"fmt"
	"os"

	probecmd "polyglot-rosetta/internal/commands/probe"
)

func main() {
	if err := probecmd.Execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
