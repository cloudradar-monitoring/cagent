// +build windows

package hyperv

import (
	"fmt"
	"time"

	"github.com/cloudradar-monitoring/cagent/perfcounters"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat"
	"github.com/cloudradar-monitoring/cagent/pkg/wmi"

	log "github.com/sirupsen/logrus"
)

func init() {
	prov := newImpl()

	err := vmstat.RegisterVMProvider(prov)
	if err != nil {
		panic(err.Error())
	}
}

type impl struct {
	watcher *perfcounters.WinPerfCountersWatcher
}

var _ vmstat.Provider = (*impl)(nil)

func newImpl() vmstat.Provider {
	return &impl{
		watcher: monitoring.GetWatcher(),
	}
}

func (im *impl) Run() error {
	interval := time.Second * 1

	err := im.watcher.StartContinuousQuery(hypervPath, interval)
	if err != nil {
		log.Printf("Failed to StartQuery: %s", err)
		return err
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
		return fmt.Errorf("%s %s", vmstat.ErrCheck.Error(), err.Error())
	}

	if st != wmiutil.FeatureInstallStateEnabled {
		return vmstat.ErrNotAvailable
	}

	return nil
}
