// +build darwin

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

	cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("command %s -h", smartctl))
	buildString := &bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(buildString)
	if err := cmd.Run(); err != nil {
		return "", "", errors.Wrap(err, "smart: cannot get smartctl version string")
	}

	cmd = exec.Command("/bin/sh", "-c", "command -v diskutil")
	cmd.Stdout = nil
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("smart: diskutil is not installed")
	}

	return buildString.String(), smartctl, nil
}
