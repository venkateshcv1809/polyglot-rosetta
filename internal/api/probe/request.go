package probe

// Request selects inspection or removal of the saved readiness snapshot.
type Request struct {
	Language string `json:"language,omitempty"`
	Clear    bool   `json:"clear,omitempty"`
}
