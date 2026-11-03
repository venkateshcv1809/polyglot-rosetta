package execution

import "polyglot-rosetta/internal/domain/language"

// Each language declares one optional build command, one per-case run command,
// and any supporting files. A nil build means no separate compilation step.
// Commands are data here; build.go and evaluator.go own their execution.
type programDefinition struct {
	build *Command
	run   Command
	files map[string]string
}

var languagePrograms = map[language.Language]programDefinition{
	language.Go: {
		build: &Command{Name: "go", Args: []string{"build", "-buildvcs=false", "-o", "solution", "."}, Env: []string{"GOWORK=off"}},
		run:   Command{Name: "./solution"},
		files: map[string]string{"go.mod": "module prosetta-evaluation\n\ngo 1.22\n"},
	},
	language.Python: {
		run: Command{Name: "python3", Args: []string{"main.py"}},
	},
	language.Rust: {
		build: &Command{Name: "rustc", Args: []string{"--edition=2021", "-C", "opt-level=3", "main.rs", "-o", "solution"}},
		run:   Command{Name: "./solution"},
	},
	language.TypeScript: {
		build: &Command{Name: "tsc", Args: []string{"--target", "ES2022", "--module", "commonjs", "--outDir", "build", "solution.ts", "main.ts"}},
		run:   Command{Name: "node", Args: []string{"build/main.js"}},
	},
	language.Zig: {
		build: &Command{Name: "zig", Args: []string{"build-exe", "main.zig", "-O", "ReleaseFast", "-lc", "-femit-bin=solution"}},
		run:   Command{Name: "./solution"},
	},
}

type preparedProgram struct {
	build *Command
	run   Command
}
