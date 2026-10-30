package catalog

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"

	"polyglot-rosetta/internal/config"
	"polyglot-rosetta/internal/domain/testcase"
)

// Catalog is the ordered concept graph shared by local and published content.
// Content fields are populated only when a complete catalog is requested.
type Catalog struct {
	ID           string             `json:"id"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	Categories   []Category         `json:"categories"`
	ConceptsByID map[string]Concept `json:"concepts_by_id"`
}

// Category is an ordered branch in a catalog.
type Category struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Order       int          `json:"order"`
	Categories  []Category   `json:"categories,omitempty"`
	Concepts    []ConceptRef `json:"concepts"`
}

// ConceptRef connects a category to a concept and its sibling order.
type ConceptRef struct {
	ID    string `json:"id"`
	Order int    `json:"order"`
}

// Concept holds metadata and, when loaded, the exercise content at one path.
type Concept struct {
	Slug        string              `json:"slug"`
	Path        string              `json:"path"`
	Title       string              `json:"title"`
	Difficulty  Difficulty          `json:"difficulty"`
	Description string              `json:"description"`
	Order       int                 `json:"order"`
	Problem     string              `json:"problem,omitempty"`
	Explanation string              `json:"explanation,omitempty"`
	TestCases   []testcase.TestCase `json:"test_cases,omitempty"`
	Solutions   map[string]string   `json:"solutions,omitempty"`
	Starters    map[string]string   `json:"starters,omitempty"`
}

// ResolveConcept finds a concept by full ID or unique ID prefix.
func (catalog Catalog) ResolveConcept(id string) (Concept, error) {
	return ResolveConceptID(catalog.ConceptsByID, id)
}

// Build scans a local authoring tree into the canonical metadata catalog.
func Build(problemsDir string) (Catalog, error) {
	info, err := os.Stat(problemsDir)
	if err != nil {
		return Catalog{}, fmt.Errorf("cannot access problems directory %q: %w", problemsDir, err)
	}
	if !info.IsDir() {
		return Catalog{}, fmt.Errorf("problems path %q is not a directory", problemsDir)
	}
	return LoadMetadata(os.DirFS(problemsDir))
}

// LoadMetadata scans registered info.json files without loading exercise bodies.
func LoadMetadata(files fs.FS) (Catalog, error) {
	data, err := readInfo(files, ".")
	if err != nil {
		return Catalog{}, fmt.Errorf("read root info.json: %w", err)
	}
	var root ParentInfo
	if err := json.Unmarshal(data, &root); err != nil {
		return Catalog{}, fmt.Errorf("decode root info.json: %w", err)
	}
	if strings.TrimSpace(root.ID) == "" || strings.TrimSpace(root.Title) == "" {
		return Catalog{}, fmt.Errorf("catalog root must define id and title")
	}

	catalog := Catalog{
		ID: root.ID, Title: root.Title, Description: root.Description,
		Categories: []Category{}, ConceptsByID: make(map[string]Concept),
	}
	for _, ref := range sortedRefs(root.Items) {
		category, err := scanCategory(files, ".", ref, catalog.ConceptsByID)
		if err != nil {
			return Catalog{}, err
		}
		catalog.Categories = append(catalog.Categories, category)
	}
	return catalog, nil
}

func scanCategory(files fs.FS, parent string, ref ItemRef, concepts map[string]Concept) (Category, error) {
	directory, err := childDir(files, parent, ref.ID)
	if err != nil {
		return Category{}, fmt.Errorf("invalid category %q: %w", ref.ID, err)
	}
	info, _, leaf, err := readNode(files, directory)
	if err != nil {
		return Category{}, fmt.Errorf("read category %q: %w", ref.ID, err)
	}
	if leaf {
		return Category{}, fmt.Errorf("expected category %q but found a concept", ref.ID)
	}
	if info.ID != ref.ID {
		return Category{}, fmt.Errorf("category metadata ID %q does not match reference %q", info.ID, ref.ID)
	}
	return scanCategoryInfo(files, directory, ref, info, concepts)
}

func scanCategoryInfo(files fs.FS, directory string, ref ItemRef, info ParentInfo, concepts map[string]Concept) (Category, error) {
	if strings.TrimSpace(info.Title) == "" {
		return Category{}, fmt.Errorf("category %q has no title", info.ID)
	}
	category := Category{
		ID: info.ID, Title: info.Title, Description: info.Description, Order: ref.Order,
		Categories: []Category{}, Concepts: []ConceptRef{},
	}
	for _, childRef := range sortedRefs(info.Items) {
		childPath, err := childDir(files, directory, childRef.ID)
		if err != nil {
			return Category{}, fmt.Errorf("invalid child %q under category %q: %w", childRef.ID, ref.ID, err)
		}
		parent, concept, leaf, err := readNode(files, childPath)
		if err != nil {
			return Category{}, fmt.Errorf("read child %q under category %q: %w", childRef.ID, ref.ID, err)
		}
		if !leaf {
			if parent.ID != childRef.ID {
				return Category{}, fmt.Errorf("child metadata ID %q does not match reference %q", parent.ID, childRef.ID)
			}
			nested, err := scanCategoryInfo(files, childPath, childRef, parent, concepts)
			if err != nil {
				return Category{}, err
			}
			category.Categories = append(category.Categories, nested)
			continue
		}
		if concept.ID != childRef.ID {
			return Category{}, fmt.Errorf("child metadata ID %q does not match reference %q", concept.ID, childRef.ID)
		}
		if strings.TrimSpace(concept.Title) == "" {
			return Category{}, fmt.Errorf("concept %q has no title", concept.ID)
		}
		if err := concept.Difficulty.Validate(); err != nil {
			return Category{}, fmt.Errorf("concept %q: %w", concept.ID, err)
		}
		id, err := ConceptID(childPath)
		if err != nil {
			return Category{}, err
		}
		if previous, exists := concepts[id]; exists {
			return Category{}, fmt.Errorf("concept ID collision %q at %q and %q", id, previous.Path, childPath)
		}
		concepts[id] = Concept{
			Slug: concept.ID, Path: childPath, Title: concept.Title,
			Difficulty: concept.Difficulty, Description: concept.Description, Order: childRef.Order,
		}
		category.Concepts = append(category.Concepts, ConceptRef{ID: id, Order: childRef.Order})
	}
	return category, nil
}

func readNode(files fs.FS, directory string) (ParentInfo, ConceptInfo, bool, error) {
	data, err := readInfo(files, directory)
	if err != nil {
		return ParentInfo{}, ConceptInfo{}, false, err
	}
	var marker struct {
		Difficulty Difficulty `json:"difficulty"`
	}
	if err := json.Unmarshal(data, &marker); err != nil {
		return ParentInfo{}, ConceptInfo{}, false, fmt.Errorf("decode %q: %w", path.Join(directory, config.CatalogInfoFileName), err)
	}
	if marker.Difficulty == "" {
		var parent ParentInfo
		if err := json.Unmarshal(data, &parent); err != nil {
			return ParentInfo{}, ConceptInfo{}, false, err
		}
		return parent, ConceptInfo{}, false, nil
	}
	var concept ConceptInfo
	if err := json.Unmarshal(data, &concept); err != nil {
		return ParentInfo{}, ConceptInfo{}, false, err
	}
	return ParentInfo{}, concept, true, nil
}

func readInfo(files fs.FS, directory string) ([]byte, error) {
	entries, err := fs.ReadDir(files, directory)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.Name() != config.CatalogInfoFileName {
			continue
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return nil, fmt.Errorf("metadata file %q must not be a symbolic link", path.Join(directory, config.CatalogInfoFileName))
		}
		if !entry.Type().IsRegular() {
			return nil, fmt.Errorf("metadata file %q is not regular", path.Join(directory, config.CatalogInfoFileName))
		}
		return fs.ReadFile(files, path.Join(directory, config.CatalogInfoFileName))
	}
	return nil, fmt.Errorf("metadata file %q is missing", path.Join(directory, config.CatalogInfoFileName))
}

func childDir(files fs.FS, parent, id string) (string, error) {
	if !ValidNodeID(id) {
		return "", fmt.Errorf("ID must be lowercase kebab-case")
	}
	entries, err := fs.ReadDir(files, parent)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.Name() != id {
			continue
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return "", fmt.Errorf("directory %q must not be a symbolic link", path.Join(parent, id))
		}
		if !entry.IsDir() {
			return "", fmt.Errorf("%q is not a directory", path.Join(parent, id))
		}
		return path.Join(parent, id), nil
	}
	return "", fmt.Errorf("directory %q is missing", path.Join(parent, id))
}

func sortedRefs(items []ItemRef) []ItemRef {
	result := append([]ItemRef(nil), items...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Order == result[j].Order {
			return result[i].ID < result[j].ID
		}
		return result[i].Order < result[j].Order
	})
	return result
}
