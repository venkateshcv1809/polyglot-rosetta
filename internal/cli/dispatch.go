package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Command describes one node in a command tree. Branch commands declare
// Subcommands; leaf commands provide Run.
type Command struct {
	Name        string
	Summary     string
	Subcommands []Command
	Run         func(commandPath string, args []string) error
}

// IsHelp reports whether arg is a help request.
func IsHelp(arg string) bool {
	return arg == "help" || arg == "-h" || arg == "--help"
}

// ExitOnError reports a command failure and terminates the process. A nil error
// and flag.ErrHelp are successful outcomes.
func ExitOnError(err error) {
	if code := ExitCode(err, os.Stderr); code != 0 {
		os.Exit(code)
	}
}

// ExitCode reports a command failure to stderr and returns its process exit code.
func ExitCode(err error, stderr io.Writer) int {
	if err == nil || errors.Is(err, flag.ErrHelp) {
		return 0
	}
	fmt.Fprintf(stderr, "error: %v\n", err)
	return 1
}

// Execute walks a command tree and keeps the complete command path available
// to leaf handlers and nested help output.
func Execute(command Command, args []string) error {
	return ExecuteWithWriters(command, args, os.Stdout, os.Stderr)
}

// ExecuteWithWriters executes a command tree using explicit writers for command
// lists and routing errors. Leaf handlers remain responsible for their own I/O.
func ExecuteWithWriters(command Command, args []string, stdout, stderr io.Writer) error {
	if err := ValidateCommand(command); err != nil {
		return err
	}
	return execute(command, command.Name, args, stdout, stderr)
}

// ValidateCommand verifies that a command tree is structurally unambiguous.
func ValidateCommand(command Command) error {
	return validateCommand(command, command.Name)
}

func validateCommand(command Command, commandPath string) error {
	if strings.TrimSpace(command.Name) == "" {
		return fmt.Errorf("command name cannot be empty")
	}
	if strings.ContainsAny(command.Name, " \t\r\n") {
		return fmt.Errorf("command name %q cannot contain whitespace", command.Name)
	}
	if len(command.Subcommands) == 0 {
		if command.Run == nil {
			return fmt.Errorf("command %q has no handler", commandPath)
		}
		return nil
	}
	if command.Run != nil {
		return fmt.Errorf("command %q cannot define both subcommands and a handler", commandPath)
	}

	seen := make(map[string]struct{}, len(command.Subcommands))
	for _, child := range command.Subcommands {
		if _, exists := seen[child.Name]; exists {
			return fmt.Errorf("command %q has duplicate subcommand %q", commandPath, child.Name)
		}
		seen[child.Name] = struct{}{}
		if err := validateCommand(child, commandPath+" "+child.Name); err != nil {
			return err
		}
	}
	return nil
}

func execute(command Command, commandPath string, args []string, stdout, stderr io.Writer) error {
	if len(command.Subcommands) == 0 {
		return command.Run(commandPath, args)
	}

	if len(args) == 0 {
		PrintCommandList(stderr, commandPath, command.Summary, command.Subcommands)
		return fmt.Errorf("missing command")
	}
	if IsHelp(args[0]) {
		PrintCommandList(stdout, commandPath, command.Summary, command.Subcommands)
		return nil
	}

	for _, child := range command.Subcommands {
		if child.Name == args[0] {
			return execute(child, commandPath+" "+child.Name, args[1:], stdout, stderr)
		}
	}

	PrintCommandList(stderr, commandPath, command.Summary, command.Subcommands)
	return fmt.Errorf("unknown command %q", args[0])
}

// PrintCommandList writes standard multi-command usage.
func PrintCommandList(w io.Writer, commandPath, summary string, commands []Command) {
	fmt.Fprintf(w, "Usage:\n  %s <command> [options]\n\n", commandPath)
	if summary != "" {
		fmt.Fprintf(w, "%s\n\n", summary)
	}
	fmt.Fprintln(w, "Commands:")
	width := 0
	for _, command := range commands {
		if n := len(command.Name); n > width {
			width = n
		}
	}
	for _, command := range commands {
		fmt.Fprintf(w, "  %-*s  %s\n", width, command.Name, command.Summary)
	}
}
