package bundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"strings"
	"testing"
	"testing/fstest"
)

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
