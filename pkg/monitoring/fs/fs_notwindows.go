// +build !windows

package fs

import (
	"github.com/shirou/gopsutil/disk"
)

func getPartitionIOCounters(deviceName string) (*disk.IOCountersStat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fsInfoRequestTimeout)
	defer cancel()
	name := filepath.Base(deviceName)
	result, err := disk.IOCountersWithContext(ctx, name)
	if err != nil {
		return nil, err
	}
	ret := result[name]
	return &ret, nil
}
