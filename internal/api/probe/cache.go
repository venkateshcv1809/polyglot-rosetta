package probe

import "polyglot-rosetta/internal/api"

const CacheSchemaVersion = api.SchemaVersion

// CacheEntry records whether one language can run immediately.
type CacheEntry struct {
	Required    []string `json:"required"`
	Available   bool     `json:"available"`
	Unavailable []string `json:"unavailable,omitempty"`
}

// CacheSnapshot is the JSON stored in the native cache file and browser localStorage.
type CacheSnapshot struct {
	SchemaVersion int                   `json:"schema_version"`
	Languages     map[string]CacheEntry `json:"languages"`
}

// Ready reports whether language can execute now according to the snapshot.
func (snapshot CacheSnapshot) Ready(language string) bool {
	if snapshot.SchemaVersion != CacheSchemaVersion || snapshot.Languages == nil {
		return false
	}
	entry, ok := snapshot.Languages[language]
	return ok && entry.Available
}
