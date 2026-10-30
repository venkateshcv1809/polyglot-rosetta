package bundle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"polyglot-rosetta/internal/catalog"
)

type decodedCategory struct {
	archive  []byte
	concepts map[string]catalog.Concept
}

// Store holds metadata and lazily activated content. Failed loads never replace
// the active snapshot. Unchanged category archives reuse their decoded content.
type Store struct {
	current   *catalog.Catalog
	index     []byte
	templates []byte
	decoded   map[string]decodedCategory
}

// ParseIndex validates catalog metadata without requiring exercise archives.
func ParseIndex(data []byte) (catalog.Catalog, error) {
	var index catalog.Catalog
	if err := json.Unmarshal(data, &index); err != nil {
		return index, fmt.Errorf("decode bundle index: %w", err)
	}
	if !catalog.ValidNodeID(index.ID) || strings.TrimSpace(index.Title) == "" || index.Categories == nil || index.ConceptsByID == nil {
		return index, fmt.Errorf("invalid catalog index")
	}
	seen := map[string]bool{}
	var visit func([]catalog.Category, string) error
	visit = func(categories []catalog.Category, parent string) error {
		siblings := map[string]bool{}
		for _, category := range categories {
			if !catalog.ValidNodeID(category.ID) || siblings[category.ID] || strings.TrimSpace(category.Title) == "" {
				return fmt.Errorf("invalid catalog category %q", category.ID)
			}
			siblings[category.ID] = true
			prefix := parent + category.ID + "/"
			for _, ref := range category.Concepts {
				concept, ok := index.ConceptsByID[ref.ID]
				id, err := catalog.ConceptID(concept.Path)
				if !ok || err != nil || id != ref.ID || seen[ref.ID] || !strings.HasPrefix(concept.Path, prefix) || strings.Contains(strings.TrimPrefix(concept.Path, prefix), "/") || strings.TrimSpace(concept.Title) == "" {
					return fmt.Errorf("invalid catalog concept %q", ref.ID)
				}
				if err := concept.Difficulty.Validate(); err != nil {
					return err
				}
				seen[ref.ID] = true
			}
			if err := visit(category.Categories, prefix); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(index.Categories, ""); err != nil {
		return index, err
	}
	if len(seen) != len(index.ConceptsByID) {
		return index, fmt.Errorf("catalog index contains unregistered concepts")
	}
	return index, nil
}

// LoadIndex activates metadata only. Cached decoded content remains reusable
// when the same index and archive bytes are subsequently supplied.
func (s *Store) LoadIndex(data []byte) error {
	index, err := ParseIndex(data)
	if err != nil {
		return err
	}
	if !bytes.Equal(s.index, data) {
		s.decoded = nil
		s.templates = nil
	}
	s.index = bytes.Clone(data)
	s.current = &index
	return nil
}

// Load activates the supplied categories together with the complete index.
func (s *Store) Load(indexData, templates []byte, archives map[string][]byte) error {
	index, err := ParseIndex(indexData)
	if err != nil {
		return err
	}
	if len(archives) == 0 {
		return fmt.Errorf("catalog requires at least one category archive")
	}
	reusable := bytes.Equal(indexData, s.index) && bytes.Equal(templates, s.templates)
	decoded := make(map[string]decodedCategory, len(archives))
	if reusable {
		for name, cached := range s.decoded {
			decoded[name] = cached
		}
	}
	for name, archive := range archives {
		cached, ok := s.decoded[name]
		if !reusable || !ok || !bytes.Equal(cached.archive, archive) {
			loaded, err := LoadSelected(indexData, templates, map[string][]byte{name: archive})
			if err != nil {
				return err
			}
			cached = decodedCategory{archive: bytes.Clone(archive), concepts: loaded.Catalog.ConceptsByID}
		}
		decoded[name] = cached
		for id, concept := range cached.concepts {
			index.ConceptsByID[id] = concept
		}
	}
	s.index, s.templates, s.decoded, s.current = bytes.Clone(indexData), bytes.Clone(templates), decoded, &index
	return nil
}

// Catalog returns the active snapshot, which callers must treat as read-only.
func (s *Store) Catalog() (*catalog.Catalog, error) {
	if s.current == nil {
		return nil, fmt.Errorf("catalog content is not loaded; call ProsettaCore.catalog with catalog bytes first")
	}
	return s.current, nil
}
