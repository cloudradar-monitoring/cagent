package cagent

import (
	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/networking"
)

func (ca *Cagent) GetNetworkWatcher() *networking.NetWatcher {
	if ca.netWatcher != nil {
		return ca.netWatcher
	}

	maxSpeed, err := ca.Config.GetParsedNetInterfaceMaxSpeed()
	if err != nil {
		logrus.Errorf("invalid net_interface_max_speed value supplied: %s. network max speed will be detected automatically.", err.Error())
	}
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
