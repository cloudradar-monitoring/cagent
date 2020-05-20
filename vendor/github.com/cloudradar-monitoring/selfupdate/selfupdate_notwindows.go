//+build !windows

package selfupdate

import "github.com/pkg/errors"

func runPackageInstaller(packageFilePath string) error {
	return errors.New("package update is not implemented for this operating system")
}

func verifyPackageSignature(packageFilePath string) error {
	return nil
}
