package sensors

import (
	"context"
	"strings"
	"time"

	"github.com/shirou/gopsutil/host"
	log "github.com/sirupsen/logrus"
)

const readTimeout = time.Second * 10
const unitCelsius = "centigrade"

type TemperatureSensorInfo struct {
	SensorName  string  `json:"sensor_name"`
	Temperature float64 `json:"temperature"`
	Unit        string  `json:"unit"`
}

func ReadTemperatureSensors() ([]*TemperatureSensorInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	sensors, err := host.SensorsTemperaturesWithContext(ctx)
	if err != nil {
		l := log.WithField("package", "sensors").WithError(err)
		errText := strings.ToLower(err.Error())
		if strings.Contains(errText, "not implemented") {
			l.Info("reading temperature sensors is not implemented for this OS")
		} else if strings.Contains(errText, "not supported") {
			l.Debugf("not supported by BIOS or driver required")
		} else {
			l.Error("failed to read temperature sensors")
		}
		return nil, nil
	}

	result := make([]*TemperatureSensorInfo, 0)
	for _, sensor := range sensors {
		result = append(result, &TemperatureSensorInfo{
			SensorName:  sensor.SensorKey,
			Temperature: sensor.Temperature,
			Unit:        unitCelsius,
		})
	}
	return result, nil
}
