package runcmd

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// Execute handles code execution against public or all test cases.
func Execute(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.Usage = PrintUsage

	concept := fs.String("concept", "", "Target concept identifier")
	lang := fs.String("lang", "go", "Target programming language")
	all := fs.Bool("all", false, "Execute against all test cases (public + hidden)")

	// Input source options
	path := fs.String("path", "", "Path to source file (CLI execution)")
	useStdin := fs.Bool("stdin", false, "Read source code from standard input (Neovim)")
	code := fs.String("code", "", "Inline source code string (Browser/WASM mode)")

	if err := fs.Parse(args); err != nil {
		PrintUsage()
		return err
	}

	if *concept == "" {
		PrintUsage()
		return fmt.Errorf("missing required flag: --concept")
	}

	// Validate that at least one code input mechanism is provided
	hasPath := *path != ""
	hasStdin := *useStdin
	hasCode := *code != ""

	if !hasPath && !hasStdin && !hasCode {
		PrintUsage()
		return fmt.Errorf("missing input source: you must provide one of --path, --stdin, or --code")
	}

	var sourceContent string
	mode := "public test cases"
	if *all {
		mode = "all test cases (including hidden)"
	}

	fmt.Printf("Executing concept %q for language %q against %s...\n", *concept, *lang, mode)

	// Extract source code based on the active vector
	if hasPath {
		fmt.Printf("-> Using file path: %s\n", *path)
		// TODO: Read file contents into sourceContent
	} else if hasStdin {
		fmt.Printf("-> Reading source code from stdin (Neovim stream)\n")
		bytesData, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		sourceContent = string(bytesData)
	} else if hasCode {
		fmt.Printf("-> Using inline code string (WASM context)\n")
		sourceContent = *code
	}

	_ = sourceContent

	// TODO: Hook into internal workspace provisioning and harness execution engine
	return nil
}

// PrintUsage outputs the standard help text for the run command.
func PrintUsage() {
	fmt.Println("Run Commands:")
	fmt.Println("  run --concept=<id> --path=<file> [options]")
	fmt.Println("  run --concept=<id> --stdin [options]")
	fmt.Println("  run --concept=<id> --code=<string> [options]")
	fmt.Println("\nInput Sources (Requires at least one):")
	fmt.Println("  --path     Path to source file (used for repo CLI execution)")
	fmt.Println("  --stdin    Read source code from standard input (used for Neovim)")
	fmt.Println("  --code     Inline code string (used for browser WASM mode)")
	fmt.Println("\nOptions:")
	fmt.Println("  --concept  Target concept identifier (required)")
	fmt.Println("  --lang     Target programming language (default \"go\")")
	fmt.Println("  --all      Execute against all test cases, including hidden ones")
}
