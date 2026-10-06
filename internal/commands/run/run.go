package runcmd

import (
	"fmt"
	"io"
	"os"
)

func run(opts options) error {
	mode := "public test cases"
	if opts.all {
		mode = "all test cases (including hidden)"
	}
	fmt.Printf("Executing concept %q for language %q against %s...\n", opts.concept, opts.language, mode)

	source, err := sourceCode(opts)
	if err != nil {
		return err
	}
	if source == "" {
		return fmt.Errorf("source code is empty")
	}

	// TODO: Hook into internal workspace provisioning and harness execution engine.
	return fmt.Errorf("execution engine is not implemented")
}

func sourceCode(opts options) (string, error) {
	switch {
	case opts.path != "":
		fmt.Printf("-> Using file path: %s\n", opts.path)
		data, err := os.ReadFile(opts.path)
		if err != nil {
			return "", fmt.Errorf("failed to read source file %q: %w", opts.path, err)
		}
		return string(data), nil
	case opts.stdin:
		fmt.Println("-> Reading source code from stdin (Neovim stream)")
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		return string(data), nil
	default:
		fmt.Println("-> Using inline code string (WASM context)")
		return opts.code, nil
	}
}
