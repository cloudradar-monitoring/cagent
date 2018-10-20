// +build windows

package win_perf_counters

import (
	"fmt"
	"strings"
)

type WinPerfCountersWatcher struct {
	query PerformanceQuery
}

func (m *WinPerfCountersWatcher) Query(counterPath string, instance string) (float64, error) {
	var counterHandle PDH_HCOUNTER
	var err error

	if err = m.query.Open(); err != nil {
		return 0, err
	}

	defer m.query.Close()

	if !m.query.IsVistaOrNewer() {
		counterHandle, err = m.query.AddCounterToQuery(counterPath)
		if err != nil {
			return 0, err
		}
	} else {
		counterHandle, err = m.query.AddEnglishCounterToQuery(counterPath)
		if err != nil {
			return 0, err
		}
	}

	if err = m.query.CollectData(); err != nil {
		return 0, err
	}

	// For iterate over the known metrics and get the samples.
	// collect
	counterValues, err := m.query.GetFormattedCounterArrayDouble(counterHandle)
	if err == nil {
		for _, cValue := range counterValues {
			var ok bool
			if instance == "*" && !strings.Contains(cValue.InstanceName, "_Total") {
				// Catch if set to * and that it is not a '*_Total*' instance.
				ok = true
			} else if instance == "total" && strings.Contains(cValue.InstanceName, "_Total") {
				// Catch if set to * and that it is not a '*_Total*' instance.
				ok = true
			} else if instance == cValue.InstanceName {
				// Catch if we set it to total or some form of it
				ok = true
			} else if strings.Contains(instance, "#") && strings.HasPrefix(instance, cValue.InstanceName) {
				// If you are using a multiple instance identifier such as "w3wp#1"
				// phd.dll returns only the first 2 characters of the identifier.
				ok = true
			}

			if ok {
				return cValue.Value, nil
			} else {
				continue
			}
		}
		return 0, fmt.Errorf("error while getting value for counter %s: can't find value for instance %s", counterPath, instance)
	} else {
		//ignore invalid data as some counters from process instances returns this sometimes
		if isKnownCounterDataError(err) {
			return 0, nil
		} else {
			return 0, fmt.Errorf("error while getting value for counter %s: %v", counterPath, err)
		}
	}

}

func isKnownCounterDataError(err error) bool {
	if pdhErr, ok := err.(*PdhError); ok && (pdhErr.ErrorCode == PDH_INVALID_DATA ||
		pdhErr.ErrorCode == PDH_CALC_NEGATIVE_VALUE ||
		pdhErr.ErrorCode == PDH_CSTATUS_INVALID_DATA) {
		return true
	}
	return false
}

func Watcher() *WinPerfCountersWatcher {
	return &WinPerfCountersWatcher{query: &PerformanceQueryImpl{}}
}
