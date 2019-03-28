// +build linux

package sensors

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ReadTemperatureSensors() ([]*TemperatureSensorInfo, error) {
	var temperatures []*TemperatureSensorInfo
	files, err := filepath.Glob(getHostSys("/class/hwmon/hwmon*/temp*_*"))
	if err != nil {
		logger.WithError(err).Error("failed to list sensors")
		return temperatures, err
	}
	if len(files) == 0 {
		// CentOS has an intermediate /device directory:
		files, err = filepath.Glob(getHostSys("/class/hwmon/hwmon*/device/temp*_*"))
		if err != nil {
			logger.WithError(err).Error("failed to list sensors")
			return temperatures, err
		}
	}

	// Example directory:
	//   device/           temp1_crit_alarm  temp2_crit_alarm  temp3_crit_alarm  temp4_crit_alarm
	//   name              temp1_input       temp2_input       temp3_input       temp4_input
	//   power/            temp1_label       temp2_label       temp3_label       temp4_label
	//   subsystem/        temp1_max         temp2_max         temp3_max         temp4_max
	//   temp1_crit        temp2_crit        temp3_crit        temp4_crit        uevent
	for _, file := range files {
		filename := strings.SplitN(filepath.Base(file), "_", 2)
		baseFileName := filename[0]
		suffix := filename[1]
		if suffix != "input" {
			// skip
			continue
		}

		labelBytes, _ := ioutil.ReadFile(filepath.Join(filepath.Dir(file), baseFileName+"_label"))
		label := strings.TrimSpace(string(labelBytes))

		nameBytes, err := ioutil.ReadFile(filepath.Join(filepath.Dir(file), "name"))
		if err != nil {
			logger.WithError(err).Debug("could not read 'name' file")
			continue
		}
		name := strings.TrimSpace(string(nameBytes))

		temperature, err := readTemperatureFromFile(file)
		if err != nil {
			logger.WithError(err).Debugf("could not read temperature from file: %s", file)
			continue
		}

		criticalTemp, _ := readTemperatureFromFile(filepath.Join(filepath.Dir(file), baseFileName+"_crit"))

		temperatures = append(temperatures, &TemperatureSensorInfo{
			SensorName:  fmt.Sprintf("%s:%s:%s", name, label, baseFileName),
			Temperature: temperature,
			Critical:    criticalTemp,
			Unit:        unitCelsius,
		})
	}
	return temperatures, nil
}

func readTemperatureFromFile(filePath string) (float64, error) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return 0, err
	}
	temperature, err := strconv.ParseFloat(strings.TrimSpace(string(fileContent)), 64)
	if err != nil {
		return 0, err
	}
	return temperature / 1000.0, nil
}

func getHostSys(combineWith string) string {
	var result = "/sys"
	if hostProc := os.Getenv("HOST_SYS"); hostProc != "" {
		result = hostProc
	}

	return filepath.Join(result, combineWith)
}