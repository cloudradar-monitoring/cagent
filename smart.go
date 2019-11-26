package cagent

import (
	"strings"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

func (ca *Cagent) getSMARTMeasurements() common.MeasurementsMap {
	if ca.smart != nil {
		res, errs := ca.smart.Parse()

		if len(errs) > 0 {
			var errStr []string
			for _, e := range errs {
				errStr = append(errStr, e.Error())
			}

			if res == nil {
				res = make(common.MeasurementsMap)
			}

			res["messages"] = strings.Join(errStr, "; ")
		}

		return res
	}

	return nil
}
