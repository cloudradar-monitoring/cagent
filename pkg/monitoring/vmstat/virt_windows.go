// +build windows

package vmstat

import (
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat/hyperv"
)

// Init instantiate virtual machines that is only windows specific
func Init() error {
	prov := hyperv.New()
	if err := RegisterVMProvider(prov); err != nil {
		return err
	}

	return nil
}
