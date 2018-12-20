package cagent

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat"

	log "github.com/sirupsen/logrus"
)

type Cagent struct {
	Config *Config

	// internal use
	hubHTTPClient *http.Client

	cpuWatcher           *CPUWatcher
	fsWatcher            *FSWatcher
	netWatcher           *NetWatcher
	windowsUpdateWatcher *WindowsUpdateWatcher // nolint: structcheck,megacheck
	vmWatchers           map[string]vmstat.Provider

	rootCAs *x509.CertPool

	version string
}

func New(cfg *Config, version string) *Cagent {
	ca := &Cagent{
		Config:     cfg,
		version:    version,
		vmWatchers: make(map[string]vmstat.Provider),
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

	if ca.Config.LogFile != "" {
		err := addLogFileHook(ca.Config.LogFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			log.Error("Can't write logs to file: ", err.Error())
		}
	}

	if err := InitVMStat(); err != nil {
		log.Error("cannot instantiate virtual machines API: ", err.Error())
	}

	for _, name := range ca.Config.VirtualMachinesStat {
		vm, err := vmstat.Acquire(name)
		if err != nil {
			log.Warnf("acquire vm provider \"%s\": %s", name, err.Error())
		} else {
			if err = vm.IsAvailable(); err != nil {
				log.Warnf("vm provider \"%s\" either not available or not enabled for host system %s: %s", name, runtime.GOOS, err.Error())
			}

			ca.vmWatchers[name] = vm
		}
	}

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
