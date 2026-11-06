//go:build !windows

package browserenv

import (
	"errors"
	"os"
	"syscall"
)

func openProfileLock(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
}

func lockProfile(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func unlockProfile(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}

func profileLockBusy(err error) bool {
	return errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN)
}

func killProfileProcess(pid int) {
	if group, err := syscall.Getpgid(pid); err == nil && group == pid {
		_ = syscall.Kill(-group, syscall.SIGKILL)
	}
	_ = syscall.Kill(pid, syscall.SIGKILL)
}
