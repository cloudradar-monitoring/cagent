// +build windows

package cagent

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring"
)

const memGetTimeout = time.Second * 10

func (ca *Cagent) MemResults() (MeasurementsMap, error) {
	results := MeasurementsMap{}

	var errs []string
	ctx, cancel := context.WithTimeout(context.Background(), memGetTimeout)
	defer cancel()

	results = map[string]interface{}{
		"total_B":           nil,
		"free_B":            nil,
		"free_percent":      nil,
		"cached_B":          nil,
		"cached_percent":    nil,
		"used_B":            nil,
		"used_percent":      nil,
		"available_B":       nil,
		"available_percent": nil,
	}

	memStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		errs = append(errs, err.Error())
		log.Errorf("[MEM] Failed to get virtual memory stat: %s", err.Error())
	} else {
		results["total_B"] = memStat.Total
		results["available_B"] = memStat.Available
		results["available_percent"] = floatToIntPercentRoundUP((float64(memStat.Available)) / float64(memStat.Total))
	}

	free, err := monitoring.GetWatcher().Query(`\Memory\Free & Zero Page List Bytes`, "*")
	if err != nil {
		errs = append(errs, err.Error())
		log.Errorf("[MEM] Failed to get free memory: %s", err.Error())
	} else {
		results["used_B"] = int(memStat.Total) - int(free)
		results["used_percent"] = floatToIntPercentRoundUP((float64(memStat.Total) - free) / float64(memStat.Total))
		results["free_B"] = int(free)
		results["free_percent"] = floatToIntPercentRoundUP(free / float64(memStat.Total))
	}

	cached, err := monitoring.GetWatcher().Query(`\Memory\Cache Bytes`, "*")
	if err != nil {
		errs = append(errs, err.Error())
		log.Errorf("[MEM] Failed to get cached memory: %s", err.Error())
	} else {
		results["cached_B"] = int(cached)
		results["cached_percent"] = floatToIntPercentRoundUP(float64(cached) / float64(memStat.Total))
	}

	if len(errs) == 0 {
		return results, nil
	}

	return results, errors.New("MEM: " + strings.Join(errs, "; "))
}
