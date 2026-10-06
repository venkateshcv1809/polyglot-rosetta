// Command generate provides maintainer tooling for scaffolding catalog content
// and compiling the generated index.
package main

import (
	"os"

	"polyglot-rosetta/internal/cli"
	generatecmd "polyglot-rosetta/internal/commands/generate"
)

func main() {
	cli.ExitOnError(generatecmd.Execute(os.Args[1:]))
}
