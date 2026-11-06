package sitecmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var themePaths = []string{
	"404.html",
	"_layouts/default.html",
	"_layouts/release.html",
	"_includes/breadcrumb.html",
	"_includes/head.html",
	"_includes/header.html",
	"_includes/footer.html",
	"_plugins",
	"_sass",
	"assets/css/light.scss",
	"assets/css/dark.scss",
	"assets/fonts",
	"assets/favicon",
	"assets/emoji",
	"assets/icons",
}

func runStage(opts options) error {
	_, err := prepareSite(opts)
	return err
}

func prepareSite(opts options) (string, error) {
	root, err := resolvePortfolioRoot(opts.source)
	if err != nil {
		return "", err
	}
	if err := stageTheme(root, opts.output); err != nil {
		return "", err
	}
	if err := stageCatalog(opts.catalog, opts.output); err != nil {
		return "", err
	}
	return root, nil
}

func stageCatalog(bundleDir, destRoot string) error {
	if strings.TrimSpace(bundleDir) == "" {
		return nil
	}
	entries, err := os.ReadDir(bundleDir)
	if err != nil {
		return fmt.Errorf("read catalog bundle %s: %w", bundleDir, err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("catalog bundle %s is empty; run just release-catalog", bundleDir)
	}
	dest := filepath.Join(destRoot, "catalog")
	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("remove old staged catalog %s: %w", dest, err)
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if err := copyPath(filepath.Join(bundleDir, entry.Name()), filepath.Join(dest, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func stageTheme(sourceRoot, destRoot string) error {
	for _, rel := range themePaths {
		if err := copyPath(filepath.Join(sourceRoot, rel), filepath.Join(destRoot, rel)); err != nil {
			return err
		}
	}
	for _, font := range []string{"MesloLGSNerdFont-Regular.ttf", "MesloLGSNerdFont-Bold.ttf"} {
		if !regularFile(filepath.Join(destRoot, "assets", "fonts", font)) {
			return fmt.Errorf("MesloLGS Nerd Font Regular and Bold must be present in assets/fonts (missing %s)", font)
		}
	}
	return nil
}

func copyPath(source, target string) error {
	info, err := os.Lstat(source)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("missing required theme file: %s", source)
		}
		return err
	}
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("remove %s: %w", target, err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return copyEntry(source, target, info)
}

func copyEntry(source, target string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(source)
		if err != nil {
			return err
		}
		return os.Symlink(link, target)
	}
	if info.IsDir() {
		if err := os.Mkdir(target, info.Mode().Perm()); err != nil {
			return err
		}
		entries, err := os.ReadDir(source)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			child := filepath.Join(source, entry.Name())
			childInfo, err := os.Lstat(child)
			if err != nil {
				return err
			}
			if err := copyEntry(child, filepath.Join(target, entry.Name()), childInfo); err != nil {
				return err
			}
		}
		return nil
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("unsupported file type: %s", source)
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func regularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
