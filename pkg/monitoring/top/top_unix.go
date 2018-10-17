// +build darwin linux, !windows

package top

import (
	"bytes"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func (t *Top) GetProcesses() ([]*ProcessInfo, error) {
	var buff bytes.Buffer

	// Command to list processes
	cmdPS := exec.Command("ps", "ax", "-o", "pid,%cpu,command")
	cmdPS.Stdout = &buff
	err := cmdPS.Run()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to call: ps ax -o")
	}

	lines := strings.Split(buff.String(), "\n")
	processes := make([]*ProcessInfo, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		parts1 := strings.Split(line, "   ")

		// If load is >= 10 there are only two spaces
		if len(parts1) != 2 {
			parts1 = strings.Split(line, "  ")
		}

		// Workaround if format is a bit off for some reason.
		// Maybe could make sense to log these and investigate later
		if len(parts1) < 2 {
			continue
		}

		parts2 := strings.SplitN(parts1[1], " ", 2)

		parsedLoad, err := strconv.ParseFloat(parts2[0], 64)
		if err != nil {
			log.Printf("Failed to parse load from: %s", parts2[0])
			continue
		}

		parsedPID, err := strconv.ParseUint(parts1[0], 10, 32)
		if err != nil {
			log.Printf("Failed to parse PID from: %s", parts1[0])
			continue
		}

		// Workaround if format is a bit off for some reason.
		// Maybe could make sense to log these and investigate later
		if len(parts2) < 2 {
			log.Println("Splitting error2:")
			log.Printf("%+v", line)
			log.Printf("%+v", parts1)
			log.Printf("%+v", parts2)
			continue
		}

		p := &ProcessInfo{
			PID:     uint32(parsedPID),
			Command: parts2[1],
			Load:    parsedLoad,
		}

		processes = append(processes, p)
	}

	return processes, nil
}
