package toolchain

import (
	"testing"

	"polyglot-rosetta/internal/domain/language"
)

func TestNewOSChecker(t *testing.T) {
	checker, err := NewOSChecker(language.Go)
	if err != nil {
		t.Errorf("Expected no error for Go checker, got %v", err)
	}
	if checker.Bin != "go" {
		t.Errorf("Expected binary 'go', got '%s'", checker.Bin)
	}

	_, err = NewOSChecker("ruby")
	if err == nil {
		t.Errorf("Expected error for unsupported language 'ruby', got nil")
	}
}

func TestOSCheckerMissingBinary(t *testing.T) {
	checker := &OSChecker{
		Lang: language.Go,
		Bin:  "nonexistent-binary-xyz-123",
		Args: []string{"--version"},
	}

	info, err := checker.Check()
	if err != nil {
		t.Errorf("Expected no system error when binary is missing, got %v", err)
	}
	if info.Available {
		t.Errorf("Expected toolchain to be marked unavailable for a non-existent binary")
	}
}
