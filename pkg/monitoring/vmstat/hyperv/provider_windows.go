// +build windows

package hyperv

import (
	"fmt"
	"time"

	"github.com/cloudradar-monitoring/cagent/perfcounters"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat/types"
	"github.com/cloudradar-monitoring/cagent/pkg/wmi"
)

type impl struct {
	watcher *perfcounters.WinPerfCountersWatcher
}

var _ vmstattypes.Provider = (*impl)(nil)

func New() vmstattypes.Provider {
	return &impl{
		watcher: monitoring.GetWatcher(),
	}
}

func (im *impl) Run() error {
	interval := time.Second * 1

	err := im.watcher.StartContinuousQuery(hypervPath, interval)
	if err != nil {
		return fmt.Errorf("vmstat run failure: %s", err.Error())
	}

	return nil
}

func (im *impl) Shutdown() error {
	return nil
}

func (im *impl) Name() string {
	return "hyper-v"
}

func (im *impl) IsAvailable() error {
	st, err := wmiutil.CheckOptionalFeatureStatus(wmiutil.FeatureMicrosoftHyperV)

	if err != nil {
		return fmt.Errorf("%s %s", vmstattypes.ErrCheck.Error(), err.Error())
	}

	if st != wmiutil.FeatureInstallStateEnabled {
		return vmstattypes.ErrNotAvailable
	}

	return nil
}
