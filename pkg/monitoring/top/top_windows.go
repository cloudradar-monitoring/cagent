// +build windows

package top

import (
	"bytes"
	"log"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func (t *Top) GetProcesses() ([]*ProcessInfo, error) {
	var buff bytes.Buffer
	var errBuff bytes.Buffer

	// cmdStr :=

	// log.Printf("cmd: %s", cmdStr)

	// Command to list processes
	cmdPS := exec.Command("typeperf", "\\Process(*)\\% Processor Time", "-sc", "1")
	cmdPS.Stdout = &buff
	cmdPS.Stderr = &errBuff
	err := cmdPS.Run()
	if err != nil {
		log.Printf("Err: %s", buff.String())
		return nil, errors.Wrapf(err, "Failed to call: ps ax -o")
	}

	lines := strings.Split(buff.String(), "\n")

	processesSplit := strings.Split(lines[1], ",")
	loadsSplit := strings.Split(lines[2], ",")

	if len(processesSplit) != len(loadsSplit) {
		log.Printf("lists have different size")
	}

	for i := range processesSplit {
		log.Printf("%s - %s", processesSplit[i], loadsSplit[i])
	}

	return nil, errors.New("Not implemented")
}
