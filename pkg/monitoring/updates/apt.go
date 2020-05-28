// +build !windows

package updates

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

type pkgMgrApt struct {
}

func (a *pkgMgrApt) GetBinaryPath() string {
	return "/usr/bin/apt-get"
}

func (a *pkgMgrApt) FetchUpdates(timeout time.Duration) error {
	_, err := common.RunCommandWithTimeout(timeout, "sudo", a.GetBinaryPath(), "update", "-q", "-y")
	if err == common.ErrCommandExecutionTimeout {
		return fmt.Errorf("timeout of %s exceeded while fetching new updates", timeout)
	}

	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			return errors.Wrap(exitErr, "while executing fetch command")
		}
	}
	return err
}

func (a *pkgMgrApt) GetAvailableUpdatesCount() (int, *int, error) {
	totalUpgrades, securityUpgrades, err := tryCallAptCheck()
	if err == nil {
		return totalUpgrades, &securityUpgrades, nil
	}
	log.WithError(err).Debugf("apt-check call failed. Falling back to apt-get")

	totalUpgrades, err = a.tryCallAptGet()
	return totalUpgrades, nil, err
}

func (a *pkgMgrApt) tryCallAptGet() (int, error) {
	// disable gosec G204 cmd audit:
	/* #nosec */
	cmd := exec.Command("sudo", a.GetBinaryPath(), "upgrade", "--dry-run")
	out, err := cmd.Output()
	if err != nil {
		return 0, errors.Wrap(err, "while trying to list available updates")
	}

	totalUpgrades := 0
	outLines := strings.Split(string(out), "\n")
	for _, line := range outLines {
		if strings.HasPrefix(line, "Inst ") { // this prefix is locale-independent
			totalUpgrades++
		}
	}
	return totalUpgrades, nil
}

func tryCallAptCheck() (int, int, error) {
	cmd := exec.Command("/usr/lib/update-notifier/apt-check")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Split(string(out), ";")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("unexpected output of apt-check: %s", out)
	}

	upgrades, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("can't parse upgrades count: %s", parts[0])
	}
	securityUpgrades, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("can't parse security upgrades count: %s", parts[1])
	}
	return upgrades, securityUpgrades, nil
}
