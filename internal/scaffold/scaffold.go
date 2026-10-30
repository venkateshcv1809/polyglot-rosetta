// Package scaffold creates and registers catalog categories and concepts from
// repository templates.
package scaffold

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/config"
)

const (
	rootID          = "root"
	rootTitle       = "Polyglot Rosetta"
	rootDescription = "Catalog"
	tmplID          = "{{ID}}"
	tmplTitle       = "{{TITLE}}"
	tmplDescription = "{{DESCRIPTION}}"
	tmplDifficulty  = "{{DIFFICULTY}}"
)

// Scaffolder creates catalog nodes from the repository templates.
type Scaffolder struct {
	RootDir     string
	TemplateDir string
}

// New returns a Scaffolder using repository defaults for empty paths.
func New(rootDir, templateDir string) *Scaffolder {
	if rootDir == "" {
		rootDir = config.CatalogDir
	}
	if templateDir == "" {
		templateDir = config.TemplateDir
	}
	return &Scaffolder{
		RootDir:     rootDir,
		TemplateDir: templateDir,
	}
}

func (s *Scaffolder) resolveParentDir(relPath string) (string, error) {
	relPath = strings.TrimSpace(relPath)
	if filepath.IsAbs(relPath) {
		return "", fmt.Errorf("parent path %q must be relative to the catalog", relPath)
	}

	cleaned := filepath.Clean(relPath)
	if cleaned == "" || cleaned == "." {
		cleaned = "."
	} else if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("parent path %q escapes the catalog", relPath)
	}

	root, err := filepath.Abs(s.RootDir)
	if err != nil {
		return "", fmt.Errorf("resolve catalog root %q: %w", s.RootDir, err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve catalog root %q: %w", root, err)
	}
	resolvedParent, err := filepath.EvalSymlinks(filepath.Join(root, cleaned))
	if err != nil {
		return "", fmt.Errorf("resolve parent path %q: %w", relPath, err)
	}
	if !pathWithin(resolvedRoot, resolvedParent) {
		return "", fmt.Errorf("parent path %q resolves outside the catalog", relPath)
	}
	info, err := os.Stat(resolvedParent)
	if err != nil {
		return "", fmt.Errorf("access parent path %q: %w", relPath, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("parent path %q is not a directory", relPath)
	}
	return resolvedParent, nil
}

func (s *Scaffolder) ensureRootInfoExists() error {
	if err := os.MkdirAll(s.RootDir, 0755); err != nil {
		return fmt.Errorf("failed to create root dir %s: %w", s.RootDir, err)
	}

	infoPath := filepath.Join(s.RootDir, config.CatalogInfoFileName)
	exists, err := pathExists(infoPath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	rootTmplPath := filepath.Join(s.TemplateDir, "category", config.CatalogInfoFileName)
	exists, err = pathExists(rootTmplPath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("required root template file missing at: %s", rootTmplPath)
	}

	data, err := os.ReadFile(rootTmplPath)
	if err != nil {
		return fmt.Errorf("failed to read root template %s: %w", rootTmplPath, err)
	}

	content, err := renderTemplate(data, map[string]string{
		tmplID:          rootID,
		tmplTitle:       rootTitle,
		tmplDescription: rootDescription,
	}, true)
	if err != nil {
		return fmt.Errorf("failed to render root template %s: %w", rootTmplPath, err)
	}

	return writeFileAtomic(infoPath, content, 0o644)
}

// ValidateParent verifies that a scaffold target is an existing category.
func (s *Scaffolder) ValidateParent(parentRelPath string) error {
	if err := s.ensureRootInfoExists(); err != nil {
		return err
	}

	parentDir, err := s.resolveParentDir(parentRelPath)
	if err != nil {
		return err
	}

	infoPath := filepath.Join(parentDir, config.CatalogInfoFileName)
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return fmt.Errorf("failed to read parent info.json at %s: %w", infoPath, err)
	}

	var node struct {
		Difficulty catalog.Difficulty `json:"difficulty"`
		Items      json.RawMessage    `json:"items"`
	}
	if err := json.Unmarshal(data, &node); err != nil {
		return fmt.Errorf("failed to parse parent info.json at %s: %w", infoPath, err)
	}
	if node.Difficulty != "" {
		return fmt.Errorf("cannot scaffold under %q: parent is a leaf concept", parentRelPath)
	}
	if len(node.Items) == 0 {
		return fmt.Errorf("cannot scaffold under %q: parent metadata has no items array", parentRelPath)
	}

	return nil
}

func (s *Scaffolder) registerInParent(parentRelPath, childID string) error {
	parentDir, err := s.resolveParentDir(parentRelPath)
	if err != nil {
		return err
	}
	infoPath := filepath.Join(parentDir, config.CatalogInfoFileName)

	data, err := os.ReadFile(infoPath)
	if err != nil {
		return fmt.Errorf("failed to read parent info.json at %s: %w", infoPath, err)
	}

	var parent catalog.ParentInfo
	if err := json.Unmarshal(data, &parent); err != nil {
		return fmt.Errorf("failed to parse parent info.json at %s: %w", infoPath, err)
	}

	nextOrder := 1
	for _, item := range parent.Items {
		if item.ID == childID {
			return fmt.Errorf("child %q is already registered in parent %s", childID, infoPath)
		}
		if item.Order >= nextOrder {
			nextOrder = item.Order + 1
		}
	}

	parent.Items = append(parent.Items, catalog.ItemRef{
		ID:    childID,
		Order: nextOrder,
	})
	sort.Slice(parent.Items, func(i, j int) bool {
		if parent.Items[i].Order == parent.Items[j].Order {
			return parent.Items[i].ID < parent.Items[j].ID
		}
		return parent.Items[i].Order < parent.Items[j].Order
	})

	updatedData, err := json.MarshalIndent(parent, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal parent info.json: %w", err)
	}

	return writeFileAtomic(infoPath, append(updatedData, '\n'), 0o644)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("access %q: %w", path, err)
}

func pathWithin(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func writeFileAtomic(path string, data []byte, mode fs.FileMode) (err error) {
	temp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("create temporary file for %q: %w", path, err)
	}
	tempPath := temp.Name()
	defer func() {
		if closeErr := temp.Close(); closeErr != nil && !errors.Is(closeErr, os.ErrClosed) {
			err = errors.Join(err, fmt.Errorf("close temporary file for %q: %w", path, closeErr))
		}
		if removeErr := os.Remove(tempPath); removeErr != nil && !errors.Is(removeErr, fs.ErrNotExist) {
			err = errors.Join(err, fmt.Errorf("remove temporary file for %q: %w", path, removeErr))
		}
	}()

	if err := temp.Chmod(mode); err != nil {
		return fmt.Errorf("set permissions on temporary file for %q: %w", path, err)
	}
	if _, err := temp.Write(data); err != nil {
		return fmt.Errorf("write temporary file for %q: %w", path, err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("sync temporary file for %q: %w", path, err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary file for %q: %w", path, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace %q: %w", path, err)
	}
	return nil
}

func renderTemplate(data []byte, replacements map[string]string, jsonTemplate bool) ([]byte, error) {
	content := string(data)
	for placeholder, value := range replacements {
		if jsonTemplate {
			encoded, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("encode replacement for %s: %w", placeholder, err)
			}
			value = string(encoded[1 : len(encoded)-1])
		}
		content = strings.ReplaceAll(content, placeholder, value)
	}

	result := []byte(content)
	if jsonTemplate && !json.Valid(result) {
		return nil, fmt.Errorf("rendered template is not valid JSON")
	}
	return result, nil
}
