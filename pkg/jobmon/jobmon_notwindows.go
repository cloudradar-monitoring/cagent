// +build !windows

package jobmon

import (
	"os/exec"
	"syscall"
)

func osSpecificCommandConfig(cmd *exec.Cmd) {
	// create a job in a different process group
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}
}

func osSpecificCommandTermination(cmd *exec.Cmd) {
	processGroupID, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		// fallback to default kill
		_ = cmd.Process.Kill()
		return
	}

	_ = syscall.Kill(-processGroupID, syscall.SIGTERM)
}
