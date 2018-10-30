// +build windows

package top

import (
	"log"
	"time"

	"github.com/cloudradar-monitoring/cagent/perfcounters"
	"github.com/pkg/errors"
)

var watcher *perfcounters.WinPerfCountersWatcher
var counterPath = "\\Process(*)\\% Processor Time"

func (t *Top) startCollect(interval time.Duration) {
	watcher = perfcounters.Watcher()

	err := watcher.StartQuery(counterPath, interval)
	if err != nil {
		log.Printf("Failed to StartQuery: %s", err)
		return
	}
}

func (t *Top) GetProcesses() ([]*ProcessInfo, error) {
	res, err := watcher.GetFormattedQueryData(counterPath)
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

		// Only pay attention to processes that do something
		// if c.Value != 0 {
		pi := &ProcessInfo{
			Identifier: c.InstanceName,
			Command:    c.InstanceName,
			Load:       c.Value,
		}
		result = append(result, pi)
		// }
	}

	return result, nil
}
