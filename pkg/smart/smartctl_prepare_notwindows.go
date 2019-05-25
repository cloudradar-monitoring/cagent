// +build !windows

package smart

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

func (sm *SMART) smartctlPrepare(disk string) *exec.Cmd {
	var smartctlPrefix string

	// on linux smartctl should be invoked with sudo rights
	if runtime.GOOS == "linux" && common.IsCommandAvailable("sudo") {
		smartctlPrefix = "sudo "
	}

	return exec.Command("/bin/sh", "-c", fmt.Sprintf("%s%s -j -a %s", smartctlPrefix, sm.smartctl, disk))
}
