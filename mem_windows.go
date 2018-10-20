// +build windows

package cagent

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
	"github.com/cloudradar-monitoring/cagent/win_perf_counters"
)

const memGetTimeout = time.Second * 10
var watcher = win_perf_counters.Watcher()

func (ca *Cagent) MemResults() (MeasurementsMap, error) {
	results := MeasurementsMap{}

	var errs []string
	ctx, cancel := context.WithTimeout(context.Background(), memGetTimeout)
	defer cancel()

	memStat, err := mem.VirtualMemoryWithContext(ctx)

	results["free_B"] = nil
	results["shared_B"] = nil
	results["buff_B"] = nil

	if err != nil {
		log.Errorf("[MEM] Failed to get virtual memory stat: %s", err.Error())
		errs = append(errs, err.Error())
		results["total_B"] = nil
		results["used_B"] = nil
		results["available_B"] = nil
	} else {
		results["total_B"] = memStat.Total
		results["used_B"] = memStat.Used
		results["available_B"] = memStat.Available
	}

	free, err := watcher.Query(`\Memory\Free & Zero Page List Bytes`, "*")

	if err != nil {
		log.Errorf("[MEM] Failed to get free memory: %s", err.Error())
	} else {
		results["free_B"] = free
	}

	if len(errs) == 0 {
		return results, nil
	}

	return results, errors.New("MEM: " + strings.Join(errs, "; "))
}
