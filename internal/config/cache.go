package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	cacheSchemaVersion = 1
	maxConfigBytes     = 64 << 10
	configEnvironment  = "PROSETTA_CONFIG"
)

// CachePaths are the resolved native cache locations shared by all commands.
type CachePaths struct {
	Root      string
	Catalog   string
	Toolchain string
}

type cacheFile struct {
	SchemaVersion int `json:"schema_version"`
	Cache         struct {
		Root      string `json:"root"`
		Catalog   string `json:"catalog"`
		Toolchain string `json:"toolchain"`
	} `json:"cache"`
}

// ConfigPath returns the single native configuration file location.
func ConfigPath() (string, error) {
	if value := strings.TrimSpace(os.Getenv(configEnvironment)); value != "" {
		return filepath.Abs(value)
	}
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	return filepath.Join(directory, AppDirName, ConfigFileName), nil
}

// NativeCachePaths reads the shared configuration, using defaults until a file exists.
func NativeCachePaths() (CachePaths, error) {
	if err := rejectLegacyCacheSettings(); err != nil {
		return CachePaths{}, err
	}
	location, err := ConfigPath()
	if err != nil {
		return CachePaths{}, err
	}
	settings, err := readCacheFile(location)
	if errors.Is(err, os.ErrNotExist) {
		if err := EnsureCacheConfig(); err != nil {
			return CachePaths{}, err
		}
		settings, err = readCacheFile(location)
	}
	if err != nil {
		return CachePaths{}, err
	}
	return resolveCachePaths(settings)
}

// EnsureCacheConfig creates the default JSON once, without replacing existing settings.
func EnsureCacheConfig() error {
	if err := rejectLegacyCacheSettings(); err != nil {
		return err
	}
	location, err := ConfigPath()
	if err != nil {
		return err
	}
	if _, err := readCacheFile(location); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	settings, err := defaultCacheFile()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(location), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	file, err := os.CreateTemp(filepath.Dir(location), ConfigTempPattern)
	if err != nil {
		return fmt.Errorf("create cache config: %w", err)
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(append(data, '\n')); err != nil {
		file.Close()
		return fmt.Errorf("write cache config: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync cache config: %w", err)
	}
	if err = file.Close(); err != nil {
		return fmt.Errorf("close cache config: %w", err)
	}
	if err := os.Link(file.Name(), location); errors.Is(err, os.ErrExist) {
		_, err = readCacheFile(location)
		return err
	} else if err != nil {
		return fmt.Errorf("publish cache config: %w", err)
	}
	return nil
}

func rejectLegacyCacheSettings() error {
	for _, name := range []string{"PROSETTA_CATALOG_CACHE", "PROSETTA_TOOLCHAIN_CACHE"} {
		if os.Getenv(name) != "" {
			return fmt.Errorf("%s is no longer supported; configure cache paths in config.json (use PROSETTA_CONFIG to locate it)", name)
		}
	}
	return nil
}

func defaultCacheFile() (cacheFile, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return cacheFile{}, fmt.Errorf("resolve user cache directory: %w", err)
	}
	var settings cacheFile
	settings.SchemaVersion = cacheSchemaVersion
	settings.Cache.Root = filepath.Join(root, AppDirName)
	settings.Cache.Catalog = CatalogCacheDirName
	settings.Cache.Toolchain = ToolchainCacheFileName
	return settings, nil
}

func readCacheFile(location string) (cacheFile, error) {
	file, err := os.Open(location)
	if err != nil {
		return cacheFile{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxConfigBytes+1))
	if err != nil {
		return cacheFile{}, fmt.Errorf("read cache config %q: %w", location, err)
	}
	if len(data) > maxConfigBytes {
		return cacheFile{}, fmt.Errorf("cache config %q exceeds size limit", location)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var settings cacheFile
	if err := decoder.Decode(&settings); err != nil {
		return cacheFile{}, fmt.Errorf("decode cache config %q: %w", location, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return cacheFile{}, fmt.Errorf("cache config %q must contain one JSON object", location)
	}
	if _, err := resolveCachePaths(settings); err != nil {
		return cacheFile{}, fmt.Errorf("invalid cache config %q: %w", location, err)
	}
	return settings, nil
}

func resolveCachePaths(settings cacheFile) (CachePaths, error) {
	if settings.SchemaVersion != cacheSchemaVersion {
		return CachePaths{}, fmt.Errorf("unsupported cache schema version %d", settings.SchemaVersion)
	}
	root := settings.Cache.Root
	if !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return CachePaths{}, fmt.Errorf("cache root must be an absolute clean path")
	}
	for _, entry := range []struct{ name, value string }{
		{"catalog", settings.Cache.Catalog}, {"toolchain", settings.Cache.Toolchain},
	} {
		if !filepath.IsLocal(entry.value) || filepath.Clean(entry.value) != entry.value {
			return CachePaths{}, fmt.Errorf("cache %s path must be a clean relative path", entry.name)
		}
	}
	catalog := filepath.Join(root, settings.Cache.Catalog)
	toolchain := filepath.Join(root, settings.Cache.Toolchain)
	if pathWithin(catalog, toolchain) || pathWithin(toolchain, catalog) {
		return CachePaths{}, fmt.Errorf("catalog and toolchain cache paths must not overlap")
	}
	return CachePaths{Root: root, Catalog: catalog, Toolchain: toolchain}, nil
}

func pathWithin(parent, child string) bool {
	relative, err := filepath.Rel(parent, child)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
