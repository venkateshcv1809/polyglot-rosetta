package versioncmd

import (
	versionapi "polyglot-rosetta/internal/api/version"
)

func errorResult(err error) versionapi.Result {
	result := versionapi.NewResult(versionapi.StatusError)
	result.Error = err.Error()
	return result
}
