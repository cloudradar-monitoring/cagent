// +build !windows

package cagent

type WindowsUpdateWatcher struct{}

func (ca *Cagent) WindowsUpdatesWatcher() *WindowsUpdateWatcher {
	return &WindowsUpdateWatcher{}
}

func (ca *WindowsUpdateWatcher) WindowsUpdates() (MeasurementsMap, error) {
	return MeasurementsMap{}, nil
}
