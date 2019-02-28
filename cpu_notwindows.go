// +build !windows

package cagent

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/cpu"
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
