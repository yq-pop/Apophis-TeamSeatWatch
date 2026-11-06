//go:build windows

package platform

import "os/exec"

// Windows has no POSIX process-group or parent-death semantics; the profile
// barrier still closes the browser through CDP before this process is released.
func configureBrowserProcess(*exec.Cmd) {}
