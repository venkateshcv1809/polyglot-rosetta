package schema

import "time"

// IndexNode represents a node in the compiled hierarchy tree.
type IndexNode struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Type        NodeType     `json:"type"`
	Order       int          `json:"order"`
	Description string       `json:"description,omitempty"`
	Difficulty  Difficulty   `json:"difficulty,omitempty"`
	Path        string       `json:"path"`
	Languages   []Language   `json:"languages,omitempty"`
	Children    []*IndexNode `json:"children,omitempty"`
}

// IndexSchema represents the final root index.json output.
type IndexSchema struct {
	GeneratedAt time.Time  `json:"generated_at"`
	Root        *IndexNode `json:"root"`
}
