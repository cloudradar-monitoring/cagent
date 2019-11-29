package cagent

import (
	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat/types"
)

func (ca *Cagent) getVMStatMeasurements(f func(string, common.MeasurementsMap, error)) {
	ca.vmstatLazyInit.Do(func() {
		if err := vmstat.Init(); err != nil {
			logrus.Error("vmstat: cannot instantiate virtual machines API: ", err.Error())
			return
		}

		for _, name := range ca.Config.VirtualMachinesStat {
			vm, err := vmstat.Acquire(name)
			if err != nil {
				if err != types.ErrNotAvailable {
					logrus.Warnf("vmstat: Error while acquiring vm provider \"%s\": %s", name, err.Error())
				}
			} else {
				ca.vmWatchers[name] = vm
			}
		}
	})

	for name, p := range ca.vmWatchers {
		res, err := p.GetMeasurements()
		f(name, res, err)
	}
}
