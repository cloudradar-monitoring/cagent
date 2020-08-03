package raid

import (
	"regexp"
	"strconv"
	"strings"
)

type raidArrays []raidInfo

type raidInfo struct {
	Name         string
	Type         string
	State        string
	RaidLevel    int
	Devices      []string
	Inactive     []int
	Active       []int
	Failed       []int
	IsRebuilding bool
}

var raidStatusRegex = regexp.MustCompile(`\[([U_]+)\]`)

func (r raidInfo) GetFailedDevices() (failedDevices []string) {
	for _, deviceIndex := range r.Failed {
		if deviceIndex < len(r.Devices) {
			failedDevices = append(failedDevices, r.Devices[deviceIndex])
		}
	}

	return failedDevices
}

func parseMdstat(data string) raidArrays {
	var raids []raidInfo
	var err error
	lines := strings.Split(data, "\n")

	for n, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Personalities") || strings.HasPrefix(line, "unused") {
			continue
		}

		line = strings.ReplaceAll(line, "(auto-read-only)", "")

		parts := strings.Fields(line)
		if len(parts) < 5 || parts[1] != ":" {
			continue
		}
		raidState := parts[2]
		raidType := ""
		raidLevel := 0
		deviceIndex := 3
		if raidState != "inactive" {
			raidType = parts[3]
			raidLevel, err = strconv.Atoi(strings.TrimPrefix(raidType, "raid"))
			if err != nil {
				log.WithError(err).Warnf("could not determine raid level from line '%s'", line)
				raidLevel = -1
			}
			deviceIndex = 4
		}
		raid := raidInfo{Name: parts[0], State: raidState, Type: raidType, RaidLevel: raidLevel, Devices: parts[4:]}

		raid.Devices = parts[deviceIndex:]
		for i, device := range raid.Devices {
			p := strings.Index(device, "[")
			if p > 0 {
				raid.Devices[i] = device[0:p]
				if strings.Contains(device, "(F)") {
					raid.Failed = append(raid.Failed, i)
				}
			}
		}

		if len(lines) <= n+3 {
			log.Errorf("error parsing %s: too few lines for md device", raid.Name)
			return raids
		}

		raid.Inactive, raid.Active = parseStatusLine(lines[n+1])

		syncLineIdx := n + 2
		if strings.Contains(lines[n+2], "bitmap") {
			// skip bitmap line
			syncLineIdx++
		}

		isRecovering := strings.Contains(lines[syncLineIdx], "recovery")
		if isRecovering {
			raid.IsRebuilding = true
		}

		raids = append(raids, raid)
	}
	return raids
}

func parseStatusLine(line string) ([]int, []int) {
	var inactiveDevs, activeDevs []int
	matches := raidStatusRegex.FindStringSubmatch(line)
	if len(matches) > 0 {
		// Parse raid array status from mdstat output e.g. "[UUU_]"
		// if device is up("U") or down/missing ("_")
		for i := 0; i < len(matches[1]); i++ {
			if matches[1][i:i+1] == "_" {
				inactiveDevs = append(inactiveDevs, i)
			} else if matches[1][i:i+1] == "U" {
				activeDevs = append(activeDevs, i)
			}
		}
	}
	return inactiveDevs, activeDevs
}
