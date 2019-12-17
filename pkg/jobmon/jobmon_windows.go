// +build windows

package jobmon

import (
	"os/exec"
)

func osSpecificCommandConfig(cmd *exec.Cmd) {
}

func osSpecificCommandTermination(cmd *exec.Cmd) {
	_ = cmd.Process.Kill()
}
