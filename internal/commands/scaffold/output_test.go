package scaffoldcmd

import (
	"bytes"
	"testing"

	"polyglot-rosetta/internal/scaffold"
)

func TestPrintCreatedCategoryShowsPath(t *testing.T) {
	for _, test := range []struct {
		name   string
		parent string
		want   string
	}{
		{name: "root", parent: ".", want: "problems/fundamentals"},
		{name: "nested", parent: "algorithms", want: "problems/algorithms/fundamentals"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := printCreatedCategory(&output, scaffold.CreateCategoryOptions{ParentPath: test.parent, ID: "fundamentals"}); err != nil {
				t.Fatal(err)
			}
			if got, want := output.String(), "\nOK:\n  Category created\n  Path:      "+test.want+"\n"; got != want {
				t.Fatalf("output = %q, want %q", got, want)
			}
		})
	}
}
