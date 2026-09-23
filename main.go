package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "run":
		handleRun(os.Args[2:])
	case "submit":
		handleSubmit(os.Args[2:])
	case "scaffold":
		handleScaffold(os.Args[2:])
	case "check":
		handleCheck(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Polyglot Rosetta (prosetta) - Multi-modal DSA & Design Pattern Runner")
	fmt.Println("\nUsage:")
	fmt.Println("  prosetta <command> [options]")
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  run       Execute public test cases for a concept")
	fmt.Println("  submit    Execute public and hidden test cases for final scoring")
	fmt.Println("  scaffold  Generate language template files for a concept")
	fmt.Println("  check     Verify repository structure and manifest validity")
}

func handleRun(args []string) {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	concept := runCmd.String("concept", "", "Target concept identifier (e.g., two-sum)")
	lang := runCmd.String("lang", "go", "Target programming language (go, python)")

	runCmd.Parse(args)

	if *concept == "" {
		fmt.Fprintln(os.Stderr, "Error: --concept flag is required")
		os.Exit(1)
	}

	fmt.Printf("Running public tests for concept '%s' in language '%s'...\n", *concept, *lang)
	// TODO: Wire up Workspace provisioning, code runner, and UI summary printer here
}

func handleSubmit(args []string) {
	submitCmd := flag.NewFlagSet("submit", flag.ExitOnError)
	concept := submitCmd.String("concept", "", "Target concept identifier")
	lang := submitCmd.String("lang", "go", "Target programming language")

	submitCmd.Parse(args)

	if *concept == "" {
		fmt.Fprintln(os.Stderr, "Error: --concept flag is required")
		os.Exit(1)
	}

	fmt.Printf("Evaluating all test vectors (including hidden) for concept '%s'...\n", *concept)
	// TODO: Wire up full suite aggregator and evaluation report printer here
}

func handleScaffold(args []string) {
	scaffoldCmd := flag.NewFlagSet("scaffold", flag.ExitOnError)
	concept := scaffoldCmd.String("concept", "", "Concept identifier to scaffold")
	scaffoldCmd.Parse(args)

	fmt.Printf("Scaffolding files for concept: %s\n", *concept)
}

func handleCheck(args []string) {
	fmt.Println("Checking repository structure and root manifest integrity...")
}
