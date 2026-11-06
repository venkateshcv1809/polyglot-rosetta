package sitecmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"polyglot-rosetta/internal/config"
)

var clonePortfolio = clonePortfolioRepo

func resolvePortfolioRoot(explicit string) (string, error) {
	if root := strings.TrimSpace(explicit); root != "" {
		return filepath.Abs(root)
	}
	if root := strings.TrimSpace(os.Getenv("PORTFOLIO_ROOT")); root != "" {
		return filepath.Abs(root)
	}

	return cloneFallback()
}

func cloneFallback() (string, error) {
	cloneRoot := filepath.Join(os.TempDir(), config.PortfolioCloneDir)
	info, err := os.Stat(cloneRoot)
	if err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("temporary portfolio path %s is not a directory", cloneRoot)
		}
		remote, err := exec.Command("git", "-C", cloneRoot, "remote", "get-url", "origin").Output()
		if err != nil || strings.TrimSpace(string(remote)) != config.PortfolioRepositoryURL {
			return "", fmt.Errorf("temporary portfolio path %s is not a clone of %s; set --source or PORTFOLIO_ROOT, or remove that path", cloneRoot, config.PortfolioRepositoryURL)
		}
		return filepath.Abs(cloneRoot)
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect temporary portfolio path %s: %w", cloneRoot, err)
	}
	if err := clonePortfolio(cloneRoot); err != nil {
		return "", fmt.Errorf("clone portfolio repository: %w", err)
	}
	return filepath.Abs(cloneRoot)
}

func clonePortfolioRepo(dest string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", config.PortfolioRepositoryURL, dest)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
