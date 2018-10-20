// +build !windows

package cagent

import (
	"context"
	"errors"
	"strings"
	"time"

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

	if err != nil {
		log.Errorf("[MEM] Failed to get virtual memory stat: %s", err.Error())
		errs = append(errs, err.Error())
		results["total_B"] = nil
		results["used_B"] = nil
		results["free_B"] = nil
		results["shared_B"] = nil
		results["buff_B"] = nil
		results["available_B"] = nil
	} else {
		results["total_B"] = memStat.Total
		results["used_B"] = memStat.Used
		results["free_B"] = memStat.Free
		results["shared_B"] = memStat.Shared
		results["buff_B"] = memStat.Buffers
		results["available_B"] = memStat.Available
	}

	if len(errs) == 0 {
		return results, nil
	}

	return results, errors.New("MEM: " + strings.Join(errs, "; "))
}
