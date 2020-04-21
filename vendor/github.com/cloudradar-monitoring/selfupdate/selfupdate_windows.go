//+build windows

package selfupdate

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

func verifyPackageSignature(packageFilePath string) error {
	expectedDisplayName := config.SigningCertificatedName
	if expectedDisplayName == "" {
		return nil
	}

	certDisplayName, err := getMSICertificateDisplayName(packageFilePath)
	if err != nil {
		return nil
	}

	if expectedDisplayName != certDisplayName {
		return fmt.Errorf("package certificate (%s) does not match expected (%s)", certDisplayName, expectedDisplayName)
	}
	return nil
}

func runPackageInstaller(packageFilePath string) error {
	msiExecPath, err := exec.LookPath("msiexec.exe")
	if err != nil {
		return errors.Wrap(err, "could not find msiexec executable")
	}

	msiExecPath, err = filepath.Abs(msiExecPath)
	if err != nil {
		return errors.Wrap(err, "could not find full path to msiexec")
	}

	// add these arguments to get installation debug log: "/l*vx", "C:\\msi.log"
	cmd := exec.Command(msiExecPath, "/quiet", "/qn", "/i", filepath.Base(packageFilePath))
	cmd.Dir = filepath.Dir(packageFilePath)

	err = cmd.Start()
	if err != nil {
		return errors.Wrapf(err, "while executing installer with args %v", cmd.Args)
	}

	// It is not clear from docs, but Release actually detaches the process
	err = cmd.Process.Release()
	if err != nil {
		return errors.Wrap(err, "could not detach child process")
	}

	log.Debugf("installer %s executed in detached mode.", packageFilePath)

	return nil
}
