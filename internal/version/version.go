// Package version exposes build metadata shared by application entrypoints.
package version

import "strings"

// Value is replaced with the release tag through the Go linker.
var Value = "dev"

// IsDevelopment reports whether this binary is an untagged local build.
func IsDevelopment() bool {
	current := strings.TrimPrefix(strings.TrimSpace(Value), "v")
	return current == "" || current == "dev"
}
