package probecmd

import probeapi "polyglot-rosetta/internal/api/probe"

func errorResult(err error) probeapi.Result {
	result := probeapi.NewResult(probeapi.StatusError)
	result.Error = err.Error()
	return result
}
