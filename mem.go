// +build !windows

package cagent

import (
	"context"
	"errors"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

const memGetTimeout = time.Second * 10

func (ca *Cagent) MemResults() (common.MeasurementsMap, *mem.VirtualMemoryStat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), memGetTimeout)
	defer cancel()

	results := common.MeasurementsMap{
		"total_B":           nil,
		"free_B":            nil,
		"free_percent":      nil,
		"cached_B":          nil,
		"cached_percent":    nil,
		"shared_B":          nil,
		"shared_percent":    nil,
		"buff_B":            nil,
		"buff_percent":      nil,
		"used_B":            nil,
		"used_percent":      nil,
		"available_B":       nil,
		"available_percent": nil,
	}

	memStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		log.Errorf("[MEM] Failed to get virtual memory stat: %s", err.Error())
		return results, memStat, errors.New("MEM: " + err.Error())
	}

	results["total_B"] = memStat.Total
	results["used_B"] = memStat.Used
	results["used_percent"] = floatToIntPercentRoundUP(float64(memStat.Used) / float64(memStat.Total))
	results["free_B"] = memStat.Free
	results["free_percent"] = floatToIntPercentRoundUP(float64(memStat.Free) / float64(memStat.Total))
	results["shared_B"] = memStat.Shared
	results["shared_percent"] = floatToIntPercentRoundUP(float64(memStat.Shared) / float64(memStat.Total))
	results["cached_B"] = memStat.Cached
	results["cached_percent"] = floatToIntPercentRoundUP(float64(memStat.Cached) / float64(memStat.Total))
	results["shared_percent"] = floatToIntPercentRoundUP(float64(memStat.Shared) / float64(memStat.Total))
	results["buff_B"] = memStat.Buffers
	results["buff_percent"] = floatToIntPercentRoundUP(float64(memStat.Buffers) / float64(memStat.Total))

	var hasAvailableMemory bool
	// linux has native MemAvailable metric since 3.14 kernel
	// others calculated within github.com/shirou/gopsutil
	switch runtime.GOOS {
	case "linux", "freebsd", "openbsd", "darwin":
		hasAvailableMemory = true
	default:
		hasAvailableMemory = false
	}

	if memStat != nil && hasAvailableMemory {
		results["available_B"] = int(memStat.Available)
		results["available_percent"] = floatToIntPercentRoundUP(float64(results["available_B"].(int)) / float64(memStat.Total))
	}

	return results, memStat, nil
}
