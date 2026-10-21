// Command prosetta provides the public CLI for catalog synchronization,
// toolchain probing, and solution evaluation.
package main

import (
	"os"

	"polyglot-rosetta/internal/cli"
	catalogcmd "polyglot-rosetta/internal/commands/catalog"
	probecmd "polyglot-rosetta/internal/commands/probe"
	runcmd "polyglot-rosetta/internal/commands/run"
)

func main() {
	cli.ExitOnError(execute(os.Args[1:]))
}

func execute(args []string) error {
	return cli.Execute(cli.Command{
		Name:    "prosetta",
		Summary: "Synchronize problems, probe toolchains, and evaluate solutions.",
		Subcommands: []cli.Command{
			catalogcmd.Command(),
			probecmd.Command(),
			runcmd.Command(),
		},
	}, args)
}
