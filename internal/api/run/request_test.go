package run

import (
	"testing"

	"polyglot-rosetta/internal/catalog"
)

func TestRequestValidate(t *testing.T) {
	helloID, _ := catalog.ConceptID("fundamentals/hello-world")
	arrayID, _ := catalog.ConceptID("arrays/two-sum")
	tests := []struct {
		name    string
		request Request
		wantErr bool
	}{
		{
			name: "valid public request",
			request: Request{
				ID: helloID, Language: "go", Source: "package main", Mode: ModePublic,
			},
		},
		{
			name: "valid full request",
			request: Request{
				ID: arrayID, Language: "python", Source: "def solve(): pass", Mode: ModeAll,
			},
		},
		{name: "invalid ID", request: Request{ID: "../secret", Language: "go", Source: "x", Mode: ModePublic}, wantErr: true},
		{name: "path instead of ID", request: Request{ID: "arrays/traversal", Language: "go", Source: "x", Mode: ModePublic}, wantErr: true},
		{name: "short ID", request: Request{ID: helloID[:10], Language: "go", Source: "x", Mode: ModePublic}},
		{name: "invalid language", request: Request{ID: helloID, Language: "ruby", Source: "x", Mode: ModePublic}, wantErr: true},
		{name: "empty source", request: Request{ID: helloID, Language: "go", Source: " ", Mode: ModePublic}, wantErr: true},
		{name: "invalid mode", request: Request{ID: helloID, Language: "go", Source: "x", Mode: "hidden"}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
