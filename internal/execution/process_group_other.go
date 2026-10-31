//go:build !(aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris)

package execution

import (
	"os/exec"
)

func configureProcessGroup(cmd *exec.Cmd) {
	cmd.WaitDelay = defaultProcessWaitDelay
}
