package listcmd

import (
	"bytes"
	"errors"
	"testing"

	"polyglot-rosetta/internal/catalog"
)

func TestWriteListTemplates(t *testing.T) {
	for _, test := range []struct {
		name   string
		rows   []listedConcept
		search string
		want   string
	}{
		{
			name: "no concepts",
			want: "\nConcepts:\n  No concepts found.\n",
		},
		{
			name:   "no matches",
			search: "traversal",
			want:   "\nConcepts:\n  No concepts match \"traversal\".\n",
		},
		{
			name: "one concept",
			rows: []listedConcept{{id: "c-1234567890abcdef", concept: catalog.Concept{Path: "arrays/traversal", Title: "Traversal"}}},
			want: "\nConcepts:\n" +
				"  PATH              SHORT ID    FULL ID             TITLE\n" +
				"  arrays/traversal  c-12345678  c-1234567890abcdef  Traversal\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := writeList(&output, test.rows, test.search); err != nil {
				t.Fatalf("writeList: %v", err)
			}
			if got := output.String(); got != test.want {
				t.Fatalf("output = %q, want %q", got, test.want)
			}
		})
	}
}

func TestWriteListPropagatesWriterError(t *testing.T) {
	want := errors.New("write failed")
	err := writeList(failingWriter{err: want}, nil, "")
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want write failure", err)
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }
