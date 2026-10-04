package probecmd

import (
	"fmt"
	"os/exec"
)

var supportedLanguages = []string{"go", "python", "rust", "typescript", "zig"}

func probe(opts options) error {
	languages := supportedLanguages
	if opts.language != "" {
		languages = []string{opts.language}
	}

	fmt.Println("Probing local system environment for language toolchains...")
	for _, language := range languages {
		binary := binaryName(language)
		path, err := exec.LookPath(binary)
		if err != nil {
			fmt.Printf("  FAIL  %-12s  not found (looked for %q)\n", language, binary)
			continue
		}
		fmt.Printf("  OK    %-12s  %s\n", language, path)
	}
	return nil
}

func binaryName(language string) string {
	switch language {
	case "python":
		return "python3"
	case "typescript", "ts":
		return "tsc"
	case "rust":
		return "rustc"
	case "go", "zig":
		return language
	default:
		return language
	}
}
