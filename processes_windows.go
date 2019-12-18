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
	procByPid, threadsByProcPid, err := winapi.GetSystemProcessInformation(false)
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
	windowByProcessId, err := winapi.WindowByProcessId()
	if err != nil {
		log.Errorf("[PROC] failed to list all windows by processId")
	}

	for pid, proc := range procByPid {
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

		allSuspended := true
		for _, thread := range threadsByProcPid[pid] {
			if thread.ThreadState != winapi.SystemThreadStateWait {
				allSuspended = false
			} else {
				if thread.WaitReason != winapi.SystemThreadWaitReasonSuspended {
					allSuspended = false
				}
			}
		}

		// default state is running
		var state = "running"

		if allSuspended {
			// all threads suspended so mark the process as suspended
			state = "suspended"
		} else if windowByProcessId != nil {
			if window, exists := windowByProcessId[pid]; exists {
				isHanging, err := winapi.IsHangWindow(window)
				if err != nil {
					log.Errorf("[PROC] can't query hang window got error: %s", err.Error())
				} else if isHanging {
					state = "not responding"
				}
			}
		}

		memoryUsagePercent := (float64(proc.WorkingSetSize) / float64(systemMemorySize)) * 100
		ps := ProcStat{
			PID:                    int(pid),
			ParentPID:              int(proc.InheritedFromUniqueProcessID),
			State:                  state,
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
