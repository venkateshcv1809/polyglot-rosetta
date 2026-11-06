package sitecmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"polyglot-rosetta/internal/config"
)

var runJekyllServe = serveJekyll

func runDev(opts options) error {
	root, err := prepareSite(opts)
	if err != nil {
		return err
	}
	pagesRoot, err := filepath.Abs(opts.output)
	if err != nil {
		return err
	}
	destination := filepath.Join(os.TempDir(), config.SitePreviewDirName)
	return runJekyllServe(root, pagesRoot, destination)
}

func serveJekyll(portfolioRoot, pagesRoot, destination string) error {
	if !regularFile(filepath.Join(portfolioRoot, "Gemfile")) {
		return fmt.Errorf("portfolio checkout %s is missing Gemfile", portfolioRoot)
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	cmd := exec.Command("bundle", "exec", "jekyll", "serve",
		"--source", pagesRoot,
		"--destination", destination,
		"--livereload",
		"--host", "0.0.0.0",
	)
	cmd.Dir = portfolioRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
