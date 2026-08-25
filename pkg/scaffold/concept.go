package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"polyglot-rosetta/pkg/constants"
	"polyglot-rosetta/pkg/schema"
)

func (s *Scaffolder) CreateConcept(opts CreateConceptOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	if err := s.ValidateParent(opts.ParentPath); err != nil {
		return err
	}

	parentDir := s.resolvePath(opts.ParentPath)
	targetDir := filepath.Join(parentDir, opts.ID)
	if fileExists(targetDir) {
		return fmt.Errorf("target concept directory already exists: %s", targetDir)
	}

	tmplConceptDir := filepath.Join(s.TemplateDir, "concept")
	if !fileExists(tmplConceptDir) {
		return fmt.Errorf("concept template directory missing at: %s", tmplConceptDir)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create concept directory %s: %w", targetDir, err)
	}

	// 1. Copy required metadata and static files directly
	staticFiles := []string{"info.json", "testcases.json", "README.md", "PROBLEM.md"}
	for _, fileName := range staticFiles {
		srcFile := filepath.Join(tmplConceptDir, fileName)
		dstFile := filepath.Join(targetDir, fileName)

		if !fileExists(srcFile) {
			return fmt.Errorf("required template file missing: %s", srcFile)
		}

		if err := s.copyConceptTemplateFile(srcFile, dstFile, opts); err != nil {
			return err
		}
	}

	// 2. Copy language starters driven directly by schema definitions
	for _, lang := range schema.SupportedLanguages {
		srcPath := filepath.Join(tmplConceptDir, lang.TemplateRelPath())
		dstPath := filepath.Join(targetDir, lang.TargetRelPath())

		if !fileExists(srcPath) {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", lang, err)
		}

		if err := s.copyConceptTemplateFile(srcPath, dstPath, opts); err != nil {
			return err
		}
	}

	// 3. Register in parent's items array
	return s.registerInParent(opts.ParentPath, opts.ID)
}

func (s *Scaffolder) copyConceptTemplateFile(src, dst string, opts CreateConceptOptions) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read concept template %s: %w", src, err)
	}

	desc := opts.Description
	if desc == "" {
		desc = fmt.Sprintf("Implementation of %s across multiple languages.", opts.Title)
	}

	diff := string(opts.Difficulty)
	if diff == "" {
		diff = constants.ConceptDifficulty
	}

	content := string(data)
	content = strings.ReplaceAll(content, constants.TmplID, opts.ID)
	content = strings.ReplaceAll(content, constants.TmplTitle, opts.Title)
	content = strings.ReplaceAll(content, constants.TmplDifficulty, diff)
	content = strings.ReplaceAll(content, constants.TmplDescription, desc)

	return os.WriteFile(dst, []byte(content), 0644)
}
