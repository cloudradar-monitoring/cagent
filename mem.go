// +build !windows

package cagent

import (
	"context"
	"errors"
	"strings"
	"time"
	"runtime"

	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
)

const memGetTimeout = time.Second * 10

func (ca *Cagent) MemResults() (MeasurementsMap, error) {
	results := MeasurementsMap{}

	var errs []string
	ctx, cancel := context.WithTimeout(context.Background(), memGetTimeout)
	defer cancel()

	memStat, err := mem.VirtualMemoryWithContext(ctx)

	results = map[string]interface{}{
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

	if err != nil {
		log.Errorf("[MEM] Failed to get virtual memory stat: %s", err.Error())
		errs = append(errs, err.Error())
	} else {
		results["total_B"] = memStat.Total
		results["used_B"] = memStat.Used
		results["used_percent"] = floatToIntPercent(float64(memStat.Used) / float64(memStat.Total))
		results["free_B"] = memStat.Free
		results["free_percent"] = floatToIntPercent(float64(memStat.Free) / float64(memStat.Total))
		results["shared_B"] = memStat.Shared
		results["shared_percent"] = floatToIntPercent(float64(memStat.Shared) / float64(memStat.Total))
		results["cached_B"] = memStat.Cached
		results["cached_percent"] = floatToIntPercent(float64(memStat.Cached) / float64(memStat.Total))
		results["shared_percent"] = floatToIntPercent(float64(memStat.Shared) / float64(memStat.Total))
		results["buff_B"] = memStat.Buffers
		results["buff_percent"] = floatToIntPercent(float64(memStat.Buffers) / float64(memStat.Total))

		var hasAvailableMememory bool
		// linux has native MemAvailable metric since 3.14 kernel
		// others calculated within github.com/shirou/gopsutil
		switch runtime.GOOS {
		case "linux", "freebsd", "openbsd", "darwin":
			hasAvailableMetric = true
		default:
			hasAvailableMememory = false
		}

		if memStat != nil && hasAvailableMememory {
			results["available_B"] = int(memStat.Available)
			results["available_percent"] = floatToIntPercent(float64(results["available_B"].(int)) / float64(memStat.Total))
		} else {
			results["available_B"] = nil
		}
	}

	if len(errs) == 0 {
		return results, nil
	}

	return results, errors.New("MEM: " + strings.Join(errs, "; "))
}
