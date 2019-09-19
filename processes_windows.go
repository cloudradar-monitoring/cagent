// +build windows

package cagent

import (
	"runtime"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/winapi"
)

var monitoredProcessCache = make(map[uint32]*winapi.SystemProcessInformation)
var lastProcessQueryTime time.Time

func processes(systemMemorySize uint64) ([]ProcStat, error) {
	procs, err := winapi.GetSystemProcessInformation()
	if err != nil {
		return nil, errors.Wrap(err, "[PROC] can't get system processes")
	}

	now := time.Now()
	timeElapsedReal := 0.0
	if !lastProcessQueryTime.IsZero() {
		timeElapsedReal = now.Sub(lastProcessQueryTime).Seconds()
	}

	var result []ProcStat
	var updatedProcessCache = make(map[uint32]*winapi.SystemProcessInformation)
	cmdLineRetrievalFailuresCount := 0
	logicalCPUCount := uint8(runtime.NumCPU())
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

		oldProcessInfo, oldProcessInfoExists := monitoredProcessCache[pid]
		cpuUsagePercent := 0.0
		if oldProcessInfoExists && timeElapsedReal > 0 {
			cpuUsagePercent = winapi.CalculateProcessCPUUsagePercent(oldProcessInfo, proc, timeElapsedReal, logicalCPUCount)
		}

		memoryUsagePercent := (float64(proc.WorkingSetSize) / float64(systemMemorySize)) * 100
		ps := ProcStat{
			PID:                    int(pid),
			ParentPID:              int(proc.InheritedFromUniqueProcessID),
			State:                  "running",
			Name:                   proc.ImageName.String(),
			Cmdline:                cmdLine,
			CPUAverageUsagePercent: float32(common.RoundToTwoDecimalPlaces(cpuUsagePercent)),
			RSS:                    uint64(proc.WorkingSetPrivateSize),
			VMS:                    uint64(proc.VirtualSize),
			MemoryUsagePercent:     float32(common.RoundToTwoDecimalPlaces(memoryUsagePercent)),
		}

		updatedProcessCache[pid] = proc
		result = append(result, ps)
	}
	lastProcessQueryTime = now
	monitoredProcessCache = updatedProcessCache

	if cmdLineRetrievalFailuresCount > 0 {
		log.Debugf("[PROC] could not get command line for %d processes", cmdLineRetrievalFailuresCount)
	}

	return result, nil
}
