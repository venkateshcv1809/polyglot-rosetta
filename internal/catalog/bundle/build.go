// Package bundle builds and loads distributable catalog archives.
package bundle

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/config"
	"polyglot-rosetta/internal/version"
)

// Build writes the metadata index and one archive per registered root category.
func Build(problemsDir, templatesDir, outputDir string) error {
	index, err := catalog.Build(problemsDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create bundle directory: %w", err)
	}
	encoded, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("encode index: %w", err)
	}
	encoded = append(encoded, '\n')
	indexAsset, err := writeBytes(outputDir, config.CatalogIndexFileName, encoded)
	if err != nil {
		return err
	}
	templateAsset, err := writeArchive(outputDir, config.TemplatesArchiveName, templatesDir, "")
	if err != nil {
		return err
	}
	result := Manifest{SchemaVersion: 1, MinEngine: version.MinimumEngine(), Index: indexAsset, Templates: templateAsset, Categories: []Asset{}}
	for _, category := range index.Categories {
		name := category.ID + config.CategoryArchiveSuffix
		categoryAsset, err := writeArchive(outputDir, name, filepath.Join(problemsDir, category.ID), category.ID)
		if err != nil {
			return err
		}
		result.Categories = append(result.Categories, categoryAsset)
	}
	sort.Slice(result.Categories, func(i, j int) bool { return result.Categories[i].Path < result.Categories[j].Path })
	encoded, err = json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode bundle manifest: %w", err)
	}
	_, err = writeBytes(outputDir, config.ManifestFileName, append(encoded, '\n'))
	return err
}

func writeArchive(outputDir, name, sourceDir, prefix string) (Asset, error) {
	file, err := os.CreateTemp(outputDir, config.BundleTempPattern)
	if err != nil {
		return Asset{}, fmt.Errorf("create %s: %w", name, err)
	}
	defer os.Remove(file.Name())
	zip := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(zip)
	root := os.DirFS(sourceDir)
	err = fs.WalkDir(root, ".", func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("bundle source %q is not a regular file", filePath)
		}
		name := path.Join(prefix, filepath.ToSlash(filePath))
		if !fs.ValidPath(name) || strings.HasPrefix(name, "../") {
			return fmt.Errorf("invalid archive path %q", name)
		}
		if err := tarWriter.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: info.Size()}); err != nil {
			return err
		}
		input, err := root.Open(filePath)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tarWriter, input)
		closeErr := input.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	if closeErr := tarWriter.Close(); err == nil {
		err = closeErr
	}
	if closeErr := zip.Close(); err == nil {
		err = closeErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return Asset{}, fmt.Errorf("build %s: %w", name, err)
	}
	return recordFile(outputDir, name, file.Name())
}

func writeBytes(outputDir, name string, data []byte) (Asset, error) {
	file, err := os.CreateTemp(outputDir, config.BundleTempPattern)
	if err != nil {
		return Asset{}, err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(data); err != nil {
		file.Close()
		return Asset{}, err
	}
	if err := file.Close(); err != nil {
		return Asset{}, err
	}
	return recordFile(outputDir, name, file.Name())
}

func recordFile(outputDir, name, temporary string) (Asset, error) {
	file, err := os.Open(temporary)
	if err != nil {
		return Asset{}, err
	}
	defer file.Close()
	hasher := sha256.New()
	size, err := io.Copy(hasher, file)
	if err != nil {
		return Asset{}, err
	}
	if err := os.Chmod(temporary, 0o644); err != nil {
		return Asset{}, err
	}
	if err := os.Rename(temporary, filepath.Join(outputDir, name)); err != nil {
		return Asset{}, err
	}
	return Asset{Path: name, Size: size, SHA256: hex.EncodeToString(hasher.Sum(nil))}, nil
}
