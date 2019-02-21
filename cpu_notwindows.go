// +build !windows

package cagent

import (
	"context"
	"github.com/shirou/gopsutil/cpu"
	"time"
)

const cpuGetUtilisationTimeout = time.Second * 10

func getCPUTimes() ([]cpu.TimesStat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cpuGetUtilisationTimeout)
	defer cancel()

	times, err := cpu.TimesWithContext(ctx, true)
	if err == context.DeadlineExceeded {
		return times, TimeoutError{"gopsutil cpu.Times call", cpuGetUtilisationTimeout}
	}
	return times, err
}
