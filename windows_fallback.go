// +build !windows

package cagent

func (ca *Cagent) WindowsUpdates() (MeasurementsMap, error){
	return MeasurementsMap{}, nil
}