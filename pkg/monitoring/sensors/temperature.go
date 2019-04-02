package sensors

import log "github.com/sirupsen/logrus"

const unitCelsius = "centigrade"

type TemperatureSensorInfo struct {
	SensorName  string  `json:"sensor_name"`
	Temperature float64 `json:"temperature"`
	Critical    float64 `json:"critical"` // a temperature threshold set by sensor, driver or configuration
	Unit        string  `json:"unit"`
}

var logger = log.WithField("package", "sensors")
