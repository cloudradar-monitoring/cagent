// +build windows

package top

import (
	"runtime"
	"time"

	"github.com/cloudradar-monitoring/cagent/pkg/winapi"
)

func (t *Top) GetProcesses(interval time.Duration) ([]*ProcessInfoSnapshot, error) {
	processesOld, _, err := winapi.GetSystemProcessInformation(true)
	if err != nil {
		return nil, err
	}
	startTime := time.Now()

	time.Sleep(interval)

	processesNew, _, err := winapi.GetSystemProcessInformation(true)
	if err != nil {
		return nil, err
	}
	finishTime := time.Now()

	timeElapsedReal := finishTime.Sub(startTime).Seconds()

	result := make([]*ProcessInfoSnapshot, 0, len(processesOld))
	for pid, newProcessInfo := range processesNew {
		if oldProcessInfo, exists := processesOld[pid]; exists {
			if pid == 0 {
				continue
			}
			info := &ProcessInfoSnapshot{
				Name:      newProcessInfo.ImageName.String(),
				PID:       pid,
				ParentPID: uint32(newProcessInfo.InheritedFromUniqueProcessID),
				Command:   "",
				Load:      winapi.CalculateProcessCPUUsagePercent(oldProcessInfo, newProcessInfo, timeElapsedReal, t.logicalCPUCount),
			}
			result = append(result, info)
		}
	}

	// calling to winapi functions can consume a lot of memory, so we trigger GC to clean things up
	runtime.GC()

	return result, nil
}
