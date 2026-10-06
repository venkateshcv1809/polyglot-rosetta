// Command run executes a solution against the test cases for a catalog concept.
package main

import (
	"os"

	"polyglot-rosetta/internal/cli"
	runcmd "polyglot-rosetta/internal/commands/run"
)

func main() {
	cli.ExitOnError(runcmd.Execute(os.Args[1:]))
}
