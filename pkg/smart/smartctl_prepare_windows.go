// +build windows

package smart

import (
	"os/exec"
)

func smartctlPrepare(disk string) *exec.Cmd {
	return exec.Command("smartctl.exe", "-j", "-a", disk)
}
