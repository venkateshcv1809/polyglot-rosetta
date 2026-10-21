package execution

import (
	"fmt"

	"polyglot-rosetta/internal/domain/language"
	"polyglot-rosetta/internal/domain/testcase"
)

type preparedProgram struct {
	build []Command
	run   Command
}

func prepareProgram(workspace *Workspace, lang language.Language, source string, testCases []testcase.TestCase, marker string) (preparedProgram, error) {
	var (
		files   map[string]string
		program preparedProgram
		err     error
	)

	switch lang {
	case language.Go:
		files, program, err = prepareGo(testCases, marker)
	case language.Python:
		files, program, err = preparePython(testCases, marker)
	case language.Rust:
		files, program, err = prepareRust(testCases, marker)
	case language.TypeScript:
		files, program, err = prepareTypeScript(testCases, marker)
	case language.Zig:
		files, program, err = prepareZig(testCases, marker)
	default:
		return preparedProgram{}, fmt.Errorf("unsupported language %q", lang)
	}
	if err != nil {
		return preparedProgram{}, err
	}
	files[sourceFilename(lang)] = source

	for name, content := range files {
		if _, err := workspace.WriteFile(name, content); err != nil {
			return preparedProgram{}, err
		}
	}
	for i := range program.build {
		program.build[i].Dir = workspace.Dir
	}
	program.run.Dir = workspace.Dir
	return program, nil
}
