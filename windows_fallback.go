// +build !windows

package cagent

import (
	"github.com/cloudradar-monitoring/cagent/types"
)

type WindowsUpdateWatcher struct{}

func (ca *Cagent) WindowsUpdatesWatcher() *WindowsUpdateWatcher {
	return &WindowsUpdateWatcher{}
}

func (ca *WindowsUpdateWatcher) WindowsUpdates() (types.MeasurementsMap, error) {
	return types.MeasurementsMap{}, nil
}
