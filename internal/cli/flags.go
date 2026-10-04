package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// FlagParser binds command options, parses arguments, and renders their usage.
type FlagParser struct {
	commandPath string
	forms       []string
	options     []Option
	flags       *flag.FlagSet
	output      io.Writer
}

// NewFlagParser creates a parser for one leaf command.
func NewFlagParser(commandPath string, forms ...string) *FlagParser {
	parser := &FlagParser{
		commandPath: commandPath,
		forms:       forms,
		flags:       flag.NewFlagSet(commandPath, flag.ContinueOnError),
		output:      os.Stdout,
	}
	parser.flags.Usage = parser.PrintUsage
	return parser
}

// StringVar registers a string option and includes it in command help.
func (p *FlagParser) StringVar(target *string, defaultValue string, option Option) {
	if option.Default == "" && defaultValue != "" {
		option.Default = defaultValue
	}
	p.options = append(p.options, option)
	p.flags.StringVar(target, option.Name, defaultValue, option.Description)
}

// BoolVar registers a boolean option and includes it in command help.
func (p *FlagParser) BoolVar(target *bool, defaultValue bool, option Option) {
	if option.Default == "" && defaultValue {
		option.Default = "true"
	}
	p.options = append(p.options, option)
	p.flags.BoolVar(target, option.Name, defaultValue, option.Description)
}

// Parse parses options and rejects positional arguments.
func (p *FlagParser) Parse(args []string) error {
	if len(args) > 0 && IsHelp(args[0]) {
		p.PrintUsage()
		return flag.ErrHelp
	}
	if err := p.flags.Parse(args); err != nil {
		return err
	}
	if p.flags.NArg() != 0 {
		p.PrintUsage()
		return fmt.Errorf("unexpected argument %q", p.flags.Arg(0))
	}
	return nil
}

// PrintUsage writes this parser's formatted command help.
func (p *FlagParser) PrintUsage() {
	PrintUsage(p.output, p.commandPath, p.forms, p.options)
}

// SetOutput redirects parser diagnostics and usage. It is primarily useful in tests.
func (p *FlagParser) SetOutput(output io.Writer) {
	p.output = output
	p.flags.SetOutput(output)
}
