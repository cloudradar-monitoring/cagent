package cagent

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/fs"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/networking"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/sensors"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/updates"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat/types"
	"github.com/cloudradar-monitoring/cagent/pkg/smart"
)

// variables set on build. Example:
// go build -o cagent -ldflags="-X github.com/cloudradar-monitoring/cagent.Version=$(git --git-dir=src/github.com/cloudradar-monitoring/cagent/.git describe --always --long --dirty --tag)" github.com/cloudradar-monitoring/cagent/cmd/cagent
var (
	Version     string
	LicenseInfo = "released under MIT license. https://github.com/cloudradar-monitoring/cagent/"
)

type Cagent struct {
	Config         *Config
	ConfigLocation string

	hubClient     *http.Client
	hubClientOnce sync.Once

	cpuWatcher             *CPUWatcher
	cpuUtilisationAnalyser *CPUUtilisationAnalyser

	fsWatcher            *fs.FileSystemWatcher
	netWatcher           *networking.NetWatcher
	windowsUpdateWatcher *WindowsUpdateWatcher // nolint: structcheck,megacheck

	vmstatLazyInit sync.Once
	vmWatchers     map[string]types.Provider
	hwInventory    sync.Once
	smart          *smart.SMART
}

func New(cfg *Config, cfgPath string) *Cagent {
	ca := &Cagent{
		Config:         cfg,
		ConfigLocation: cfgPath,
		vmWatchers:     make(map[string]types.Provider),
	}

	ca.configureLogger()

	if ca.Config.SMARTMonitoring && ca.Config.SMARTCtl != "" {
		var err error
		ca.smart, err = smart.New(smart.Executable(ca.Config.SMARTCtl, false))
		if err != nil {
			logrus.Error(err.Error())
		}
	}

	return ca
}

func (ca *Cagent) userAgent() string {
	if Version == "" {
		Version = "{undefined}"
	}
	return fmt.Sprintf("Cagent v%s %s %s", Version, runtime.GOOS, runtime.GOARCH)
}

func (ca *Cagent) Shutdown() {
	defer sensors.Shutdown()
	defer updates.Shutdown()

	for name, p := range ca.vmWatchers {
		if err := vmstat.Release(p); err != nil {
			logrus.WithFields(logrus.Fields{
				"name": name,
			}).WithError(err).Warnln("unable to release vm provider")
		}

		delete(ca.vmWatchers, name)
	}
}
