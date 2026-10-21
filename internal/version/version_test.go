package version

import "testing"

func TestValueHasDevelopmentDefault(t *testing.T) {
	if Value == "" {
		t.Fatal("Value is empty")
	}
}
