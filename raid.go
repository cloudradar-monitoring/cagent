package cagent

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

const (
	RaidStatusActive   = "active"
	RaidStatusDegraded = "degraded"
)

type RaidArrays []Raid

type Raid struct {
	Name    string
	Type    string
	State   string
	Devices []string
	Failed  []int
}

var failed = regexp.MustCompile("\\[([U_]+)\\]")

func (r Raid) GetFailedPhysicalDevices() []string {
	devices := []string{}
	for _, fd := range r.Failed {
		if fd < len(r.Devices) {
			devices = append(devices, r.Devices[fd])
		} else {
			devices = append(devices, fmt.Sprintf("missing_device[%d]", fd))
		}
	}
	return devices
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
				}
			}
		}
		if raid.State == RaidStatusActive && len(raid.Failed) > 0 {
			raid.State = RaidStatusDegraded
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

		if raid.State == RaidStatusDegraded {
			results[raid.Name+".failed"] = strings.Join(raid.GetFailedPhysicalDevices(), "; ")
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
