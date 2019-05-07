// +build windows

package smart

import (
	"os/exec"
)

func (sm *SMART) smartctlPrepare(disk string) *exec.Cmd {
	return exec.Command("cmd", "/c", sm.smartctl, "-j", "-a", disk)
}
