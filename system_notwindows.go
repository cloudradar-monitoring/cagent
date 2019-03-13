// +build !windows

package cagent

import (
	"context"
	"errors"

	"github.com/shirou/gopsutil/cpu"
)

func getProcessorModelName() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cpuInfoTimeout)
	defer cancel()

	cpuInfos, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return "", err
	}

	if len(cpuInfos) == 0 {
		return "", errors.New("no information about CPU")
	}

	return cpuInfos[0].ModelName, nil
}
