package raid

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring"
)

func helperInitModule(fileName string) *RAID {
	path := filepath.Join("testdata", fileName)
	return &RAID{
		mdstatFilePath: path,
	}
}

func TestRAIDModule(t *testing.T) {
	type expectedValues struct {
		moduleEnabled bool
		alerts        []monitoring.Alert
		warnings      []monitoring.Warning
	}

	var noAlerts = make([]monitoring.Alert, 0)
	var noWarnings = make([]monitoring.Warning, 0)
	const alertNonOptimal = "Raid status not optimal (Needs Attention)"

	var testMap = map[string]expectedValues{
		"mdstat_empty":          {false, noAlerts, noWarnings},
		"mdstat_not_configured": {false, noAlerts, noWarnings},

		"mdstat_degraded_fail":          {true, []monitoring.Alert{"Raid md1 degraded. Devices failing: sde1.", "Raid md1 degraded. Missing 1 devices.", alertNonOptimal}, noWarnings},
		"mdstat_degraded_phys_missing1": {true, []monitoring.Alert{"Raid md0 degraded. Missing 1 devices.", alertNonOptimal}, noWarnings},
		"mdstat_degraded_phys_missing2": {true, []monitoring.Alert{"Raid md2 degraded. Missing 1 devices.", alertNonOptimal}, noWarnings},

		"mdstat_good1":        {true, noAlerts, noWarnings},
		"mdstat_good2":        {true, noAlerts, noWarnings},
		"mdstat_good3_bitmap": {true, noAlerts, noWarnings},

		"mdstat_recovery": {true, noAlerts, []monitoring.Warning{"Raid md127 rebuilding."}},
	}

	for fileName, expected := range testMap {
		t.Run(fmt.Sprintf("test-%s", fileName), func(t *testing.T) {
			m := helperInitModule(fileName)
			assert.Equal(t, expected.moduleEnabled, m.IsEnabled())

			if expected.moduleEnabled {
				reports, err := m.Run()
				assert.NoError(t, err)
				assert.Len(t, reports, 1)

				r := reports[0]
				assert.EqualValues(t, expected.alerts, r.Alerts)
				assert.EqualValues(t, expected.warnings, r.Warnings)
			}
		})
	}
}
