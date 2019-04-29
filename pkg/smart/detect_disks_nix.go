// +build !windows,!darwin

package smart

import (
	"bufio"
	"bytes"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func detectDisks() (*bytes.Buffer, error) {
	cmd := exec.Command("/bin/sh", "-c", "sudo smartctl --scan")

	buf := &bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(buf)

	if err := cmd.Run(); err != nil {
		log.Errorf("smart: execute smartctl.exe: %s", err.Error())
		return nil, ErrUnderlyingToolNotFound
	}

	return buf, nil
}
