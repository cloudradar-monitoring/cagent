// +build darwin

package smart

import (
	"bufio"
	"bytes"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func detectDisks() (*bytes.Buffer, error) {
	cmd := exec.Command("/bin/sh", "-c", `diskutil list | grep "^/dev/" | grep -v synthesized | grep -v external | grep -v "disk image"`)

	buf := &bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(buf)

	if err := cmd.Run(); err != nil {
		log.Errorf("smart: execute diskutil: %s", err.Error())
		return nil, ErrUnderlyingToolNotFound
	}

	return buf, nil
}
