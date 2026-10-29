// Package version exposes build metadata shared by application entrypoints.
package version

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	prosetta "polyglot-rosetta"
)

var metadataJSON = []byte(prosetta.VersionMetadata())

type metadata struct {
	Version   string `json:"version"`
	MinEngine string `json:"min_engine"`
}

var release = loadMetadata()

// Value is the application version embedded at compile time from version.json.
var Value = release.Version

// MinimumEngine returns the catalog's default minimum compatible core version.
func MinimumEngine() string {
	return release.MinEngine
}

func loadMetadata() metadata {
	info, err := parseMetadata(metadataJSON)
	if err != nil {
		panic(err)
	}
	return info
}

func parseMetadata(data []byte) (metadata, error) {
	var info metadata
	if err := json.Unmarshal(data, &info); err != nil {
		return metadata{}, fmt.Errorf("decode embedded version.json: %w", err)
	}
	valid := regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$`)
	for _, field := range []struct{ name, value string }{
		{"version", info.Version}, {"min_engine", info.MinEngine},
	} {
		if field.name == "version" && field.value == "dev" {
			continue
		}
		if !valid.MatchString(field.value) {
			expected := "a release version"
			if field.name == "version" {
				expected += " or dev"
			}
			return metadata{}, fmt.Errorf("embedded version.json %s must be %s, got %q", field.name, expected, field.value)
		}
	}
	return info, nil
}

// IsDevelopment reports whether the application version is empty or dev.
func IsDevelopment() bool {
	current := strings.TrimPrefix(strings.TrimSpace(Value), "v")
	return current == "" || current == "dev"
}
