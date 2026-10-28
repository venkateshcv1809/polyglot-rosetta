package version

import "testing"

func TestValueHasDevelopmentDefault(t *testing.T) {
	if Value == "" {
		t.Fatal("Value is empty")
	}
	if !IsDevelopment() {
		t.Fatalf("IsDevelopment() = false for %q", Value)
	}
	previous := Value
	t.Cleanup(func() { Value = previous })
	Value = "v0.1.0"
	if IsDevelopment() {
		t.Fatal("released version treated as development")
	}
}
