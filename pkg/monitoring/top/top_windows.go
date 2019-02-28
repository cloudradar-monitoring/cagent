// +build windows

package top

import (
	"github.com/cloudradar-monitoring/cagent/pkg/winapi"
	"runtime"
	"time"
)

func (t *Top) GetProcesses(interval time.Duration) ([]*ProcessInfo, error) {
	processesOld, err := winapi.GetSystemProcessInformation()
	if err != nil {
		return nil, err
	}
	startTime := time.Now()

	time.Sleep(interval)

	processesNew, err := winapi.GetSystemProcessInformation()
	if err != nil {
		return nil, err
	}
	finishTime := time.Now()

	timeElapsedReal := (finishTime.Sub(startTime).Seconds()) * float64(t.logicalCPUCount)

	result := make([]*ProcessInfo, 0, len(processesOld))
	for pid, newProcessInfo := range processesNew {
		if oldProcessInfo, exists := processesOld[pid]; exists {
			if pid == 0 {
				continue
			}
			info := &ProcessInfo{
				Name:    newProcessInfo.ImageName.String(),
				PID:     pid,
				Command: "",
				Load:    calculateUsagePercent(oldProcessInfo, newProcessInfo, timeElapsedReal),
			}
			result = append(result, info)
		}
	}

	// calling to winapi functions can consume a lot of memory, so we trigger GC to clean things up
	runtime.GC()

	return result, nil
}

func calculateUsagePercent(p1, p2 *winapi.SystemProcessInformation, delta float64) float64 {
	if delta == 0 {
		return 0
	}

	deltaProc := float64((p2.UserTime+p2.KernelTime)-(p1.UserTime+p1.KernelTime)) * winapi.HundredNSToTick
	overallPercent := (deltaProc / delta) * 100
	return overallPercent
}
