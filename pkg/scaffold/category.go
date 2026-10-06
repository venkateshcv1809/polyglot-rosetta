package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"polyglot-rosetta/pkg/constants"
)

func (s *Scaffolder) CreateCategory(opts CreateCategoryOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	if err := s.ValidateParent(opts.ParentPath); err != nil {
		return err
	}

	parentDir := s.resolvePath(opts.ParentPath)
	targetDir := filepath.Join(parentDir, opts.ID)
	if fileExists(targetDir) {
		return fmt.Errorf("target category directory already exists: %s", targetDir)
	}

	categoryTmplPath := filepath.Join(s.TemplateDir, "category", "info.json")
	if !fileExists(categoryTmplPath) {
		return fmt.Errorf("required category template file missing: %s", categoryTmplPath)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
	}

	if err := s.copyCategoryTemplateFile(categoryTmplPath, filepath.Join(targetDir, "info.json"), opts); err != nil {
		return err
	}

	return s.registerInParent(opts.ParentPath, opts.ID)
}

func (s *Scaffolder) copyCategoryTemplateFile(src, dst string, opts CreateCategoryOptions) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read category template %s: %w", src, err)
	}

	content := string(data)
	content = strings.ReplaceAll(content, constants.TmplID, opts.ID)
	content = strings.ReplaceAll(content, constants.TmplTitle, opts.Title)
	content = strings.ReplaceAll(content, constants.TmplDescription, opts.Description)

	return os.WriteFile(dst, []byte(content), 0644)
}
