// +build !windows

package cagent

type windowsUpdateWatcher struct{}

func (ca *Cagent) WindowsUpdatesWatcher() *windowsUpdateWatcher {
	return &windowsUpdateWatcher{}
}

func (ca *windowsUpdateWatcher) WindowsUpdates() (MeasurementsMap, error) {
	return MeasurementsMap{}, nil
}
