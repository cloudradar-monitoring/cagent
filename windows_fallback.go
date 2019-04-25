// +build !windows

package cagent

import "github.com/cloudradar-monitoring/cagent/pkg/common"

type WindowsUpdateWatcher struct{}

func (ca *Cagent) WindowsUpdatesWatcher() *WindowsUpdateWatcher {
	return &WindowsUpdateWatcher{}
}

func (ca *WindowsUpdateWatcher) WindowsUpdates() (common.MeasurementsMap, error) {
	return common.MeasurementsMap{}, nil
}
