// +build windows

package cagent

import (
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/docker"
	"github.com/cloudradar-monitoring/cagent/pkg/winapi"
	"github.com/pkg/errors"
)

func processes(_ *docker.Watcher) ([]ProcStat, error) {
	procs, err := winapi.GetSystemProcessInformation()
	if err != nil {
		return nil, errors.Wrap(err, "[PROC] can't get system processes")
	}

	var result []ProcStat
	for pid, proc := range procs {
		if pid == 0 {
			continue
		}

		ps := ProcStat{
			PID:       int(pid),
			ParentPID: int(proc.InheritedFromUniqueProcessId),
			State:     "running",
			Name:      proc.ImageName.String(),
			Cmdline:   "",
		}

		result = append(result, ps)
	}

	return result, nil
}
