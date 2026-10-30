package bundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"reflect"
	"strings"
	"testing/fstest"

	"polyglot-rosetta/internal/catalog"
	"polyglot-rosetta/internal/catalog/content"
	"polyglot-rosetta/internal/config"
)

const (
	maxArchiveBytes = 128 << 20
	maxArchiveFiles = 10000
)

// Asset identifies a content file and its integrity metadata.
type Asset struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// Manifest describes one consistent release of independently cached content.
type Manifest struct {
	SchemaVersion int     `json:"schema_version"`
	MinEngine     string  `json:"min_engine,omitempty"`
	Index         Asset   `json:"index"`
	Templates     Asset   `json:"templates"`
	Categories    []Asset `json:"categories"`
}

// ParseManifest validates paths and integrity information before fetching assets.
func ParseManifest(data []byte) (Manifest, error) {
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode bundle manifest: %w", err)
	}
	if manifest.SchemaVersion != 1 || manifest.Index.Path != config.CatalogIndexFileName || manifest.Templates.Path != config.TemplatesArchiveName {
		return Manifest{}, fmt.Errorf("unsupported bundle manifest")
	}
	for _, asset := range append([]Asset{manifest.Index, manifest.Templates}, manifest.Categories...) {
		if asset.Size <= 0 || asset.Size > maxArchiveBytes || len(asset.SHA256) != 64 {
			return Manifest{}, fmt.Errorf("invalid integrity metadata for %q", asset.Path)
		}
		if _, err := hex.DecodeString(asset.SHA256); err != nil {
			return Manifest{}, fmt.Errorf("invalid digest for %q: %w", asset.Path, err)
		}
	}
	seen := map[string]bool{config.CatalogIndexFileName: true, config.TemplatesArchiveName: true}
	if len(manifest.Categories) > 1000 {
		return Manifest{}, fmt.Errorf("too many category archives")
	}
	for _, asset := range manifest.Categories {
		id := strings.TrimSuffix(asset.Path, config.CategoryArchiveSuffix)
		if !strings.HasSuffix(asset.Path, config.CategoryArchiveSuffix) || !validCategoryID(id) || seen[asset.Path] {
			return Manifest{}, fmt.Errorf("invalid category archive %q", asset.Path)
		}
		seen[asset.Path] = true
	}
	return manifest, nil
}

// Contents holds decoded files and their ready-to-use in-memory catalog.
type Contents struct {
	Problems  fstest.MapFS
	Templates fstest.MapFS
	Catalog   catalog.Catalog
}

// Load validates and combines metadata, shared templates, and category archives.
func Load(indexData, templateArchive []byte, categoryArchives map[string][]byte) (Contents, error) {
	var metadata catalog.Catalog
	if err := json.Unmarshal(indexData, &metadata); err != nil {
		return Contents{}, fmt.Errorf("decode bundle index: %w", err)
	}
	if metadata.ID == "" || metadata.Title == "" || len(categoryArchives) != len(metadata.Categories) {
		return Contents{}, fmt.Errorf("bundle index and category archives do not match")
	}
	root := catalog.ParentInfo{ID: metadata.ID, Title: metadata.Title, Description: metadata.Description, Items: []catalog.ItemRef{}}
	problems := fstest.MapFS{}
	for _, category := range metadata.Categories {
		if !validCategoryID(category.ID) {
			return Contents{}, fmt.Errorf("invalid category ID %q", category.ID)
		}
		name := category.ID + config.CategoryArchiveSuffix
		archive, ok := categoryArchives[name]
		if !ok {
			return Contents{}, fmt.Errorf("missing category archive %q", name)
		}
		if err := addArchive(problems, archive, category.ID+"/"); err != nil {
			return Contents{}, fmt.Errorf("decode %s: %w", name, err)
		}
		root.Items = append(root.Items, catalog.ItemRef{ID: category.ID, Order: category.Order})
	}
	rootData, err := json.Marshal(root)
	if err != nil {
		return Contents{}, err
	}
	problems[config.CatalogInfoFileName] = &fstest.MapFile{Data: rootData, Mode: 0o644}
	templates := fstest.MapFS{}
	if err := addArchive(templates, templateArchive, ""); err != nil {
		return Contents{}, fmt.Errorf("decode templates: %w", err)
	}
	actual, err := catalog.LoadMetadata(problems)
	if err != nil {
		return Contents{}, err
	}
	actualData, err := json.Marshal(actual)
	if err != nil {
		return Contents{}, fmt.Errorf("encode archive metadata: %w", err)
	}
	var indexed, rebuilt any
	if err := json.Unmarshal(indexData, &indexed); err != nil {
		return Contents{}, err
	}
	if err := json.Unmarshal(actualData, &rebuilt); err != nil {
		return Contents{}, err
	}
	if !reflect.DeepEqual(indexed, rebuilt) {
		return Contents{}, fmt.Errorf("bundle index does not match category archives")
	}
	loaded, err := content.Enrich(actual, problems, templates)
	if err != nil {
		return Contents{}, err
	}
	return Contents{Problems: problems, Templates: templates, Catalog: loaded}, nil
}

func addArchive(target fstest.MapFS, compressed []byte, prefix string) error {
	if len(compressed) == 0 || len(compressed) > maxArchiveBytes {
		return fmt.Errorf("archive size is outside supported limits")
	}
	zip, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return err
	}
	defer zip.Close()
	reader := tar.NewReader(zip)
	var total int64
	var count int
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := header.Name
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			return fmt.Errorf("unsupported archive entry %q", name)
		}
		if name == "." || !fs.ValidPath(name) || path.Clean(name) != name || strings.Contains(name, "\\") || (prefix != "" && !strings.HasPrefix(name, prefix)) {
			return fmt.Errorf("unsafe archive path %q", name)
		}
		if _, exists := target[name]; exists {
			return fmt.Errorf("duplicate archive path %q", name)
		}
		count++
		total += header.Size
		if header.Size < 0 || total > maxArchiveBytes || count > maxArchiveFiles {
			return fmt.Errorf("archive exceeds supported limits")
		}
		data, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		target[name] = &fstest.MapFile{Data: data, Mode: 0o644}
	}
}

func validCategoryID(id string) bool {
	if id == "" || id[0] == '-' || id[len(id)-1] == '-' {
		return false
	}
	previousHyphen := false
	for _, r := range id {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return false
		}
		if r == '-' && previousHyphen {
			return false
		}
		previousHyphen = r == '-'
	}
	return true
}
