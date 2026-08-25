package scaffold

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"polyglot-rosetta/pkg/constants"
	"polyglot-rosetta/pkg/schema"
)

type Scaffolder struct {
	RootDir     string
	TemplateDir string
}

func New(rootDir, templateDir string) *Scaffolder {
	if rootDir == "" {
		rootDir = constants.SrcDir
	}
	if templateDir == "" {
		templateDir = constants.TemplateDir
	}
	return &Scaffolder{
		RootDir:     rootDir,
		TemplateDir: templateDir,
	}
}

func (s *Scaffolder) resolvePath(relPath string) string {
	relPath = strings.TrimPrefix(strings.TrimSpace(relPath), "/")
	if relPath == "" || relPath == "." {
		return s.RootDir
	}
	return filepath.Join(s.RootDir, relPath)
}

func (s *Scaffolder) ensureRootInfoExists() error {
	if err := os.MkdirAll(s.RootDir, 0755); err != nil {
		return fmt.Errorf("failed to create root dir %s: %w", s.RootDir, err)
	}

	infoPath := filepath.Join(s.RootDir, "info.json")
	if fileExists(infoPath) {
		return nil
	}

	rootTmplPath := filepath.Join(s.TemplateDir, "root", "info.json")
	if !fileExists(rootTmplPath) {
		return fmt.Errorf("required root template file missing at: %s", rootTmplPath)
	}

	data, err := os.ReadFile(rootTmplPath)
	if err != nil {
		return fmt.Errorf("failed to read root template %s: %w", rootTmplPath, err)
	}

	content := string(data)
	content = strings.ReplaceAll(content, constants.TmplID, constants.RootID)
	content = strings.ReplaceAll(content, constants.TmplTitle, constants.RootTitle)
	content = strings.ReplaceAll(content, constants.TmplDescription, constants.RootDescription)

	return os.WriteFile(infoPath, []byte(content), 0644)
}

func (s *Scaffolder) ValidateParent(parentRelPath string) error {
	if err := s.ensureRootInfoExists(); err != nil {
		return err
	}

	parentDir := s.resolvePath(parentRelPath)
	if !fileExists(parentDir) {
		return fmt.Errorf("parent directory does not exist: %s", parentDir)
	}

	infoPath := filepath.Join(parentDir, "info.json")
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return fmt.Errorf("failed to read parent info.json at %s: %w", infoPath, err)
	}

	// Reject if parent is a concept
	var leaf schema.LeafInfo
	if err := json.Unmarshal(data, &leaf); err == nil && leaf.Difficulty != "" {
		return fmt.Errorf("cannot scaffold under %q: parent is a leaf concept", parentRelPath)
	}

	return nil
}

func (s *Scaffolder) registerInParent(parentRelPath, childID string) error {
	parentDir := s.resolvePath(parentRelPath)
	infoPath := filepath.Join(parentDir, "info.json")

	data, err := os.ReadFile(infoPath)
	if err != nil {
		return fmt.Errorf("failed to read parent info.json at %s: %w", infoPath, err)
	}

	var parent schema.ParentInfo
	if err := json.Unmarshal(data, &parent); err != nil {
		return fmt.Errorf("failed to parse parent info.json at %s: %w", infoPath, err)
	}

	for _, item := range parent.Items {
		if item.ID == childID {
			return fmt.Errorf("child %q is already registered in parent %s", childID, infoPath)
		}
	}

	nextOrder := len(parent.Items) + 1
	parent.Items = append(parent.Items, schema.ItemRef{
		ID:    childID,
		Order: nextOrder,
	})

	updatedData, err := json.MarshalIndent(parent, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal parent info.json: %w", err)
	}

	return os.WriteFile(infoPath, append(updatedData, '\n'), 0644)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
