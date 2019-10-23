package cagent

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/storcli"
)

var modules []monitoring.Module

func (ca *Cagent) collectModulesMeasurements() ([]*monitoring.ModuleReport, error) {
	var result []*monitoring.ModuleReport
	var errs common.ErrorCollector

	if len(modules) == 0 {
		m := storcli.CreateModule(ca.Config.StorCLI.BinaryPath)
		if m.IsEnabled() {
			modules = append(modules, m)
		}
	}

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
