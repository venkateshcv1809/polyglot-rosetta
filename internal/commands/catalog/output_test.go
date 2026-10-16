package catalogcmd

import (
	"errors"
	"testing"

	catalogapi "polyglot-rosetta/internal/api/catalog"
)

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestWriteResultPropagatesOutputFailure(t *testing.T) {
	want := errors.New("write failed")
	for _, format := range []string{"text", "json"} {
		if err := writeResult(failingWriter{err: want}, format, catalogapi.Result{Status: catalogapi.StatusOK, Root: "/cache/problems"}); !errors.Is(err, want) {
			t.Fatalf("format %s: error = %v, want write failure", format, err)
		}
	}
}
