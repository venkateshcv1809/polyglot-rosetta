package toolchain

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"polyglot-rosetta/internal/domain/language"
)

type OSChecker struct {
	Lang language.Language
	Bin  string
	Args []string
}

func NewOSChecker(lang language.Language) (*OSChecker, error) {
	switch lang {
	case language.Go:
		return &OSChecker{Lang: lang, Bin: "go", Args: []string{"version"}}, nil
	case language.Python:
		return &OSChecker{Lang: lang, Bin: "python3", Args: []string{"--version"}}, nil
	case language.Rust:
		return &OSChecker{Lang: lang, Bin: "rustc", Args: []string{"--version"}}, nil
	case language.TypeScript:
		return &OSChecker{Lang: lang, Bin: "tsc", Args: []string{"--version"}}, nil
	case language.Zig:
		return &OSChecker{Lang: lang, Bin: "zig", Args: []string{"version"}}, nil
	default:
		return nil, fmt.Errorf("unsupported language for toolchain check: %s", lang)
	}
}

func (c *OSChecker) Check() (ToolchainInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	path, err := exec.LookPath(c.Bin)
	if err != nil {
		return ToolchainInfo{
			Language:  c.Lang,
			Available: false,
		}, nil
	}

	cmd := exec.CommandContext(ctx, c.Bin, c.Args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return ToolchainInfo{
			Language:  c.Lang,
			Available: false,
			Path:      path,
		}, fmt.Errorf("failed to execute version check: %w", err)
	}

	versionStr := strings.TrimSpace(out.String())
	if versionStr == "" {
		versionStr = strings.TrimSpace(stderr.String())
	}

	return ToolchainInfo{
		Language:  c.Lang,
		Available: true,
		Version:   versionStr,
		Path:      path,
	}, nil
}
