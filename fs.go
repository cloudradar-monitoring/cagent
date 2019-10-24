package cagent

import (
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/fs"
)

func (ca *Cagent) GetFileSystemWatcher() *fs.FileSystemWatcher {
	if ca.fsWatcher == nil {
		ca.fsWatcher = fs.NewWatcher(fs.FileSystemWatcherConfig{
			TypeInclude:                 ca.Config.FSTypeInclude,
			PathExclude:                 ca.Config.FSPathExclude,
			PathExcludeRecurse:          ca.Config.FSPathExcludeRecurse,
			Metrics:                     ca.Config.FSMetrics,
			IdentifyMountpointsByDevice: ca.Config.FSIdentifyMountpointsByDevice,
		})
	}

	return ca.fsWatcher
}
