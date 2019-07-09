// +build windows

package sensors

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/StackExchange/wmi"

	"github.com/cloudradar-monitoring/cagent/pkg/hwinfolib"
	wmiutil "github.com/cloudradar-monitoring/cagent/pkg/wmi"
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

// ReadTemperatureSensors tries to read sensor info from hwinfo.dll if it's available. Fallbacks to WMI class msAcpi_ThermalZoneTemperature
func ReadTemperatureSensors() ([]*TemperatureSensorInfo, error) {
	loaded, err := hwinfolib.TryLoadLibrary()
	if err != nil {
		logger.WithError(err).Debug("error while loading hwinfo lib")
	}

	if loaded {
		results, err := readTemperatureFromHwinfoLib()
		if err != nil {
			logger.WithError(err).Error("cannot read temperature from hwinfo lib")
			return results, nil
		}
		return results, nil
	}

	return readWMITemperatureSensors()
}

// Shutdown frees resources
func Shutdown() {
	err := hwinfolib.DeInit()
	if err != nil {
		logger.WithError(err).Debugf("while trying to DeInit the hwinfo lib")
	}
}

func readTemperatureFromHwinfoLib() ([]*TemperatureSensorInfo, error) {
	devicesCount, err := hwinfolib.GetNumberOfDetectedSensors()
	if err != nil {
		return nil, err
	}

	result := make([]*TemperatureSensorInfo, 0)

	for deviceIndex := 0; deviceIndex < devicesCount; deviceIndex++ {
		err = hwinfolib.ReadDataFromSensor(deviceIndex)
		if err != nil {
			logger.WithError(err).Debugf("while trying to ReadDataFromSensor %d", deviceIndex)
			continue
		}
		deviceName, err := hwinfolib.GetSensorName(deviceIndex)
		if err != nil {
			logger.WithError(err).Debugf("while trying to GetSensorName %d", deviceIndex)
			continue
		}

		for sensorIndex := 0; sensorIndex < 512; sensorIndex++ {
			sensorName, temperature, err := hwinfolib.GetTemperature(deviceIndex, sensorIndex)
			if err != nil {
				logger.WithError(err).Debugf("while trying to GetTemperature (%s) %d, %d", deviceName, deviceIndex, sensorIndex)
				continue
			}
			if sensorName == "" {
				continue
			}

			result = append(result, &TemperatureSensorInfo{
				SensorName:  fmt.Sprintf("%s - %s", deviceName, sensorName),
				Temperature: temperature,
				Unit:        unitCelsius,
			})
		}
	}

	return result, nil
}

func readWMITemperatureSensors() ([]*TemperatureSensorInfo, error) {
	var thermalSensors []msAcpi_ThermalZoneTemperature
	query := wmi.CreateQuery(&thermalSensors, "")

	err := wmiutil.QueryWithTimeout(readTimeout, query, &thermalSensors, wmiConnectServerArgs...)
	if err != nil {
		l := logger.WithError(err)
		errText := strings.ToLower(err.Error())
		if strings.Contains(errText, "not supported") {
			l.Debugf("not supported by BIOS or driver is required")
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
