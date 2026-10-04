package indexcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"polyglot-rosetta/pkg/constants"
)

var nodeIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type rawItemRef struct {
	ID    string `json:"id"`
	Order int    `json:"order"`
}

type rawRootInfo struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Items       []rawItemRef `json:"items"`
}

type rawNodeInfo struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Difficulty  string       `json:"difficulty"`
	Description string       `json:"description"`
	Items       []rawItemRef `json:"items"`
}

type masterIndex struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Categories  []categoryOutput `json:"categories"`
}

type categoryOutput struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Order       int              `json:"order"`
	Categories  []categoryOutput `json:"categories,omitempty"`
	Concepts    []conceptOutput  `json:"concepts"`
}

type conceptOutput struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Difficulty  string `json:"difficulty"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

func generateIndex(opts options) error {
	fmt.Println("Generating master index.json from problems tree...")

	problemsDir := constants.ProblemsDir
	if _, err := os.Stat(problemsDir); err != nil {
		return fmt.Errorf("cannot access problems directory %q: %w", problemsDir, err)
	}

	root, err := readRoot(problemsDir)
	if err != nil {
		return err
	}

	index := masterIndex{
		ID:          root.ID,
		Title:       root.Title,
		Description: root.Description,
		Categories:  []categoryOutput{},
	}
	for _, categoryRef := range root.Items {
		category, err := buildCategory(problemsDir, problemsDir, categoryRef)
		if err != nil {
			return err
		}
		index.Categories = append(index.Categories, category)
	}
	sortCategories(index.Categories)

	if err := writeIndex(opts.output, index); err != nil {
		return err
	}
	fmt.Printf("wrote master index to %q\n", opts.output)
	return nil
}

func readRoot(problemsDir string) (rawRootInfo, error) {
	path := filepath.Join(problemsDir, "info.json")
	data, err := readMetadata(path)
	if err != nil {
		return rawRootInfo{}, fmt.Errorf("failed to read root info.json at %q: %w", path, err)
	}

	var root rawRootInfo
	if err := json.Unmarshal(data, &root); err != nil {
		return rawRootInfo{}, fmt.Errorf("failed to parse root info.json: %w", err)
	}
	return root, nil
}

func buildCategory(rootDir, parentDir string, ref rawItemRef) (categoryOutput, error) {
	categoryDir, err := resolveChildDir(rootDir, parentDir, ref.ID)
	if err != nil {
		return categoryOutput{}, fmt.Errorf("invalid category %q: %w", ref.ID, err)
	}

	category, err := readNode(categoryDir)
	if err != nil {
		return categoryOutput{}, fmt.Errorf("failed to read category %q: %w", ref.ID, err)
	}
	if category.ID != ref.ID {
		return categoryOutput{}, fmt.Errorf("category metadata ID %q does not match reference %q", category.ID, ref.ID)
	}
	if category.Difficulty != "" {
		return categoryOutput{}, fmt.Errorf("expected category %q but found a concept", ref.ID)
	}

	output := categoryOutput{
		ID:          category.ID,
		Title:       category.Title,
		Description: category.Description,
		Order:       ref.Order,
		Concepts:    []conceptOutput{},
	}
	for _, childRef := range category.Items {
		childDir, err := resolveChildDir(rootDir, categoryDir, childRef.ID)
		if err != nil {
			return categoryOutput{}, fmt.Errorf("invalid child %q under category %q: %w", childRef.ID, ref.ID, err)
		}
		child, err := readNode(childDir)
		if err != nil {
			return categoryOutput{}, fmt.Errorf("failed to read child %q under category %q: %w", childRef.ID, ref.ID, err)
		}
		if child.ID != childRef.ID {
			return categoryOutput{}, fmt.Errorf("child metadata ID %q does not match reference %q", child.ID, childRef.ID)
		}

		if child.Difficulty == "" {
			nested, err := buildCategory(rootDir, categoryDir, childRef)
			if err != nil {
				return categoryOutput{}, err
			}
			output.Categories = append(output.Categories, nested)
			continue
		}

		output.Concepts = append(output.Concepts, conceptOutput{
			ID:          child.ID,
			Title:       child.Title,
			Difficulty:  child.Difficulty,
			Description: child.Description,
			Order:       childRef.Order,
		})
	}

	sortCategories(output.Categories)
	sort.SliceStable(output.Concepts, func(i, j int) bool {
		return output.Concepts[i].Order < output.Concepts[j].Order
	})
	return output, nil
}

func readNode(dir string) (rawNodeInfo, error) {
	path := filepath.Join(dir, "info.json")
	data, err := readMetadata(path)
	if err != nil {
		return rawNodeInfo{}, err
	}

	var node rawNodeInfo
	if err := json.Unmarshal(data, &node); err != nil {
		return rawNodeInfo{}, fmt.Errorf("failed to parse %q: %w", path, err)
	}
	return node, nil
}

func readMetadata(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("metadata file %q must not be a symbolic link", path)
	}
	return os.ReadFile(path)
}

func resolveChildDir(rootDir, parentDir, id string) (string, error) {
	if !nodeIDPattern.MatchString(id) {
		return "", fmt.Errorf("ID must be lowercase kebab-case")
	}

	root, err := filepath.EvalSymlinks(rootDir)
	if err != nil {
		return "", fmt.Errorf("resolve catalog root: %w", err)
	}
	parent, err := filepath.EvalSymlinks(parentDir)
	if err != nil {
		return "", fmt.Errorf("resolve parent directory: %w", err)
	}
	if !pathWithin(root, parent) {
		return "", fmt.Errorf("parent directory resolves outside catalog root")
	}

	candidate := filepath.Join(parent, id)
	info, err := os.Lstat(candidate)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("directory %q must not be a symbolic link", candidate)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%q is not a directory", candidate)
	}

	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve child directory: %w", err)
	}
	if !pathWithin(root, resolved) {
		return "", fmt.Errorf("child directory resolves outside catalog root")
	}
	return resolved, nil
}

func pathWithin(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func sortCategories(categories []categoryOutput) {
	sort.SliceStable(categories, func(i, j int) bool {
		return categories[i].Order < categories[j].Order
	})
}

func writeIndex(path string, index masterIndex) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create output directory %q: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal master index: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write master index to %q: %w", path, err)
	}
	return nil
}
