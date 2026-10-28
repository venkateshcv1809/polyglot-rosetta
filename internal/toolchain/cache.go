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

	probeapi "polyglot-rosetta/internal/api/probe"
	"polyglot-rosetta/internal/config"
	"polyglot-rosetta/internal/domain/language"
)

const maxCacheBytes = 64 << 10

// ErrUnavailable means the cached probe found missing or unusable tools.
var ErrUnavailable = errors.New("language toolchain unavailable")

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
	snapshot := probeapi.CacheSnapshot{
		SchemaVersion: probeapi.CacheSchemaVersion,
		Languages:     make(map[string]probeapi.CacheEntry, len(results)),
	}
	for _, result := range results {
		snapshot.Languages[string(result.Language)] = entryFromResult(result)
	}
	if err := writeCache(path, snapshot); err != nil {
		return results, fmt.Errorf("save toolchain cache: %w", err)
	}
	return results, nil
}

// EnsureAvailable uses a complete cached snapshot, probing only when that
// snapshot is missing or invalid. A cached negative result requires a manual
// probe refresh after support is installed.
func EnsureAvailable(path string, lang language.Language, lookPath func(string) (string, error), inspect func(string, string) (string, error)) error {
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

func entryFromResult(result Result) probeapi.CacheEntry {
	entry := probeapi.CacheEntry{Required: RequiredBinaries(result.Language), Available: true}
	for _, tool := range result.Tools {
		if tool.Error != nil {
			entry.Available = false
			entry.Unavailable = append(entry.Unavailable, tool.Binary)
		}
	}
	return entry
}

func availabilityError(lang language.Language, entry probeapi.CacheEntry) error {
	if entry.Available {
		return nil
	}
	return fmt.Errorf("%w for %s (%s); run prosetta probe --lang=%s after installing support", ErrUnavailable, lang, strings.Join(entry.Unavailable, ", "), lang)
}

func readCache(path string) (probeapi.CacheSnapshot, error) {
	file, err := os.Open(path)
	if err != nil {
		return probeapi.CacheSnapshot{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxCacheBytes+1))
	if err != nil {
		return probeapi.CacheSnapshot{}, err
	}
	if len(data) > maxCacheBytes {
		return probeapi.CacheSnapshot{}, fmt.Errorf("toolchain cache exceeds size limit")
	}
	var snapshot probeapi.CacheSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return probeapi.CacheSnapshot{}, fmt.Errorf("decode toolchain cache: %w", err)
	}
	if snapshot.SchemaVersion != probeapi.CacheSchemaVersion || snapshot.Languages == nil {
		return probeapi.CacheSnapshot{}, fmt.Errorf("unsupported toolchain cache schema")
	}
	if len(snapshot.Languages) != len(language.AllSupported()) {
		return probeapi.CacheSnapshot{}, fmt.Errorf("incomplete toolchain cache")
	}
	for _, lang := range language.AllSupported() {
		if _, found := snapshot.Languages[string(lang)]; !found {
			return probeapi.CacheSnapshot{}, fmt.Errorf("incomplete toolchain cache: missing %s", lang)
		}
	}
	for name, entry := range snapshot.Languages {
		if !language.IsValid(name) || entry.Available && len(entry.Unavailable) != 0 || !entry.Available && len(entry.Unavailable) == 0 {
			return probeapi.CacheSnapshot{}, fmt.Errorf("invalid toolchain cache entry for %q", name)
		}
		if !slices.Equal(entry.Required, RequiredBinaries(language.Language(name))) {
			return probeapi.CacheSnapshot{}, fmt.Errorf("stale toolchain requirements for %q", name)
		}
	}
	return snapshot, nil
}

func writeCache(path string, snapshot probeapi.CacheSnapshot) error {
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
