package cagent

import (
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/networking"
)

func (ca *Cagent) GetNetworkWatcher() *networking.NetWatcher {
	if ca.netWatcher != nil {
		return ca.netWatcher
	}

	maxSpeed, _ := ca.Config.GetParsedNetInterfaceMaxSpeed()
	ca.netWatcher = networking.NewWatcher(
		networking.NetWatcherConfig{
			NetInterfaceExclude:             ca.Config.NetInterfaceExclude,
			NetInterfaceExcludeRegex:        ca.Config.NetInterfaceExcludeRegex,
			NetInterfaceExcludeDisconnected: ca.Config.NetInterfaceExcludeDisconnected,
			NetInterfaceExcludeLoopback:     ca.Config.NetInterfaceExcludeLoopback,
			NetMetrics:                      ca.Config.NetMetrics,
			NetInterfaceMaxSpeed:            maxSpeed,
		},
	)
	return ca.netWatcher
}
