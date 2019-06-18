package cagent

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/docker"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat/types"
	"github.com/cloudradar-monitoring/cagent/pkg/smart"
)

type Cagent struct {
	Config         *Config
	ConfigLocation string

	hubClient     *http.Client
	hubClientOnce sync.Once

	cpuWatcher             *CPUWatcher
	cpuUtilisationAnalyser *CPUUtilisationAnalyser

	fsWatcher            *FSWatcher
	netWatcher           *NetWatcher
	windowsUpdateWatcher *WindowsUpdateWatcher // nolint: structcheck,megacheck
	dockerWatcher        *docker.Watcher

	vmstatLazyInit sync.Once
	vmWatchers     map[string]types.Provider
	hwInventory    sync.Once
	smart          *smart.SMART

	rootCAs *x509.CertPool

	version string
}

func New(cfg *Config, cfgPath string, version string) *Cagent {
	ca := &Cagent{
		Config:         cfg,
		ConfigLocation: cfgPath,
		version:        version,
		vmWatchers:     make(map[string]types.Provider),
		dockerWatcher:  &docker.Watcher{},
	}

	if rootCertsPath != "" {
		if _, err := os.Stat(rootCertsPath); err == nil {
			certPool := x509.NewCertPool()

			b, err := ioutil.ReadFile(rootCertsPath)
			if err != nil {
				logrus.WithError(err).Warnln("Failed to read cacert.pem")
			} else {
				ok := certPool.AppendCertsFromPEM(b)
				if ok {
					ca.rootCAs = certPool
				}
			}
		}
	}

	ca.configureLogger()

	if ca.Config.SMARTMonitoring && ca.Config.SMARTCtl != "" {
		var err error
		ca.smart, err = smart.New(smart.Executable(ca.Config.SMARTCtl, false))
		if err != nil {
			logrus.Error(err.Error())
		}
	}
	logrus.Debug("DEBUG MESSAGE")
	logrus.Info("INFO MESSAGE")
	logrus.Warn("WARNING MESSAGE")

	return ca
}

func (ca *Cagent) SetVersion(version string) {
	ca.version = version
}

func (ca *Cagent) userAgent() string {
	if ca.version == "" {
		ca.version = "{undefined}"
	}
	return fmt.Sprintf("Cagent v%s %s %s", ca.version, runtime.GOOS, runtime.GOARCH)
}

func (ca *Cagent) Shutdown() error {
	for name, p := range ca.vmWatchers {
		if err := vmstat.Release(p); err != nil {
			logrus.WithFields(logrus.Fields{
				"name": name,
			}).WithError(err).Warnln("unable to release vm provider")
		}

		delete(ca.vmWatchers, name)
	}

	return nil
}
