// +build windows

package top

import (
	"log"
	"time"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring"

	"github.com/pkg/errors"
)

var counterPath = "\\Process(*)\\% Processor Time"

func (t *Top) startCollect(interval time.Duration) {
	err := monitoring.GetWatcher().StartContinuousQuery(counterPath, interval)
	if err != nil {
		log.Printf("Failed to StartQuery: %s", err)
		return
	}
}

func (t *Top) GetProcesses(interval time.Duration) ([]*ProcessInfo, error) {
	// The non-windows implementation needs 5s to return the results.
	// To keep things in sync the windows implementation should do the same.
	time.Sleep(interval)

	res, err := monitoring.GetWatcher().GetFormattedQueryData(counterPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to call GetFormattedQueryData")
	}

	// Iterate over the reported counter values
	result := make([]*ProcessInfo, 0, len(res))
	for _, c := range res {
		// Some filtering...
		switch c.InstanceName {
		case "Idle":
			continue
		case "_Total":
			continue
		}

		pi := &ProcessInfo{
			Name:    c.InstanceName,
			Command: c.InstanceName,
			Load:    c.Value,
		}
		result = append(result, pi)
	}

	return result, nil
}
