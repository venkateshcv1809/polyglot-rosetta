package version

import (
	"strings"
	"testing"
)

func TestEmbeddedMetadata(t *testing.T) {
	info, err := parseMetadata(metadataJSON)
	if err != nil {
		t.Fatal(err)
	}
	if Value != info.Version || MinimumEngine() != info.MinEngine || IsDevelopment() != (info.Version == "dev") {
		t.Fatalf("embedded metadata = %+v, Value = %q, minimum = %q", info, Value, MinimumEngine())
	}
}

func TestIsDevelopment(t *testing.T) {
	previous := Value
	t.Cleanup(func() { Value = previous })
	for _, value := range []string{"", "dev", "vdev", " dev ", "v0.1.0"} {
		Value = value
		if got, want := IsDevelopment(), value != "v0.1.0"; got != want {
			t.Fatalf("IsDevelopment(%q) = %t, want %t", value, got, want)
		}
	}
}

func TestParseMetadata(t *testing.T) {
	for _, value := range []string{"v0.3.0", "dev"} {
		info, err := parseMetadata([]byte(`{"version":"` + value + `","min_engine":"v0.1.0"}`))
		if err != nil || info.Version != value || info.MinEngine != "v0.1.0" {
			t.Fatalf("metadata = %+v, error = %v", info, err)
		}
	}
	for _, data := range []string{
		"{", `{}`, `{"version":"invalid","min_engine":"v0.1.0"}`,
		`{"version":"v0.3.0"}`, `{"version":"v0.3.0","min_engine":1}`,
		`{"version":"v0.3.0","min_engine":" "}`,
		`{"version":"dev","min_engine":"dev"}`,
	} {
		if _, err := parseMetadata([]byte(data)); err == nil || !strings.Contains(err.Error(), "version.json") {
			t.Fatalf("parseMetadata(%q): error = %v", data, err)
		}
	}
}
