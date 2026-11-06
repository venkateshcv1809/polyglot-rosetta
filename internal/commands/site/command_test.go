package sitecmd

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommand(t *testing.T) {
	command := Command()
	if command.Name != commandName || command.Summary != commandSummary {
		t.Fatalf("command = %#v", command)
	}
	wantNames := []string{"stage", "dev"}
	if len(command.Subcommands) != len(wantNames) {
		t.Fatalf("subcommands = %d, want %d", len(command.Subcommands), len(wantNames))
	}
	for i, want := range wantNames {
		if got := command.Subcommands[i].Name; got != want {
			t.Fatalf("subcommand %d = %q, want %q", i, got, want)
		}
	}
}

func TestExecute_Help(t *testing.T) {
	if err := Execute([]string{"--help"}); err != nil {
		t.Fatalf("help: %v", err)
	}
}

func TestExecute_MissingCommand(t *testing.T) {
	if err := Execute(nil); err == nil || !strings.Contains(err.Error(), "missing command") {
		t.Fatalf("error = %v, want missing command", err)
	}
}

func TestExecute_StageHelp(t *testing.T) {
	if err := Execute([]string{"stage", "-h"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("got %v, want flag.ErrHelp", err)
	}
}

func TestParseSiteOptions_DefaultsOutput(t *testing.T) {
	t.Parallel()
	opts, err := parseSiteOptions("site stage", nil)
	if err != nil {
		t.Fatal(err)
	}
	if opts.output != defaultPagesRoot {
		t.Fatalf("output = %q, want %q", opts.output, defaultPagesRoot)
	}
}

func TestRunStage_CopiesThemeAndCatalog(t *testing.T) {
	source := writeThemeFixture(t, t.TempDir())
	catalog := writeCatalogFixture(t, t.TempDir())
	dest := t.TempDir()
	if err := runStage(options{source: source, output: dest, catalog: catalog}); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{
		"404.html",
		"_layouts/default.html",
		"_layouts/release.html",
		"catalog/manifest.json",
		"assets/emoji/construction.webp",
		"assets/fonts/MesloLGSNerdFont-Regular.ttf",
	} {
		if !regularFile(filepath.Join(dest, rel)) {
			t.Errorf("missing staged file %s", rel)
		}
	}
}

func TestRunStage_MissingFontFails(t *testing.T) {
	source := writeThemeFixture(t, t.TempDir())
	if err := os.Remove(filepath.Join(source, "assets", "fonts", "MesloLGSNerdFont-Bold.ttf")); err != nil {
		t.Fatal(err)
	}
	err := runStage(options{source: source, output: t.TempDir(), catalog: writeCatalogFixture(t, t.TempDir())})
	if err == nil || !strings.Contains(err.Error(), "MesloLGSNerdFont-Bold.ttf") {
		t.Fatalf("error = %v, want missing bold font", err)
	}
}

func TestRunStage_ExplicitSourceDoesNotClone(t *testing.T) {
	source := writeThemeFixture(t, t.TempDir())
	original := clonePortfolio
	t.Cleanup(func() { clonePortfolio = original })
	clonePortfolio = func(string) error {
		t.Fatal("clone should not run when --source is set")
		return nil
	}
	if err := runStage(options{source: source, output: t.TempDir(), catalog: writeCatalogFixture(t, t.TempDir())}); err != nil {
		t.Fatal(err)
	}
}

func TestRunStage_InvalidSourceDoesNotClone(t *testing.T) {
	original := clonePortfolio
	t.Cleanup(func() { clonePortfolio = original })
	clonePortfolio = func(string) error {
		t.Fatal("clone should not run when --source is invalid")
		return nil
	}
	err := runStage(options{source: t.TempDir(), output: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "missing required theme file") {
		t.Fatalf("error = %v, want missing required theme file", err)
	}
}

func TestRunStage_UsesPortfolioRootEnv(t *testing.T) {
	source := writeThemeFixture(t, t.TempDir())
	t.Setenv("PORTFOLIO_ROOT", source)
	original := clonePortfolio
	t.Cleanup(func() { clonePortfolio = original })
	clonePortfolio = func(string) error {
		t.Fatal("clone should not run when PORTFOLIO_ROOT is set")
		return nil
	}
	if err := runStage(options{output: t.TempDir(), catalog: writeCatalogFixture(t, t.TempDir())}); err != nil {
		t.Fatal(err)
	}
}

func TestRunStage_CloneFallback(t *testing.T) {
	cloneSource := writeThemeFixture(t, t.TempDir())
	original := clonePortfolio
	t.Cleanup(func() { clonePortfolio = original })
	var clonedTo string
	clonePortfolio = func(dest string) error {
		clonedTo = dest
		return copyPath(cloneSource, dest)
	}
	t.Setenv("PORTFOLIO_ROOT", "")
	t.Setenv("TMPDIR", t.TempDir())
	base := t.TempDir()
	working := filepath.Join(base, "polyglot-rosetta")
	if err := os.Mkdir(working, 0o755); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Errorf("restore cwd: %v", err)
		}
	})
	if err := os.Chdir(working); err != nil {
		t.Fatal(err)
	}
	if err := runStage(options{output: t.TempDir(), catalog: writeCatalogFixture(t, t.TempDir())}); err != nil {
		t.Fatal(err)
	}
	if clonedTo == "" {
		t.Fatal("expected clone fallback")
	}
}

func TestRunDev_ServesAfterStaging(t *testing.T) {
	source := writeThemeFixture(t, t.TempDir())
	if err := os.WriteFile(filepath.Join(source, "Gemfile"), []byte("source \"https://rubygems.org\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	catalog := writeCatalogFixture(t, t.TempDir())
	original := runJekyllServe
	t.Cleanup(func() { runJekyllServe = original })
	var gotRoot, gotPages, gotDest string
	runJekyllServe = func(portfolioRoot, pagesRoot, destination string) error {
		gotRoot, gotPages, gotDest = portfolioRoot, pagesRoot, destination
		return nil
	}
	if err := runDev(options{source: source, output: dest, catalog: catalog}); err != nil {
		t.Fatal(err)
	}
	absSource, err := filepath.Abs(source)
	if err != nil {
		t.Fatal(err)
	}
	absDest, err := filepath.Abs(dest)
	if err != nil {
		t.Fatal(err)
	}
	if gotRoot != absSource {
		t.Fatalf("portfolio root = %q, want %q", gotRoot, absSource)
	}
	if gotPages != absDest {
		t.Fatalf("pages root = %q, want %q", gotPages, absDest)
	}
	if gotDest == "" {
		t.Fatal("expected jekyll destination")
	}
	if !regularFile(filepath.Join(dest, "catalog", "manifest.json")) {
		t.Fatal("dev did not stage catalog")
	}
}

func writeThemeFixture(t *testing.T, root string) string {
	t.Helper()
	files := []string{
		"404.html",
		"_layouts/default.html",
		"_layouts/release.html",
		"_includes/breadcrumb.html",
		"_includes/head.html",
		"_includes/header.html",
		"_includes/footer.html",
		"_plugins/icon.rb",
		"_sass/_release.scss",
		"assets/css/light.scss",
		"assets/css/dark.scss",
		"assets/fonts/MesloLGSNerdFont-Regular.ttf",
		"assets/fonts/MesloLGSNerdFont-Bold.ttf",
		"assets/favicon/favicon.ico",
		"assets/emoji/construction.webp",
		"assets/icons/github.svg",
	}
	for _, rel := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(rel+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func writeCatalogFixture(t *testing.T, root string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}
