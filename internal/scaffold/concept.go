package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"polyglot-rosetta/internal/domain/catalog"
	"polyglot-rosetta/internal/domain/language"
)

// CreateConcept scaffolds metadata, documentation, and language starters for a concept.
func (s *Scaffolder) CreateConcept(opts CreateConceptOptions) (err error) {
	if err := opts.Validate(); err != nil {
		return err
	}

	if err := s.ValidateParent(opts.ParentPath); err != nil {
		return err
	}
	parentDir, err := s.resolveParentDir(opts.ParentPath)
	if err != nil {
		return err
	}
	targetDir := filepath.Join(parentDir, opts.ID)
	exists, err := pathExists(targetDir)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("target concept directory already exists: %s", targetDir)
	}

	tmplConceptDir := filepath.Join(s.TemplateDir, "concept")
	exists, err = pathExists(tmplConceptDir)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("concept template directory missing at: %s", tmplConceptDir)
	}

	staticFiles := []string{"info.json", "testcases.json", "README.md", "PROBLEM.md"}
	for _, fileName := range staticFiles {
		exists, err := pathExists(filepath.Join(tmplConceptDir, fileName))
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("required template file missing: %s", filepath.Join(tmplConceptDir, fileName))
		}
	}

	stagingDir, err := os.MkdirTemp(parentDir, "."+opts.ID+"-*")
	if err != nil {
		return fmt.Errorf("create staging directory for concept %q: %w", opts.ID, err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(stagingDir); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("remove staging directory %q: %w", stagingDir, cleanupErr))
		}
	}()

	// Copy required metadata and documentation into the staging directory.
	for _, fileName := range staticFiles {
		srcFile := filepath.Join(tmplConceptDir, fileName)
		dstFile := filepath.Join(stagingDir, fileName)

		if err := s.copyConceptTemplateFile(srcFile, dstFile, opts); err != nil {
			return err
		}
	}

	// New concepts start with reference source; practice starters use the shared
	// language template unless an author explicitly adds a concept override.
	for _, lang := range language.AllSupported() {
		srcPath := filepath.Join(tmplConceptDir, templateRelPath(lang))

		exists, err := pathExists(srcPath)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filepath.Join(stagingDir, targetRelPath(lang))), 0o755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", lang, err)
		}

		if err := s.copyConceptTemplateFile(srcPath, filepath.Join(stagingDir, targetRelPath(lang)), opts); err != nil {
			return err
		}
	}

	if err := os.Rename(stagingDir, targetDir); err != nil {
		return fmt.Errorf("commit concept directory %s: %w", targetDir, err)
	}
	if err := s.registerInParent(opts.ParentPath, opts.ID); err != nil {
		cleanupErr := os.RemoveAll(targetDir)
		if cleanupErr != nil {
			cleanupErr = fmt.Errorf("remove unregistered concept %q: %w", targetDir, cleanupErr)
		}
		return errors.Join(err, cleanupErr)
	}
	return nil
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
		diff = string(catalog.DifficultyEasy)
	}

	content, err := renderTemplate(data, map[string]string{
		tmplID:          opts.ID,
		tmplTitle:       opts.Title,
		tmplDifficulty:  diff,
		tmplDescription: desc,
	}, filepath.Ext(dst) == ".json")
	if err != nil {
		return fmt.Errorf("failed to render concept template %s: %w", src, err)
	}

	return os.WriteFile(dst, content, 0644)
}

func templateRelPath(lang language.Language) string {
	return filepath.Join(string(lang), "template."+lang.FileExtension())
}

func targetRelPath(lang language.Language) string {
	return filepath.Join(string(lang), "main."+lang.FileExtension())
}
