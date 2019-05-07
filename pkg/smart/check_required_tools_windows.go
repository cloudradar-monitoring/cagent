// +build windows

package smart

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

const defaultSmartctlPath = "smartctl.exe"

func checkTools(smartctl string) (string, string, error) {
	cmd := exec.Command("cmd", "/c", fmt.Sprintf(`where %s`, smartctl))
	smartctlPathBuf := &bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(smartctlPathBuf)
	if err := cmd.Run(); err != nil {
		return "", "", errors.Wrap(err, "smart: detect full path of smartctl.exe")
	}

	smartctl = strings.TrimRight(smartctlPathBuf.String(), "\r\n")

	cmd = exec.Command("cmd", "/c", smartctl, "-h")

	buf := &bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(buf)

	if err := cmd.Run(); err != nil {
		return "", "", errors.Wrap(err, "smart: cannot get smartctl version string")
	}

	return buf.String(), smartctl, nil
}
