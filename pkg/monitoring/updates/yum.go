package updates

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type pkgMgrYUM struct {
	BinaryPath                   string
	fetchedAvailableUpdatesCount int
}

func (a *pkgMgrYUM) GetBinaryPath() string {
	return a.BinaryPath
}

func (a *pkgMgrYUM) FetchUpdates(timeout time.Duration) error {
	a.fetchedAvailableUpdatesCount = 0
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sudo", a.GetBinaryPath(), "-q", "check-update")

	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
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
				a.fetchedAvailableUpdatesCount = parseYUMOutput(&out)
				return nil
			}

			return errors.Wrap(exitErr, "while executing fetch command")
		}
	}

	return err
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

func (a *pkgMgrYUM) GetAvailableUpdatesCount() (int, error) {
	return a.fetchedAvailableUpdatesCount, nil
}
