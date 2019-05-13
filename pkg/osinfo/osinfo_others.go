// +build !linux,!windows

package osinfo

import (
	"os/exec"

	"github.com/pkg/errors"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

func osName() (string, error) {
	uname, err := exec.LookPath("uname")
	if err != nil {
		return "", errors.Wrap(err, "osinfo: lookup for \"uname\" command")
	}

	var data []byte
	data, err = common.RunCommandInBackground(uname, "-a")
	if err != nil {
		return "", errors.Wrap(err, "osinfo: while running \"uname\"")
	}

	return string(data), nil
}
