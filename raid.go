package cagent

import (
	"io/ioutil"
	"regexp"
	"strings"
)

type RaidArrays []Raid

type Raid struct {
	Name    string
	Type    string
	State   string
	Devices []string
	Failed  []int
	Active  []int
}

var failed = regexp.MustCompile("\\[([U_]+)\\]")

func (r Raid) GetFailedAndMissingPhysicalDevices() (failedDevices []string, missingDevicesCount int) {
	for _, fd := range r.Failed {
		if fd < len(r.Devices) {
			failedDevices = append(failedDevices, r.Devices[fd])
		} else {
			missingDevicesCount++
		}
	}
	return failedDevices, missingDevicesCount
}

func (r Raid) GetActivePhysicalDevices() []string {
	var activeDevices []string
	for _, fd := range r.Active {
		if fd < len(r.Devices) {
			activeDevices = append(activeDevices, r.Devices[fd])
		}
	}
	return activeDevices
}

func parseMdstat(data string) (RaidArrays, error) {

	raids := []Raid{}
	lines := strings.Split(data, "\n")

	for n, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Personalities") || strings.HasPrefix(line, "unused") {
			continue
		}

		parts := strings.Split(line, " ")

		if len(parts) < 5 || parts[1] != ":" {
			continue
		}
		raid := Raid{Name: parts[0], State: parts[2], Type: parts[3], Devices: parts[4:]}

		raid.Devices = parts[4:]
		for i, device := range raid.Devices {
			p := strings.Index(device, "[")
			if p > 0 {
				raid.Devices[i] = device[0:p]
			}
		}
		matches := failed.FindStringSubmatch(lines[n+1])

		if len(matches) > 0 {
			for i := 0; i < len(matches[1]); i++ {
				if matches[1][i:i+1] == "_" {
					raid.Failed = append(raid.Failed, i)
				} else if matches[1][i:i+1] == "U" {
					raid.Active = append(raid.Active, i)
				}
			}
		}

		raids = append(raids, raid)
	}
	return raids, nil
}

func (ar RaidArrays) Measurements() MeasurementsMap {
	results := MeasurementsMap{}

	for _, raid := range ([]Raid)(ar) {
		results[raid.Name+".state"] = raid.State
		results[raid.Name+".type"] = raid.Type

		failedDevices, missingCount := raid.GetFailedAndMissingPhysicalDevices()

		if len(failedDevices) > 0 || missingCount > 0 {
			results[raid.Name+".degraded"] = 1

			results[raid.Name+".physicaldevice.missing"] = missingCount

			for _, failedDevice := range failedDevices {
				results[raid.Name+".physicaldevice.state."+failedDevice] = "failed"
			}

		} else {
			results[raid.Name+".degraded"] = 0
		}

		activeDevices := raid.GetActivePhysicalDevices()
		for _, activeDevice := range activeDevices {
			results[raid.Name+".physicaldevice.state."+activeDevice] = "active"
		}
	}
	return results
}

func (ca *Cagent) RaidState() (MeasurementsMap, error) {
	buf, err := ioutil.ReadFile("/proc/mdstat")
	if err != nil {
		return nil, err
	}

	raidArrays, err := parseMdstat(string(buf))

	if err != nil {
		return MeasurementsMap{}, err
	}

	return raidArrays.Measurements(), nil
}
