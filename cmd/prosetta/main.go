// Command prosetta provides the public CLI for probing toolchains and running
// solutions against catalog test cases.
package main

import (
	"os"

	"polyglot-rosetta/internal/cli"
	probecmd "polyglot-rosetta/internal/commands/probe"
	runcmd "polyglot-rosetta/internal/commands/run"
)

func main() {
	cli.ExitOnError(execute(os.Args[1:]))
}

func execute(args []string) error {
	return cli.Execute(cli.Command{
		Name:    "prosetta",
		Summary: "Probe local toolchains and evaluate solutions.",
		Subcommands: []cli.Command{
			probecmd.Command(),
			runcmd.Command(),
		},
	}, args)
}
