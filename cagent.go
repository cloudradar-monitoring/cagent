package cagent

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat/types"

	log "github.com/sirupsen/logrus"
)

type Cagent struct {
	Config         *Config
	ConfigLocation string

	// internal use
	hubHTTPClient *http.Client

	cpuWatcher           *CPUWatcher
	fsWatcher            *FSWatcher
	netWatcher           *NetWatcher
	windowsUpdateWatcher *WindowsUpdateWatcher // nolint: structcheck,megacheck
	vmstatLazyInit       sync.Once
	vmWatchers           map[string]vmstattypes.Provider

	rootCAs *x509.CertPool

	version string
}

func New(cfg *Config, cfgPath string, version string) *Cagent {
	ca := &Cagent{
		Config:         cfg,
		ConfigLocation: cfgPath,
		version:        version,
		vmWatchers:     make(map[string]vmstattypes.Provider),
	}

	if rootCertsPath != "" {
		if _, err := os.Stat(rootCertsPath); err == nil {
			certPool := x509.NewCertPool()

			b, err := ioutil.ReadFile(rootCertsPath)
			if err != nil {
				log.Error("Failed to read cacert.pem: ", err.Error())
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
	parts := strings.Split(ca.version, "-")

	return fmt.Sprintf("Cagent v%s %s %s", parts[0], runtime.GOOS, runtime.GOARCH)
}

func (ca *Cagent) Shutdown() error {
	for name, p := range ca.vmWatchers {
		if err := vmstat.Release(p); err != nil {
			log.Errorf("release vm provider \"%s\": %s", name, err.Error())
		}

		delete(ca.vmWatchers, name)
	}

	return nil
}
