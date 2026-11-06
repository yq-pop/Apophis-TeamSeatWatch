//go:build windows

package browserenv

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

func openProfileLock(path string) (*os.File, error) {
	// Windows has no O_NOFOLLOW; the profile directory itself is validated.
	return os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
}

func lockProfile(file *os.File) error {
	var overlapped windows.Overlapped
	return windows.LockFileEx(windows.Handle(file.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlapped)
}

func unlockProfile(file *os.File) error {
	var overlapped windows.Overlapped
	return windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 1, 0, &overlapped)
}

func profileLockBusy(err error) bool {
	return errors.Is(err, windows.ERROR_LOCK_VIOLATION)
}

// Process discovery relies on /proc, which does not exist on Windows; there is
// no portable marker scan, so no process is killed here.
func killProfileProcess(int) {}
