package cli

// ParseFunc converts command-line arguments into command-specific options.
type ParseFunc[T any] func(commandPath string, args []string) (T, error)

// RunFunc performs a command using parsed options.
type RunFunc[T any] func(options T) error

// Handle connects a command's option parser to its execution function.
func Handle[T any](parse ParseFunc[T], run RunFunc[T]) func(string, []string) error {
	return func(commandPath string, args []string) error {
		options, err := parse(commandPath, args)
		if err != nil {
			return err
		}
		return run(options)
	}
}
