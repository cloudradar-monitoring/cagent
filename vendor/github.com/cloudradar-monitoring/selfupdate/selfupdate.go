package selfupdate

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/nightlyone/lockfile"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type UpdateInfo struct {
	Version     *version.Version
	DownloadURL string
	Checksum    string
}

type Config struct {
	AppName                 string
	CurrentVersion          string
	CheckInterval           time.Duration
	UpdatesFeedURL          string
	HTTPTimeout             time.Duration
	DownloadTimeout         time.Duration
	ShutdownCallbackTimeout time.Duration
	// RequestShutdownCallback this function will be called when update will be ready to install
	// if error is returned, no update will be performed
	RequestShutdownCallback func() error
	// SigningCertificatedName is used to check for package and Feed URL webhost certificate
	// if empty, no certificate check performed
	SigningCertificatedName string
}

func (c Config) Validate() error {
	if c.AppName == "" || c.CurrentVersion == "" || c.CheckInterval == 0 || c.UpdatesFeedURL == "" ||
		c.HTTPTimeout == 0 || c.DownloadTimeout == 0 || c.ShutdownCallbackTimeout == 0 ||
		c.RequestShutdownCallback == nil {
		return fmt.Errorf("required config fields must be non-empty. %v", c)
	}

	if c.HTTPTimeout+c.DownloadTimeout+c.ShutdownCallbackTimeout >= c.CheckInterval {
		return fmt.Errorf("check interval must be greater than sum of specified timeouts")
	}

	_, err := version.NewVersion(c.CurrentVersion)
	return err
}

var config Config
var log = logrus.New()

func DefaultConfig() Config {
	return Config{
		CheckInterval:           4 * time.Hour,
		HTTPTimeout:             10 * time.Second,
		DownloadTimeout:         6 * time.Minute,
		ShutdownCallbackTimeout: 3 * time.Second,
		RequestShutdownCallback: func() error {
			return nil
		},
	}
}

func Configure(newCfg Config) error {
	if err := newCfg.Validate(); err != nil {
		return err
	}
	config = newCfg
	return nil
}

func SetLogger(l *logrus.Logger) {
	log = l
}

func StartChecking() *Updater {
	u := newUpdater()
	go u.start()
	return u
}

func ListAvailableUpdates() ([]*UpdateInfo, error) {
	updates, err := fetchUpdatesList()
	if err != nil {
		return nil, err
	}

	return filterAvailableUpdates(updates), nil
}

func filterAvailableUpdates(updates []*UpdateInfo) []*UpdateInfo {
	currVersion, _ := version.NewVersion(config.CurrentVersion)

	var availableUpdates []*UpdateInfo
	for _, u := range updates {
		if currVersion.LessThan(u.Version) {
			availableUpdates = append(availableUpdates, u)
		}
	}
	return availableUpdates
}

func fetchUpdatesList() ([]*UpdateInfo, error) {
	client := &http.Client{}
	client.Timeout = config.HTTPTimeout
	request, err := http.NewRequest("GET", config.UpdatesFeedURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = verifyWebHostCertificate(resp.TLS)
	if err != nil {
		return nil, errors.Wrap(err, "verify web-host certificate error")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	return parseFeed(resp.Body)
}

func verifyWebHostCertificate(tlsConnection *tls.ConnectionState) error {
	if config.SigningCertificatedName == "" {
		return nil
	}

	if tlsConnection == nil {
		return errors.New("connection is not secure")
	}

	if len(tlsConnection.PeerCertificates) < 1 {
		return errors.New("no certificates present")
	}

	cert := tlsConnection.PeerCertificates[0]
	if len(cert.Subject.Organization) < 1 {
		return errors.New("no organization name specified in certificate")
	}

	if cert.Subject.Organization[0] != config.SigningCertificatedName {
		return fmt.Errorf("provided name %s does not match expected %s", cert.Subject.Organization[0], config.SigningCertificatedName)
	}
	return nil
}

func parseFeed(r io.Reader) ([]*UpdateInfo, error) {
	var feed map[string]struct {
		DownloadURL string `json:"url"`
		Checksum    string `json:"checksum"`
	}

	encodedFeed, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(encodedFeed, &feed)
	if err != nil {
		return nil, err
	}

	var result []*UpdateInfo
	for versionStr, versionInfo := range feed {
		v, err := version.NewVersion(versionStr)
		if err != nil {
			log.WithError(err).Warnf("skipping invalid version: %s", versionStr)
			continue
		}
		result = append(result, &UpdateInfo{Version: v, DownloadURL: versionInfo.DownloadURL, Checksum: versionInfo.Checksum})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Version.LessThan(result[j].Version)
	})

	return result, nil
}

func DownloadAndInstallUpdate(u *UpdateInfo) error {
	lock, err := lockfile.New(filepath.Join(os.TempDir(), fmt.Sprintf("%s-self-update.lock", config.AppName)))
	if err != nil {
		return errors.Wrap(err, "could not create lock file")
	}
	err = lock.TryLock()
	if err != nil {
		return errors.Wrap(err, "could not get lock. Probably update is already running by the other process.")
	}
	defer lock.Unlock()

	tempFolder := os.TempDir()
	packageFilePath, checksum, err := downloadFile(os.TempDir(), u.DownloadURL)
	if err != nil {
		return errors.Wrapf(err, "while downloading file to folder %s", tempFolder)
	}

	if strings.ToLower(checksum) != strings.ToLower(u.Checksum) {
		return fmt.Errorf("downloaded file checksum %s does not match expected %s", checksum, u.Checksum)
	}

	if err = verifyPackageSignature(packageFilePath); err != nil {
		return errors.Wrap(err, "package signature check:")
	}

	var shutdownErr error
	shutdownDoneCh := make(chan struct{})
	go func() {
		shutdownErr = config.RequestShutdownCallback()
		shutdownDoneCh <- struct{}{}
	}()

	select {
	case <-shutdownDoneCh:
		if shutdownErr == nil {
			err = runPackageInstaller(packageFilePath)
		} else {
			err = errors.Wrap(shutdownErr, "shutdown function returned error indicating update is impossible at this moment")
		}
	case <-time.After(config.ShutdownCallbackTimeout):
		err = errors.New("Could not shutdown main process in specified timeout. Stopping update...")
	}

	return err
}
