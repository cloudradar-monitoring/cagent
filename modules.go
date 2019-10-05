package cagent

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/storcli"
)

func (ca *Cagent) collectModulesMeasurements() ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	var errs common.ErrorCollector

	modules := storcli.CreateModules(ca.Config.StorCLI.BinaryPath, ca.Config.StorCLI.ControllerList)

	for _, m := range modules {
		if !m.IsEnabled() {
			continue
		}

		err := m.Run()
		if err != nil {
			err = errors.Wrapf(err, "while executing module '%s'", m.GetName())
			logrus.WithError(err).Debug()
			errs.Add(err)
			continue
		}

		moduleResult := map[string]interface{}{
			"name":             m.GetName(),
			"command executed": m.GetExecutedCommand(),
			"measurements":     m.GetMeasurements(),
		}

		msg := m.GetMessage()
		if len(msg) > 0 {
			moduleResult["message"] = msg
		}

		alerts := m.GetAlerts()
		if alerts == nil {
			alerts = make([]monitoring.Alert, 0)
		}
		moduleResult["alerts"] = alerts

		warnings := m.GetWarnings()
		if warnings == nil {
			warnings = make([]monitoring.Warning, 0)
		}
		moduleResult["warnings"] = warnings

		result = append(result, moduleResult)
	}

	return result, errors.Wrap(errs.Combine(), "while collecting modules measurements")
}
