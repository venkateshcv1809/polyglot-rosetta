// Package prosetta embeds version information maintained at the repository root.
package prosetta

import _ "embed"

//go:embed version.json
var versionJSON string

// VersionMetadata returns the version configuration embedded at compile time.
func VersionMetadata() string {
	return versionJSON
}
