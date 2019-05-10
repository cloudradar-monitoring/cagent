// +build !windows

package smart

import (
	"fmt"
	"os/exec"
	"runtime"
)

func (sm *SMART) smartctlPrepare(disk string) *exec.Cmd {
	var smartctlPrefix string

	// on linux smartctl should be invoked with sudo rights
	if runtime.GOOS == "linux" {
		smartctlPrefix = "sudo "
	}

	return exec.Command("/bin/sh", "-c", fmt.Sprintf("%s%s -j -a %s", smartctlPrefix, sm.smartctl, disk))
}
