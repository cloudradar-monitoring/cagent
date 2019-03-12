// +build windows

package cagent

import (
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/docker"
	"github.com/cloudradar-monitoring/cagent/pkg/winapi"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func processes(_ *docker.Watcher) ([]ProcStat, error) {
	procs, err := winapi.GetSystemProcessInformation()
	if err != nil {
		return nil, errors.Wrap(err, "[PROC] can't get system processes")
	}

	var result []ProcStat
	cmdLineRetrievalFailuresCount := 0
	for pid, proc := range procs {
		if pid == 0 {
			continue
		}

		cmdLine, err := winapi.GetProcessCommandLine(pid)
		if err != nil {
			// there are some edge-cases when we can't get cmdLine in reliable way.
			// it includes system processes, which are not accessible in user-mode and processes from outside of WOW64 when running as a 32-bit process
			cmdLineRetrievalFailuresCount++
		}

		ps := ProcStat{
			PID:       int(pid),
			ParentPID: int(proc.InheritedFromUniqueProcessId),
			State:     "running",
			Name:      proc.ImageName.String(),
			Cmdline:   cmdLine,
		}

		result = append(result, ps)
	}

	if cmdLineRetrievalFailuresCount > 0 {
		log.Debugf("[PROC] could not get command line for %d processes", cmdLineRetrievalFailuresCount)
	}

	return result, nil
}
