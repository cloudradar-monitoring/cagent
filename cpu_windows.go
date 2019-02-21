// +build windows

package cagent

import (
	"fmt"
	"github.com/cloudradar-monitoring/cagent/pkg/winapi"
	"github.com/shirou/gopsutil/cpu"
)

func systemProcessorPerformanceInfoToCPUTimesStat(core int, s *winapi.SystemProcessorPerformanceInformation) cpu.TimesStat {
	return cpu.TimesStat{
		CPU:    fmt.Sprintf("cpu%d", core), // making consistent with other gopsutil/cpu implementations
		Idle:   float64(s.IdleTime) * winapi.HundredNSToTick,
		System: float64(s.KernelTime-s.IdleTime) * winapi.HundredNSToTick,
		User:   float64(s.UserTime) * winapi.HundredNSToTick,
		Irq:    float64(s.InterruptTime) * winapi.HundredNSToTick,
	}
}

func getCPUTimes() ([]cpu.TimesStat, error) {
	var coreInfo []cpu.TimesStat

	results, err := winapi.GetSystemProcessorPerformanceInformation()
	if err != nil {
		return coreInfo, err
	}

	coreInfo = make([]cpu.TimesStat, 0, len(results))
	for core, info := range results {
		coreInfo = append(coreInfo, systemProcessorPerformanceInfoToCPUTimesStat(core, &info))
	}

	return coreInfo, nil
}
