package main

import (
	"fmt"
	"os"

	probecmd "polyglot-rosetta/internal/commands/probe"
	runcmd "polyglot-rosetta/internal/commands/run"
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
	case "probe":
		err = probecmd.Execute(subArgs)
	case "run":
		err = runcmd.Execute(subArgs)
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
	fmt.Println("  run   Execute code against test cases")
	fmt.Println("  probe Inspect local system environment for language toolchains")
}
