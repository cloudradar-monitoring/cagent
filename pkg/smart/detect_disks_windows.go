// +build windows

package smart

import (
	"bufio"
	"bytes"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func (sm *SMART) detectDisks() (*bytes.Buffer, error) {
	cmd := exec.Command("cmd", "/c", sm.smartctl, "--scan")

	buf := &bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(buf)

	if err := cmd.Run(); err != nil {
		log.Errorf("smart: execute smartctl.exe: %s", err.Error())
		return nil, ErrUnderlyingToolNotFound
	}

	return buf, nil
}
