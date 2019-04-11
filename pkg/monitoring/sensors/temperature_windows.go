// +build windows

package sensors

import (
	"math"
	"strings"
	"time"

	"github.com/StackExchange/wmi"

	"github.com/cloudradar-monitoring/cagent/pkg/wmi"
)

const readTimeout = time.Second * 10

var wmiConnectServerArgs = []interface{}{
	nil,        // use localhost
	"root/wmi", // namespace
}

// WMI class
// http://wutils.com/wmi/root/wmi/msacpi_thermalzonetemperature/
type msAcpi_ThermalZoneTemperature struct {
	CriticalTripPoint  uint32
	CurrentTemperature uint32
	InstanceName       string
}

func ReadTemperatureSensors() ([]*TemperatureSensorInfo, error) {
	var thermalSensors []msAcpi_ThermalZoneTemperature
	query := wmi.CreateQuery(&thermalSensors, "")

	err := wmiutil.QueryWithTimeout(readTimeout, query, &thermalSensors, wmiConnectServerArgs...)
	if err != nil {
		l := logger.WithError(err)
		errText := strings.ToLower(err.Error())
		if strings.Contains(errText, "not supported") {
			l.Debugf("not supported by BIOS or driver required")
			return nil, nil
		}
		l.Error("failed to read temperature sensors")
		return nil, err
	}

	result := make([]*TemperatureSensorInfo, 0)
	for _, v := range thermalSensors {
		result = append(result, &TemperatureSensorInfo{
			SensorName:        v.InstanceName,
			Temperature:       wmiTemperatureToCentigrade(v.CurrentTemperature),
			CriticalThreshold: wmiTemperatureToCentigrade(v.CriticalTripPoint),
			Unit:              unitCelsius,
		})
	}
	return result, nil
}

func wmiTemperatureToCentigrade(temp uint32) float64 {
	// WMI returns temperature in Kelvin * 10, so we need to convert it
	t := float64(temp/10) - 273.15
	return math.Trunc((t+0.5/100)*100) / 100
}
