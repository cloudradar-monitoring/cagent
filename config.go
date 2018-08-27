package cagent

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"crypto/x509"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

const (
	IOModeFile = "file"
	IOModeHTTP = "http"

	SenderModeWait     = "wait"
	SenderModeInterval = "interval"
)

var DefaultCfgPath string

type Cagent struct {
	Interval float64 `toml:"interval"` // interval to push metrics to the HUB

	PidFile  string   `toml:"pid"`
	LogFile  string   `toml:"log"`
	LogLevel LogLevel `toml:"log_level"`

	HubURL           string `toml:"hub_url"`
	HubGzip          bool   `toml:"hub_gzip"` // enable gzip when sending results to the HUB
	HubUser          string `toml:"hub_user"`
	HubPassword      string `toml:"hub_password"`
	HubProxy         string `toml:"hub_proxy"`
	HubProxyUser     string `toml:"hub_proxy_user"`
	HubProxyPassword string `toml:"hub_proxy_password"`

	CPULoadDataGather []string `toml:"cpu_load_data_gathering_mode"`
	CPUUtilDataGather []string `toml:"cpu_utilisation_gathering_mode"`
	CPUUtilTypes      []string `toml:"cpu_utilisation_types"`

	FSTypeInclude []string `toml:"fs_type_include"`
	FSPathExclude []string `toml:"fs_path_exclude"`
	FSMetrics     []string `toml:"fs_metrics"`

	SystemFields []string `toml:"system_fields"`

	// internal use
	hubHttpClient *http.Client

	rootCAs *x509.CertPool
	version string
}

func New() *Cagent {
	var defaultLogPath string
	var rootCertsPath string

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	switch runtime.GOOS {
	case "windows":
		DefaultCfgPath = filepath.Join(exPath, "./cagent.conf")
		defaultLogPath = filepath.Join(exPath, "./cagent.log")
	case "darwin":
		DefaultCfgPath = os.Getenv("HOME") + "/.cagent/cagent.conf"
		defaultLogPath = os.Getenv("HOME") + "/.cagent/cagent.log"
	default:
		rootCertsPath = "/etc/cagent/cacert.pem"
		DefaultCfgPath = "/etc/cagent/cagent.conf"
		defaultLogPath = "/var/log/cagent/cagent.log"
	}

	ca := &Cagent{
		LogFile:  defaultLogPath,
		Interval: 30,

		CPULoadDataGather: []string{"avg1"},
		CPUUtilTypes:      []string{"user", "system", "nice", "idle", "iowait"},
		CPUUtilDataGather: []string{"avg1"},

		FSTypeInclude: []string{"ext3", "ext4", "xfs", "jfs", "ntfs", "btrfs", "hfs", "apfs", "fat32"},
		FSMetrics:     []string{"free_B", "free_percent", "total_B"},
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

	if ca.HubURL == "" && os.Getenv("CAGENT_HUB_URL") != "" {
		ca.HubURL = os.Getenv("CAGENT_HUB_URL")
	}

	ca.SetLogLevel(LogLevelInfo)
	return ca
}

func secToDuration(secs float64) time.Duration {
	return time.Duration(int64(float64(time.Second) * secs))
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

func (ca *Cagent) ReadConfigFromFile(configFilePath string, createIfNotExists bool) error {
	dir := filepath.Dir(configFilePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.WithError(err).Errorf("Failed to create the config dir: '%s'", dir)
	}

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if !createIfNotExists {
			return fmt.Errorf("Config file not exists: %s", configFilePath)
		}
		f, err := os.OpenFile(configFilePath, os.O_WRONLY|os.O_CREATE, 0644)

		if err != nil {
			return fmt.Errorf("Failed to create the default config file: '%s'", configFilePath)
		}
		defer f.Close()
		enc := toml.NewEncoder(f)
		enc.Encode(ca)
	} else {
		_, err = os.Stat(configFilePath)
		if err != nil {
			return err
		}
		_, err = toml.DecodeFile(configFilePath, &ca)
		if err != nil {
			return err
		}
	}

	if ca.HubProxy != "" {
		if !strings.HasPrefix(ca.HubProxy, "http") {
			ca.HubProxy = "http://" + ca.HubProxy
		}
		_, err := url.Parse(ca.HubProxy)

		if err != nil {
			return fmt.Errorf("Failed to parse 'hub_proxy' URL")
		}
	}

	ca.SetLogLevel(ca.LogLevel)
	return addLogFileHook(ca.LogFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}
