package versioncmd

import (
	versionapi "polyglot-rosetta/internal/api/version"
	"polyglot-rosetta/internal/version"
)

// Version returns the running engine version for every transport.
func Version() versionapi.Result {
	result := versionapi.NewResult(versionapi.StatusOK)
	result.Version = version.Value
	return result
}
