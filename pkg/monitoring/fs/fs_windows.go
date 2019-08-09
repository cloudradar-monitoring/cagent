// +build windows

package fs

import (
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/disk"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/registry"

	"github.com/cloudradar-monitoring/cagent/pkg/winapi"
)

func getPartitionIOCounters(deviceName string) (*disk.IOCountersStat, error) {
	if err := enablePerformanceCounters(); err != nil {
		return nil, err
	}

	var uncPath = `\\.\` + deviceName
	diskPerformance, err := winapi.GetDiskPerformance(uncPath)
	if err != nil {
		return nil, err
	}
	return &disk.IOCountersStat{
		Name:       deviceName,
		ReadCount:  uint64(diskPerformance.ReadCount),
		WriteCount: uint64(diskPerformance.WriteCount),
		ReadBytes:  uint64(diskPerformance.BytesRead),
		WriteBytes: uint64(diskPerformance.BytesWritten),
		ReadTime:   uint64(diskPerformance.ReadTime),
		WriteTime:  uint64(diskPerformance.WriteTime),
	}, nil
}

// enablePerformanceCounters will enable performance counters by adding the EnableCounterForIoctl registry key
func enablePerformanceCounters() error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, "SYSTEM\\CurrentControlSet\\Services\\partmgr", registry.READ|registry.WRITE)
	if err != nil {
		return errors.Errorf("cannot open new key in the registry in order to enable the performance counters: %s", err)
	}
	val, _, err := key.GetIntegerValue("EnableCounterForIoctl")
	if val != 1 || err != nil {
		if err = key.SetDWordValue("EnableCounterForIoctl", 1); err != nil {
			return errors.Errorf("cannot create HKLM:SYSTEM\\CurrentControlSet\\Services\\Partmgr\\EnableCounterForIoctl key in the registry in order to enable the performance counters: %s", err)
		}
		logrus.Info("The registry key EnableCounterForIoctl at HKLM:SYSTEM\\CurrentControlSet\\Services\\Partmgr has been created in order to enable the performance counters")
	}
	return nil
}
