package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CreateCategory scaffolds a category and registers it with its parent.
func (s *Scaffolder) CreateCategory(opts CreateCategoryOptions) (err error) {
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
		return fmt.Errorf("target category directory already exists: %s", targetDir)
	}

	categoryTmplPath := filepath.Join(s.TemplateDir, "category", "info.json")
	exists, err = pathExists(categoryTmplPath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("required category template file missing: %s", categoryTmplPath)
	}

	stagingDir, err := os.MkdirTemp(parentDir, "."+opts.ID+"-*")
	if err != nil {
		return fmt.Errorf("create staging directory for category %q: %w", opts.ID, err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(stagingDir); cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("remove staging directory %q: %w", stagingDir, cleanupErr))
		}
	}()

	if err := s.copyCategoryTemplateFile(categoryTmplPath, filepath.Join(stagingDir, "info.json"), opts); err != nil {
		return err
	}
	if err := os.Rename(stagingDir, targetDir); err != nil {
		return fmt.Errorf("commit category directory %s: %w", targetDir, err)
	}

	if err := s.registerInParent(opts.ParentPath, opts.ID); err != nil {
		cleanupErr := os.RemoveAll(targetDir)
		if cleanupErr != nil {
			cleanupErr = fmt.Errorf("remove unregistered category %q: %w", targetDir, cleanupErr)
		}
		return errors.Join(err, cleanupErr)
	}
	return nil
}

func (s *Scaffolder) copyCategoryTemplateFile(src, dst string, opts CreateCategoryOptions) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read category template %s: %w", src, err)
	}

	content, err := renderTemplate(data, map[string]string{
		tmplID:          opts.ID,
		tmplTitle:       opts.Title,
		tmplDescription: opts.Description,
	}, filepath.Ext(dst) == ".json")
	if err != nil {
		return fmt.Errorf("failed to render category template %s: %w", src, err)
	}

	return os.WriteFile(dst, content, 0644)
}
