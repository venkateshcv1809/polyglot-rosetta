package catalog

import (
	"errors"
	"strings"
	"testing"
)

func TestConceptIDIsDeterministicAndPathScoped(t *testing.T) {
	hello, err := ConceptID("fundamentals/hello-world")
	if err != nil || hello != "c-2b23ca0786e0eb237f64cf2b8759bcc3" {
		t.Fatalf("published hello-world ID = %q, %v", hello, err)
	}
	first, err := ConceptID("arrays/traversal")
	if err != nil {
		t.Fatal(err)
	}
	again, err := ConceptID("arrays/traversal")
	if err != nil || first != again {
		t.Fatalf("same path IDs = %q, %q, %v", first, again, err)
	}
	second, err := ConceptID("objects/traversal")
	if err != nil || first == second {
		t.Fatalf("different path IDs = %q, %q, %v", first, second, err)
	}
	if len(first) != 34 || !strings.HasPrefix(first, "c-") {
		t.Fatalf("ID = %q, want c- plus 32 hex digits", first)
	}
	for _, invalid := range []string{"", "../traversal", "arrays//traversal", "arrays\\traversal", "Arrays/traversal"} {
		if _, err := ConceptID(invalid); err == nil {
			t.Errorf("ConceptID(%q) accepted invalid path", invalid)
		}
	}
}

func TestResolveConceptIDSupportsFullAndUniquePrefix(t *testing.T) {
	first, _ := ConceptID("arrays/traversal")
	second, _ := ConceptID("objects/traversal")
	concepts := map[string]string{first: "arrays", second: "objects"}
	for _, input := range []string{first, first[:10], first[2:10]} {
		got, err := ResolveConceptID(concepts, input)
		if err != nil || got != "arrays" {
			t.Fatalf("ResolveConceptID(%q) = %q, %v", input, got, err)
		}
	}
	if _, err := ResolveConceptID(concepts, "deadbeef"); !errors.Is(err, ErrConceptNotFound) {
		t.Fatalf("missing ID error = %v", err)
	}
	if _, err := ResolveConceptID(concepts, "short"); err == nil {
		t.Fatal("invalid prefix accepted")
	}
	shared := first[:10]
	concepts[shared+"000000000000000000000000"] = "collision"
	if _, err := ResolveConceptID(concepts, shared); !errors.Is(err, ErrConceptAmbiguous) {
		t.Fatalf("ambiguous ID error = %v", err)
	}
}
