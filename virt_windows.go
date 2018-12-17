// +build windows

package cagent

import (
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat/hyperv"
)

// InitVMStat instantiate virtual machines that is only windows specific
func InitVMStat() error {
	prov := hyperv.New()
	if err := vmstat.RegisterVMProvider(prov); err != nil {
		return err
	}

	return nil
}
