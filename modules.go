package cagent

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/mysql"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/raid"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/storcli"
)

var modules []monitoring.Module

func (ca *Cagent) initModules() {
	if len(modules) > 0 {
		return
	}

	l := []func() monitoring.Module{
		func() monitoring.Module {
			return storcli.CreateModule(ca.Config.StorCLI.BinaryPath)
		},
		func() monitoring.Module {
			return raid.CreateModule(ca.Config.SoftwareRAIDMonitoring)
		},
		func() monitoring.Module {
			return mysql.CreateModule(&ca.Config.MysqlMonitoring)
		},
	}

	for _, f := range l {
		m := f()
		if m.IsEnabled() {
			modules = append(modules, m)
		}
	}
}

func (ca *Cagent) collectModulesMeasurements() ([]*monitoring.ModuleReport, error) {
	var result []*monitoring.ModuleReport
	var errs common.ErrorCollector

	ca.initModules()

	for _, m := range modules {
		reports, err := m.Run()
		if err != nil {
			err = errors.Wrapf(err, "while executing module '%s'", m.GetDescription())
			logrus.WithError(err).Debug()
			errs.Add(err)
			continue
		}
		result = append(result, reports...)
	}

	return result, errors.Wrap(errs.Combine(), "while collecting modules measurements")
}
