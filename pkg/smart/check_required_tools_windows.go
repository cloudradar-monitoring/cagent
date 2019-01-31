// +build windows

package smart

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
)

func checkTools() (string, error) {
	cmd := exec.Command("smartctl.exe", "-h")
	buf := &bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(buf)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("smart: cannot get smartctl version")
	}

	return buf.String(), nil
}
