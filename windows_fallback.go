// +build !windows

package cagent

type EmptyWindowsUpdateWatcher struct{}

func (ca *Cagent) WindowsUpdatesWatcher() EmptyWindowsUpdateWatcher {
	return EmptyWindowsUpdateWatcher{}
}

func (ca *EmptyWindowsUpdateWatcher) WindowsUpdates() (MeasurementsMap, error) {
	return MeasurementsMap{}, nil
}
