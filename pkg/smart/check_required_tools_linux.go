// +build !darwin,!windows

package smart

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

const defaultSmartctlPath = "smartctl"

func checkTools(smartctl string) (string, string, error) {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("which %s", smartctl))
	smartctlPathBuf := &bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(smartctlPathBuf)
	if err := cmd.Run(); err != nil {
		return "", "", errors.Wrapf(ErrSmartctlNotFound, "\"%s\"", smartctl)
	}
	smartctl = strings.TrimRight(smartctlPathBuf.String(), "\n")

	cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s -h", smartctl))
	buf := &bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(buf)
	if err := cmd.Run(); err != nil {
		return "", "", errors.Wrap(err, "smart: cannot get smartctl version string")
	}

	return buf.String(), smartctl, nil
}
