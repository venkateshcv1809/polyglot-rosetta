package scaffoldcmd

import (
	"bytes"
	"testing"

	"polyglot-rosetta/internal/domain/catalog"
	"polyglot-rosetta/internal/scaffold"
)

func TestParseConceptOptions(t *testing.T) {
	opts, err := parseConceptOptions("generate concept", []string{
		"--parent=algorithms/sorting",
		"--id=quick-sort",
		"--title=Quick Sort",
		"--description=Partition-based sorting",
		"--difficulty=Medium",
	})
	if err != nil {
		t.Fatalf("parseConceptOptions: %v", err)
	}
	if opts.parent != "algorithms/sorting" || opts.id != "quick-sort" || opts.title != "Quick Sort" || opts.description != "Partition-based sorting" || opts.difficulty != string(catalog.DifficultyMedium) {
		t.Fatalf("options = %+v", opts)
	}
}

func TestPrintCreatedConceptShowsPathAndRunID(t *testing.T) {
	var output bytes.Buffer
	if err := printCreatedConcept(&output, scaffold.CreateConceptOptions{ParentPath: "fundamentals", ID: "hello-world"}); err != nil {
		t.Fatalf("printCreatedConcept: %v", err)
	}
	want := "\nOK:\n  Concept created\n  Path:      problems/fundamentals/hello-world\n  Short ID:  c-2b23ca07\n  Full ID:   c-2b23ca0786e0eb237f64cf2b8759bcc3\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestParseConceptOptionsRejectsPositionalArguments(t *testing.T) {
	if _, err := parseConceptOptions("generate concept", []string{"quick-sort"}); err == nil {
		t.Fatal("parseConceptOptions returned nil error")
	}
}
