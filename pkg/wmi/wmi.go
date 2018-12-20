// +build windows

package wmiutil

import (
	"fmt"

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
