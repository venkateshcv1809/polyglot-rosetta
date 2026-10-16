package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const (
	conceptIDPrefix       = "c-"
	conceptIDBytes        = 16
	MinConceptIDHexDigits = 8
)

var (
	conceptPathPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*(?:/[a-z0-9]+(?:-[a-z0-9]+)*)*$`)
	conceptIDPattern   = regexp.MustCompile(`^[0-9a-f]{8,32}$`)

	ErrConceptNotFound  = errors.New("concept not found")
	ErrConceptAmbiguous = errors.New("concept ID prefix is ambiguous")
)

// ConceptID derives a stable machine ID from a catalog-root-relative path.
// Changing the path changes the ID; content and absolute filesystem location do not.
func ConceptID(relativePath string) (string, error) {
	if !conceptPathPattern.MatchString(relativePath) {
		return "", fmt.Errorf("invalid concept path %q: expected lowercase kebab-case segments", relativePath)
	}
	sum := sha256.Sum256([]byte(relativePath))
	return conceptIDPrefix + hex.EncodeToString(sum[:conceptIDBytes]), nil
}

// NormalizeConceptID accepts a full ID or a unique-prefix candidate.
func NormalizeConceptID(value string) (string, error) {
	digits := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), conceptIDPrefix)
	if !conceptIDPattern.MatchString(digits) {
		return "", fmt.Errorf("concept ID must have %d to %d hexadecimal digits, optionally prefixed with %q", MinConceptIDHexDigits, conceptIDBytes*2, conceptIDPrefix)
	}
	return conceptIDPrefix + digits, nil
}

// ResolveConceptID performs direct full-ID lookup and supports unique short prefixes.
func ResolveConceptID[T any](concepts map[string]T, input string) (T, error) {
	var zero T
	id, err := NormalizeConceptID(input)
	if err != nil {
		return zero, err
	}
	if value, ok := concepts[id]; ok {
		return value, nil
	}
	var matches []string
	for candidate := range concepts {
		if strings.HasPrefix(candidate, id) {
			matches = append(matches, candidate)
		}
	}
	switch len(matches) {
	case 0:
		return zero, fmt.Errorf("%w: %q", ErrConceptNotFound, input)
	case 1:
		return concepts[matches[0]], nil
	default:
		sort.Strings(matches)
		return zero, fmt.Errorf("%w: %q matches %v", ErrConceptAmbiguous, input, matches)
	}
}
