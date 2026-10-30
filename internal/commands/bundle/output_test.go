package bundlecmd

import (
	"errors"
	"testing"
)

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestPrintCreatedBundlePropagatesOutputFailure(t *testing.T) {
	want := errors.New("write failed")
	if err := printCreatedBundle(failingWriter{err: want}, "dist/problems"); !errors.Is(err, want) {
		t.Fatalf("error = %v, want write failure", err)
	}
}
