// +build windows

package monitoring

import (
	"github.com/cloudradar-monitoring/cagent/perfcounters"
)

var watcher *perfcounters.WinPerfCountersWatcher

func init() {
	watcher = perfcounters.Watcher()
}

func GetWatcher() *perfcounters.WinPerfCountersWatcher {
	return watcher
}
