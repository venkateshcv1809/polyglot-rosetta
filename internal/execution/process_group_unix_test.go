//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package execution

import (
	"context"
	"testing"
	"time"
)

func TestRunCommandTimeoutKillsProcessGroup(t *testing.T) {
	started := time.Now()
	result := RunCommand(context.Background(), 50*time.Millisecond, helperCommand("spawn"))
	if !result.TimedOut {
		t.Fatalf("TimedOut = false, error = %v", result.Error)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("process group cancellation took %v", elapsed)
	}
}
