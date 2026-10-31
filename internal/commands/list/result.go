package listcmd

import (
	"strings"

	"polyglot-rosetta/internal/catalog"
)

type listedConcept struct {
	id      string
	concept catalog.Concept
}

func filterConcepts(rows []listedConcept, search string) []listedConcept {
	needle := strings.ToLower(strings.TrimSpace(search))
	filtered := make([]listedConcept, 0, len(rows))
	for _, item := range rows {
		if needle == "" || strings.Contains(strings.ToLower(item.concept.Path), needle) || strings.Contains(strings.ToLower(item.concept.Title), needle) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func orderedConcepts(metadata catalog.Catalog) []listedConcept {
	rows := make([]listedConcept, 0, len(metadata.ConceptsByID))
	var visit func(catalog.Category)
	visit = func(category catalog.Category) {
		// Merge sibling categories and concepts, which the catalog stores separately.
		categoryIndex, conceptIndex := 0, 0
		for categoryIndex < len(category.Categories) || conceptIndex < len(category.Concepts) {
			useCategory := conceptIndex == len(category.Concepts)
			if !useCategory && categoryIndex < len(category.Categories) {
				nested := category.Categories[categoryIndex]
				ref := category.Concepts[conceptIndex]
				concept := metadata.ConceptsByID[ref.ID]
				useCategory = nested.Order < ref.Order || nested.Order == ref.Order && nested.ID < concept.Slug
			}
			if useCategory {
				visit(category.Categories[categoryIndex])
				categoryIndex++
				continue
			}
			ref := category.Concepts[conceptIndex]
			rows = append(rows, listedConcept{id: ref.ID, concept: metadata.ConceptsByID[ref.ID]})
			conceptIndex++
		}
	}
	for _, category := range metadata.Categories {
		visit(category)
	}
	return rows
}
