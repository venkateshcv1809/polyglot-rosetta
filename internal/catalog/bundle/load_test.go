package bundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"strings"
	"testing"
	"testing/fstest"
)

func TestParseManifestAcceptsOptionalMinEngine(t *testing.T) {
	hash := strings.Repeat("a", 64)
	base := `{"schema_version":1,"index":{"path":"index.json","size":1,"sha256":"` + hash + `"},"templates":{"path":"templates.tar.gz","size":1,"sha256":"` + hash + `"},"categories":[{"path":"fundamentals.tar.gz","size":1,"sha256":"` + hash + `"}]}`
	without, err := ParseManifest([]byte(base))
	if err != nil || without.MinEngine != "" {
		t.Fatalf("without min_engine: %+v, %v", without, err)
	}
	with, err := ParseManifest([]byte(`{"schema_version":1,"min_engine":"v1.2.3","index":{"path":"index.json","size":1,"sha256":"` + hash + `"},"templates":{"path":"templates.tar.gz","size":1,"sha256":"` + hash + `"},"categories":[{"path":"fundamentals.tar.gz","size":1,"sha256":"` + hash + `"}]}`))
	if err != nil || with.MinEngine != "v1.2.3" {
		t.Fatalf("with min_engine: %+v, %v", with, err)
	}
}

func TestCompatibleEngine(t *testing.T) {
	if err := CompatibleEngine("0.1.0", ""); err != nil {
		t.Fatalf("empty min_engine: %v", err)
	}
	if err := CompatibleEngine("dev", "v9.0.0"); err != nil {
		t.Fatalf("dev should pass: %v", err)
	}
	if err := CompatibleEngine("v1.2.0", "v1.2.0"); err != nil {
		t.Fatalf("equal versions: %v", err)
	}
	if err := CompatibleEngine("v1.3.0", "1.2.0"); err != nil {
		t.Fatalf("newer engine: %v", err)
	}
	if err := CompatibleEngine("0.1.0", "v0.2.0"); err == nil {
		t.Fatal("older engine accepted")
	}
}

func TestParseManifestRejectsUnsafeCategoryPath(t *testing.T) {
	hash := strings.Repeat("a", 64)
	data := []byte(`{"schema_version":1,"index":{"path":"index.json","size":1,"sha256":"` + hash + `"},"templates":{"path":"templates.tar.gz","size":1,"sha256":"` + hash + `"},"categories":[{"path":"../escape.tar.gz","size":1,"sha256":"` + hash + `"}]}`)
	if _, err := ParseManifest(data); err == nil {
		t.Fatal("unsafe archive path accepted")
	}
}

func TestAddArchiveRejectsTraversalAndLinks(t *testing.T) {
	for _, entry := range []tar.Header{
		{Name: "../escape", Typeflag: tar.TypeReg, Size: 1},
		{Name: "source/link", Typeflag: tar.TypeSymlink, Linkname: "../../escape"},
		{Name: "other/file", Typeflag: tar.TypeReg, Size: 1},
	} {
		t.Run(entry.Name, func(t *testing.T) {
			var compressed bytes.Buffer
			zip := gzip.NewWriter(&compressed)
			writer := tar.NewWriter(zip)
			if err := writer.WriteHeader(&entry); err != nil {
				t.Fatal(err)
			}
			if entry.Size != 0 {
				if _, err := writer.Write([]byte("x")); err != nil {
					t.Fatal(err)
				}
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			if err := zip.Close(); err != nil {
				t.Fatal(err)
			}
			if err := addArchive(fstest.MapFS{}, compressed.Bytes(), "source/"); err == nil {
				t.Fatal("unsafe archive entry accepted")
			}
		})
	}
}

func TestAddArchiveRejectsSizeMismatch(t *testing.T) {
	var compressed bytes.Buffer
	zip := gzip.NewWriter(&compressed)
	writer := tar.NewWriter(zip)
	if err := writer.WriteHeader(&tar.Header{Name: "fundamentals/info.json", Typeflag: tar.TypeReg, Size: 100, Mode: 0o644}); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("{}\n")); err != nil && !strings.Contains(err.Error(), "write too long") && !strings.Contains(err.Error(), "short write") {
		// tar.Writer may reject undersized members; still close and load.
	}
	_ = writer.Close()
	_ = zip.Close()
	if err := addArchive(fstest.MapFS{}, compressed.Bytes(), "fundamentals/"); err == nil {
		t.Fatal("size mismatch accepted")
	}
}
