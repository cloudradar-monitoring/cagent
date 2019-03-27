package cagent

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/docker"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat/types"
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
				log.WithError(err).Warnln("Failed to read cacert.pem")
			} else {
				ok := certPool.AppendCertsFromPEM(b)
				if ok {
					ca.rootCAs = certPool
				}
			}
		}
	}

	ca.SetLogLevel(ca.Config.LogLevel)

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
			log.WithFields(log.Fields{
				"name": name,
			}).WithError(err).Warnln("unable to release vm provider")
		}

		delete(ca.vmWatchers, name)
	}

	return nil
}
