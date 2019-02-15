// +build windows

package wmiutil

import (
	"fmt"
	"time"

	"github.com/StackExchange/wmi"
)

const (
	FeatureMicrosoftHyperV    = "Microsoft-Hyper-V"
	FeatureMicrosoftHyperVAll = "Microsoft-Hyper-V-All"
)

type FeatureInstallState uint32

const (
	FeatureInstallStateEnabled FeatureInstallState = iota + 1
	FeatureInstallStateDisabled
	FeatureInstallStateAbsent
	FeatureInstallStateUnknown
)

const queryTimeout = 10 * time.Second

type win32_optionalfeature struct {
	InstallState FeatureInstallState
}

func (st FeatureInstallState) String() string {
	switch st {
	case 1:
		return "enabled"
	case 2:
		return "disabled"
	case 3:
		return "absent"
	default:
		return "unknown"
	}
}

func CheckOptionalFeatureStatus(feature string) (FeatureInstallState, error) {
	var dst []win32_optionalfeature
	q := wmi.CreateQuery(&dst, "WHERE name = \""+feature+"\"")
	err := QueryWithTimeout(queryTimeout, q, &dst)
	if err != nil {
		return FeatureInstallStateUnknown, fmt.Errorf("wmiutil: check feature status %s", err.Error())
	}

	if len(dst) == 0 {
		return FeatureInstallStateUnknown, fmt.Errorf("wmiutil: couldn't check feature install status")
	}

	return dst[0].InstallState, nil
}

func QueryWithTimeout(timeout time.Duration, query string, dst interface{}, connectServerArgs ...interface{}) error {
	errChan := make(chan error, 1)

	go func() {
		errChan <- wmi.Query(query, dst, connectServerArgs...)
	}()

	select {
	case <-time.After(timeout):
		return fmt.Errorf("wmiutil: query timedout")
	case err := <-errChan:
		return err
	}
}
