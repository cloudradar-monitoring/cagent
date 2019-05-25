// +build !windows,!darwin

package smart

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

func (sm *SMART) detectDisks() (*bytes.Buffer, error) {
	cmd := exec.Command("/bin/sh", "-c", common.WrapCommandToAdmin(fmt.Sprintf("%s --scan", sm.smartctl)))

	buf := &bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(buf)

	if err := cmd.Run(); err != nil {
		log.Errorf("smart: execute smartctl: %s", err.Error())
		return nil, ErrUnderlyingToolNotFound
	}

	return buf, nil
}
