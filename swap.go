package cagent

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
)

const swapGetTimeout = time.Second * 10

func (ca *Cagent) SwapResults() (MeasurementsMap, error) {
	results := MeasurementsMap{}

	var errs []string
	ctx, cancel := context.WithTimeout(context.Background(), swapGetTimeout)
	defer cancel()

	swapStat, err := mem.SwapMemoryWithContext(ctx)

	if err != nil {
		log.Errorf("[SWAP] Failed to get swap memory stat: %s", err.Error())
		errs = append(errs, err.Error())
	} else if swapStat.Total > 0 {
		results["total_B"] = swapStat.Total
		results["used_B"] = swapStat.Used
		results["free_B"] = swapStat.Free
	}

	if len(errs) == 0 {
		return results, nil
	}

	return results, errors.New("SWAP: " + strings.Join(errs, "; "))
}
