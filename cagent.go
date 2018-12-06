package cagent

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Cagent struct {
	config *Config

	// internal use
	hubHTTPClient *http.Client

	cpuWatcher           *CPUWatcher
	fsWatcher            *FSWatcher
	netWatcher           *NetWatcher
	windowsUpdateWatcher *WindowsUpdateWatcher // nolint: structcheck,megacheck

	rootCAs *x509.CertPool

	version string
}

func New(cfg *Config, version string) *Cagent {
	ca := &Cagent{
		config:  cfg,
		version: version,
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

	ca.SetLogLevel(ca.config.LogLevel)

	if ca.config.LogFile != "" {
		err := addLogFileHook(ca.config.LogFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			log.Error("Can't write logs to file: ", err.Error())
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
