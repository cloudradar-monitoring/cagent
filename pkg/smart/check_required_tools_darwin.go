// +build darwin

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

	cmd = exec.Command("/bin/sh", "-c", "command -v diskutil")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("smart: diskutil is not installed")
	}

	return buf.String(), nil
}
