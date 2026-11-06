// Command site stages the shared GitHub Pages theme and can serve a local preview.
package main

import (
	"os"

	"polyglot-rosetta/internal/cli"
	sitecmd "polyglot-rosetta/internal/commands/site"
)

func main() {
	cli.ExitOnError(sitecmd.Execute(os.Args[1:]))
}
