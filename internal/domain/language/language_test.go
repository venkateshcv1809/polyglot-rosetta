package language

import (
	"testing"
)

func TestParse(format *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Language
		wantErr bool
	}{
		{"Valid Go lowercase", "go", Go, false},
		{"Valid Go uppercase", "GO", Go, false},
		{"Valid Python mixed case", "PyThOn", Python, false},
		{"Valid Rust with spaces", "  rust  ", Rust, false},
		{"Valid TypeScript", "typescript", TypeScript, false},
		{"Valid Zig", "zig", Zig, false},
		{"Invalid language", "java", "", true},
		{"Empty string", "", "", true},
		{"Whitespace only", "   ", "", true},
	}

	for _, tt := range tests {
		format.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	if !IsValid("go") {
		t.Errorf("Expected 'go' to be valid")
	}
	if !IsValid("ZIG") {
		t.Errorf("Expected 'ZIG' to be valid")
	}
	if IsValid("ruby") {
		t.Errorf("Expected 'ruby' to be invalid")
	}
}

func TestAllSupported(t *testing.T) {
	supported := AllSupported()
	if len(supported) != 5 {
		t.Errorf("Expected 5 supported languages, got %d", len(supported))
	}
}
