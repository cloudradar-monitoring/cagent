package sensors

import "github.com/sirupsen/logrus"

const unitCelsius = "centigrade"

type TemperatureSensorInfo struct {
	SensorName        string  `json:"sensor_name"`
	Temperature       float64 `json:"temperature"`
	CriticalThreshold float64 `json:"critical_threshold"` // a temperature threshold set by sensor, driver or configuration
	Unit              string  `json:"unit"`
}

var logger = logrus.WithField("package", "sensors")
