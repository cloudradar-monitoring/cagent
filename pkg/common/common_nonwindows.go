// +build !windows

package common

import (
	"os/exec"
)

func IsCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command", "-v", name)
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func WrapCommandToAdmin(args ...string) string {
	var cmd string
	if IsCommandAvailable("sudo") {
		// expecting 'sudo' package is installed and /etc/sudoers.d/cagent-<cmd> is present
		// run sudo in non-interactive mode to prevent prompt for password
		cmd = "sudo -n"
	}

	for _, s := range args {
		if cmd != "" {
			cmd += " "
		}

		cmd += s
	}

	return cmd
}
