// +build !darwin,!windows

package smart

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const defaultSmartctlPath = "smartctl"

func checkTools(smartctl string) (string, string, error) {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("which %s", smartctl))
	smartctlPathBuf := &bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(smartctlPathBuf)
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("smart: detect full path of smartctl")
	}
	smartctl = strings.TrimRight(smartctlPathBuf.String(), "\n")

	cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("%s -h", smartctl))
	buf := &bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(buf)

	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("smart: smartctl is not installed")
	}

	return buf.String(), smartctl, nil
}
