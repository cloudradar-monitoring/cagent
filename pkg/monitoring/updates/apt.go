package updates

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type pkgMgrApt struct {
}

func (a *pkgMgrApt) GetBinaryPath() string {
	return "/usr/bin/apt-get"
}

func (a *pkgMgrApt) FetchUpdates(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sudo", a.GetBinaryPath(), "update", "-q", "-y")

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
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

func (a *pkgMgrApt) GetAvailableUpdatesCount() (int, error) {
	// disable gosec G204 cmd audit:
	/* #nosec */
	cmd := exec.Command("sudo", a.GetBinaryPath(), "upgrade", "--dry-run")
	out, err := cmd.Output()
	if err != nil {
		return 0, errors.Wrap(err, "while trying to list available updates")
	}

	result := 0
	outLines := strings.Split(string(out), "\n")
	for _, line := range outLines {
		if strings.HasPrefix(line, "Inst ") { // this prefix is locale-independent
			result++
		}
	}
	return result, nil
}
