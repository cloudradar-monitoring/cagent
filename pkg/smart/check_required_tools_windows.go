// +build windows

package smart

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
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
		// where command might not be available or smartmontools bin directory is not present in the PATH
		if _, err = os.Stat(smartctl); os.IsNotExist(err) {
			return "", "", errors.Wrapf(ErrSmartctlNotFound, "\"%s\"", smartctl)
		}

		smartctlPathBuf.Write([]byte(smartctl))
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
