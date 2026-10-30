package toolchain

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const versionTimeout = 3 * time.Second

// InspectBinary reads the first version line reported by an installed tool.
// A timeout keeps an unresponsive executable from blocking the probe command.
func InspectBinary(binary, path string) (string, error) {
	return inspectBinary(binary, path, versionTimeout)
}

func inspectBinary(binary, path string, timeout time.Duration) (string, error) {
	args := []string{"--version"}
	if binary == "go" || binary == "zig" {
		args = []string{"version"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	command := exec.CommandContext(ctx, path, args...)
	command.WaitDelay = 250 * time.Millisecond
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		return "", fmt.Errorf("run %s: %w", binary, ctx.Err())
	}
	if err != nil {
		return "", fmt.Errorf("run %s: %w", binary, err)
	}
	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", fmt.Errorf("%s returned an empty version", binary)
	}
	return strings.Split(version, "\n")[0], nil
}
