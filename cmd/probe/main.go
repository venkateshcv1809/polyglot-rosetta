// Command probe inspects the language toolchains available on the local system.
package main

import (
	"os"

	"polyglot-rosetta/internal/cli"
	probecmd "polyglot-rosetta/internal/commands/probe"
)

func main() {
	cli.ExitOnError(probecmd.Execute(os.Args[1:]))
}
