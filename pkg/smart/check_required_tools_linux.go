// +build !darwin,!windows

package smart

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
)

func checkTools() (string, error) {
	cmd := exec.Command("/bin/sh", "-c", "smartctl -h")
	buf := &bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(buf)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("smart: smartctl is not installed")
	}

	return buf.String(), nil
}
