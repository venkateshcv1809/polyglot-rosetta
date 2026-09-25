package toolchain

import (
	"errors"
	"testing"

	"polyglot-rosetta/internal/domain/language"
)

// MockChecker implements the Checker interface for testing contracts.
type MockChecker struct {
	Info  ToolchainInfo
	Error error
}

func (m *MockChecker) Check() (ToolchainInfo, error) {
	return m.Info, m.Error
}

func TestCheckerInterfaceContract(t *testing.T) {
	expectedInfo := ToolchainInfo{
		Language:  language.Go,
		Available: true,
		Version:   "go version go1.22.0 linux/amd64",
		Path:      "/usr/local/go/bin/go",
	}

	// Verify that MockChecker satisfies the Checker interface
	var checker Checker = &MockChecker{
		Info:  expectedInfo,
		Error: nil,
	}

	info, err := checker.Check()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !info.Available {
		t.Errorf("Expected toolchain to be available")
	}

	if info.Language != language.Go {
		t.Errorf("Expected language Go, got %s", info.Language)
	}

	if info.Version != expectedInfo.Version {
		t.Errorf("Expected version %s, got %s", expectedInfo.Version, info.Version)
	}
}

func TestCheckerErrorHandling(t *testing.T) {
	expectedErr := errors.New("binary not found")
	var checker Checker = &MockChecker{
		Info:  ToolchainInfo{Language: language.Python, Available: false},
		Error: expectedErr,
	}

	info, err := checker.Check()
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	if info.Available {
		t.Errorf("Expected toolchain availability to be false on error")
	}
}
