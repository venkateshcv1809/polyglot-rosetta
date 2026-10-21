package toolchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"polyglot-rosetta/internal/config"
	"polyglot-rosetta/internal/domain/language"
)

const (
	cacheSchema   = 2
	maxCacheBytes = 64 << 10
)

// ErrUnavailable means the cached probe found missing or unusable tools.
var ErrUnavailable = errors.New("language toolchain unavailable")

type cacheEntry struct {
	Required    []string `json:"required"`
	Available   bool     `json:"available"`
	Unavailable []string `json:"unavailable,omitempty"`
}

type cacheSnapshot struct {
	SchemaVersion int                   `json:"schema_version"`
	Languages     map[string]cacheEntry `json:"languages"`
}

// DefaultCachePath returns the persistent native toolchain cache location.
func DefaultCachePath() (string, error) {
	paths, err := config.NativeCachePaths()
	if err != nil {
		return "", err
	}
	return paths.Toolchain, nil
}

// ClearCache removes the saved snapshot. Clearing an absent cache succeeds so
// the next availability check can rebuild it.
func ClearCache(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("toolchain cache path is not a file: %s", path)
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// RefreshCache probes all supported languages and atomically replaces the
// availability snapshot. Unavailable results are cached too.
func RefreshCache(path string, lookPath func(string) (string, error), inspect func(string, string) (string, error)) ([]Result, error) {
	results := Probe(language.AllSupported(), lookPath, inspect)
	snapshot := cacheSnapshot{SchemaVersion: cacheSchema, Languages: make(map[string]cacheEntry, len(results))}
	for _, result := range results {
		snapshot.Languages[string(result.Language)] = entryFromResult(result)
	}
	if err := writeCache(path, snapshot); err != nil {
		return results, fmt.Errorf("save toolchain cache: %w", err)
	}
	return results, nil
}

// RequireAvailable uses a complete cached snapshot, probing only when that
// snapshot is missing or invalid. A cached negative result requires a manual
// probe refresh after support is installed.
func RequireAvailable(path string, lang language.Language, lookPath func(string) (string, error), inspect func(string, string) (string, error)) error {
	if !language.IsValid(string(lang)) {
		return fmt.Errorf("unsupported language %q", lang)
	}
	snapshot, err := readCache(path)
	if err == nil {
		return availabilityError(lang, snapshot.Languages[string(lang)])
	}
	results, err := RefreshCache(path, lookPath, inspect)
	if err != nil {
		return err
	}
	for _, result := range results {
		if result.Language == lang {
			return availabilityError(lang, entryFromResult(result))
		}
	}
	return fmt.Errorf("probe omitted supported language %q", lang)
}

func entryFromResult(result Result) cacheEntry {
	entry := cacheEntry{Required: RequiredBinaries(result.Language), Available: true}
	for _, tool := range result.Tools {
		if tool.Error != nil {
			entry.Available = false
			entry.Unavailable = append(entry.Unavailable, tool.Binary)
		}
	}
	return entry
}

func availabilityError(lang language.Language, entry cacheEntry) error {
	if entry.Available {
		return nil
	}
	return fmt.Errorf("%w for %s (%s); run prosetta probe --lang=%s after installing support", ErrUnavailable, lang, strings.Join(entry.Unavailable, ", "), lang)
}

func readCache(path string) (cacheSnapshot, error) {
	file, err := os.Open(path)
	if err != nil {
		return cacheSnapshot{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxCacheBytes+1))
	if err != nil {
		return cacheSnapshot{}, err
	}
	if len(data) > maxCacheBytes {
		return cacheSnapshot{}, fmt.Errorf("toolchain cache exceeds size limit")
	}
	var snapshot cacheSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return cacheSnapshot{}, fmt.Errorf("decode toolchain cache: %w", err)
	}
	if snapshot.SchemaVersion != cacheSchema || snapshot.Languages == nil {
		return cacheSnapshot{}, fmt.Errorf("unsupported toolchain cache schema")
	}
	if len(snapshot.Languages) != len(language.AllSupported()) {
		return cacheSnapshot{}, fmt.Errorf("incomplete toolchain cache")
	}
	for _, lang := range language.AllSupported() {
		if _, found := snapshot.Languages[string(lang)]; !found {
			return cacheSnapshot{}, fmt.Errorf("incomplete toolchain cache: missing %s", lang)
		}
	}
	for name, entry := range snapshot.Languages {
		if !language.IsValid(name) || entry.Available && len(entry.Unavailable) != 0 || !entry.Available && len(entry.Unavailable) == 0 {
			return cacheSnapshot{}, fmt.Errorf("invalid toolchain cache entry for %q", name)
		}
		if !slices.Equal(entry.Required, RequiredBinaries(language.Language(name))) {
			return cacheSnapshot{}, fmt.Errorf("stale toolchain requirements for %q", name)
		}
	}
	return snapshot, nil
}

func writeCache(path string, snapshot cacheSnapshot) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".toolchains-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(append(data, '\n')); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
