// +build !windows,!linux

package sensors

func ReadTemperatureSensors() ([]*TemperatureSensorInfo, error) {
	log.Info("not implemented for this OS")
	return nil, nil
}

func Shutdown() {
	
}