package versioncmd

import (
	versionapi "polyglot-rosetta/internal/api/version"
	"polyglot-rosetta/internal/version"
)

func versionResult() versionapi.Result {
	result := versionapi.NewResult(versionapi.StatusOK)
	result.Version = version.Value
	return result
}

func errorResult(err error) versionapi.Result {
	result := versionapi.NewResult(versionapi.StatusError)
	result.Error = err.Error()
	return result
}
