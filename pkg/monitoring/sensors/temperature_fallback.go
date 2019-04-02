// +build !windows,!linux

package sensors

func ReadTemperatureSensors() ([]*TemperatureSensorInfo, error) {
	logger.Info("not implemented for this OS")
	return nil, nil
}
