//go:build !windows

package platform

import (
	"os/exec"
	"syscall"
)

func configureBrowserProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
}
