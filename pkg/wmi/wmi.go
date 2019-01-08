// +build windows

package wmiutil

import (
	"context"
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
	err := wmi.Query(q, &dst)
	if err != nil {
		return FeatureInstallStateUnknown, fmt.Errorf("wmiutil: check feature status %s", err.Error())
	}

	if len(dst) == 0 {
		return FeatureInstallStateUnknown, fmt.Errorf("wmiutil: feature request returned empty response %s", err.Error())
	}

	return dst[0].InstallState, nil
}

func QueryWithContext(timeout time.Duration, query string, dst interface{}, connectServerArgs ...interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- wmi.Query(query, dst, connectServerArgs...)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
