// Package probecmd defines the command for inspecting locally available
// language toolchains.
package probecmd

import "polyglot-rosetta/internal/cli"

const (
	commandName    = "probe"
	commandSummary = "Inspect local language toolchains"
)

// Command returns the probe command definition for use by any CLI entrypoint.
func Command() cli.Command {
	return cli.Command{Name: commandName, Summary: commandSummary, Run: probeCommand}
}
