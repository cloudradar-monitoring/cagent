// +build !windows

package updates

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

type pkgMgrYUM struct {
	BinaryPath             string
	fetchedTotalUpdates    int
	fetchedSecurityUpdates *int
}

func (a *pkgMgrYUM) GetBinaryPath() string {
	return a.BinaryPath
}

func (a *pkgMgrYUM) FetchUpdates(timeout time.Duration) error {
	a.fetchedTotalUpdates = 0
	a.fetchedSecurityUpdates = nil

	err := a.fetchTotalUpdates(timeout)
	if err != nil {
		return err
	}

	return a.fetchSecurityUpdates(timeout)
}

func (a *pkgMgrYUM) fetchTotalUpdates(timeout time.Duration) error {
	out, err := common.RunCommandWithTimeout(timeout, "sudo", a.GetBinaryPath(), "-q", "check-update")
	if err == common.ErrCommandExecutionTimeout {
		return fmt.Errorf("timeout of %s exceeded while fetching new updates", timeout)
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()

			// Returns 0 if no packages are available for update.
			if code == 0 {
				return nil
			}

			// Returns exit value of 100 if there are packages available for an update.
			if code == 100 {
				a.fetchedTotalUpdates = parseYUMOutput(&out)
				return nil
			}

			return errors.Wrap(exitErr, "while executing fetch command")
		}
	}

	return err
}

func (a *pkgMgrYUM) fetchSecurityUpdates(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sudo", a.GetBinaryPath(), "-q", "list-security")

	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("timeout of %s exceeded while fetching security updates list", timeout)
	}

	if err != nil {
		// list-security isn't available on all systems with yum. Ignore error
		log.WithError(err).Debugf("while executing 'list-security'. Check that yum plugin installed. %s", out)
		return nil
	}

	securityUpdatesCount := parseYUMOutput(&out)
	a.fetchedSecurityUpdates = &securityUpdatesCount

	return nil
}

func parseYUMOutput(out *[]byte) int {
	outputLines := strings.Split(string(*out), "\n")
	result := 0
	for _, line := range outputLines {
		fields := strings.Fields(line)
		// the output can include two-word headers and information on obsoleted packages, which has more 4-6 fields
		if len(fields) == 3 {
			result++
		}
	}
	return result
}

func (a *pkgMgrYUM) GetAvailableUpdatesCount() (int, *int, error) {
	return a.fetchedTotalUpdates, a.fetchedSecurityUpdates, nil
}
