// +build !windows

package smart

import (
	"os/exec"
	"runtime"
)

func smartctlPrepare(disk string) *exec.Cmd {
	smartctlPrefix := "sudo smartctl"
	if runtime.GOOS == "darwin" {
		// darwin does not need smartctl to run with sudo rights
		smartctlPrefix = "smartctl"
	}

	return exec.Command("/bin/sh", "-c", smartctlPrefix+" -j -a "+disk)
}
