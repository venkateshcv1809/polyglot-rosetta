package probecmd

import (
	"flag"
	"fmt"
	"os/exec"
)

// Execute checks environment availability for language toolchains.
func Execute(args []string) error {
	fs := flag.NewFlagSet("probe", flag.ExitOnError)
	lang := fs.String("lang", "", "Target language to probe (go, python, rust, zig, ts)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	languages := []string{"go", "python", "rust", "zig", "node"}
	if *lang != "" {
		languages = []string{*lang}
	}

	fmt.Println("Probing local system environment for language toolchains...")
	for _, l := range languages {
		binName := getBinaryName(l)
		path, err := exec.LookPath(binName)
		if err != nil {
			fmt.Printf("  [ ❌ ] %-8s : Not found (looked for '%s')\n", l, binName)
		} else {
			fmt.Printf("  [ ✔ ] %-8s : Available at %s\n", l, path)
		}
	}
	return nil
}

func getBinaryName(lang string) string {
	switch lang {
	case "go":
		return "go"
	case "python":
		return "python3"
	case "rust":
		return "rustc"
	case "zig":
		return "zig"
	case "node", "ts":
		return "node"
	default:
		return lang
	}
}
