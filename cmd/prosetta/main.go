package main

import (
	"fmt"
	"os"

	scaffoldcmd "polyglot-rosetta/internal/commands/scaffold"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	subArgs := os.Args[2:]

	var err error
	switch command {
	case "scaffold":
		err = scaffoldcmd.Execute(subArgs)
	default:
		printUsage()
		err = fmt.Errorf("unknown command %q", command)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Polyglot Rosetta CLI Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  prosetta <command> [options]")
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  scaffold    Scaffold a category or concept workspace")
}
