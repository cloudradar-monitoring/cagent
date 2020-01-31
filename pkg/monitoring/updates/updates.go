package updates

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/host"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("package", "updates")

type Watcher struct {
	fetchTimeout    time.Duration
	checkInterval   time.Duration
	lastFetchedInfo map[string]interface{}
	lastError       error
	interruptChan   chan struct{}
}

var watcher *Watcher

func Shutdown() {
	if watcher != nil {
		watcher.Shutdown()
	}
}

func GetWatcher(fetchTimeout, checkInterval uint) *Watcher {
	if watcher != nil {
		return watcher
	}

	watcher = &Watcher{
		fetchTimeout:    time.Duration(uint(time.Second) * fetchTimeout),
		checkInterval:   time.Duration(uint(time.Second) * checkInterval),
		lastFetchedInfo: nil,
		interruptChan:   make(chan struct{}),
	}
	go watcher.Run()
	return watcher
}

func (w *Watcher) GetSystemUpdatesInfo() (map[string]interface{}, error) {
	return w.lastFetchedInfo, w.lastError
}

func (w *Watcher) Run() {
	for {
		w.lastFetchedInfo, w.lastError = w.tryFetchAndParseUpdatesInfo()
		select {
		case <-w.interruptChan:
			return
		case <-time.After(w.checkInterval):
			continue
		}
	}
}

func (w *Watcher) tryFetchAndParseUpdatesInfo() (map[string]interface{}, error) {
	_, family, version, err := host.PlatformInformation()
	if err != nil {
		return nil, errors.Wrap(err, "while detecting OS platform")
	}

	pkgMgrName, err := tryDetectPackageManager(family, version)
	if err != nil {
		log.WithError(err).Warn()
		// unknown or unsupported pkg manager is not an error
		return nil, nil
	}

	mgr := newPkgMgr(pkgMgrName)
	pkgMgrPath := mgr.GetBinaryPath()
	if _, err := os.Stat(pkgMgrPath); err != nil {
		log.WithError(err).Warnf("expected to find package manager binary at %s", pkgMgrPath)
		return nil, nil
	}

	err = mgr.FetchUpdates(w.fetchTimeout)
	if err != nil {
		return nil, err
	}

	totalUpgrades, securityUpgrades, err := mgr.GetAvailableUpdatesCount()
	if err != nil {
		return nil, err
	}

	results := map[string]interface{}{
		"updates_available":          totalUpgrades,
		"security_updates_available": securityUpgrades,
	}
	return results, nil
}

func tryDetectPackageManager(osFamily, osVersion string) (string, error) {
	if osFamily == "debian" {
		return pkgMgrNameApt, nil
	}

	majorVersion := tryParseMajorVersion(osVersion)
	switch osFamily {
	case "fedora":
		if majorVersion < 22 {
			return pkgMgrNameYUM, nil
		}
		return pkgMgrNameDNF, nil
	case "rhel":
		if majorVersion < 8 {
			return pkgMgrNameYUM, nil
		}
		return pkgMgrNameDNF, nil
	}

	return "", fmt.Errorf("can't detect package manager for OS family %s version %s", osFamily, osVersion)
}

func (w *Watcher) Shutdown() {
	w.interruptChan <- struct{}{}
}

// tryParseMajorVersion returns -1 if it wasn't able to parse version
func tryParseMajorVersion(versionStr string) int {
	versionStr = strings.TrimSpace(versionStr)
	versionStr = strings.TrimPrefix(versionStr, "v")

	majorStr := ""
	for _, r := range versionStr {
		if !unicode.IsDigit(r) {
			break
		}
		majorStr += string(r)
	}

	majorVer, err := strconv.Atoi(majorStr)
	if err != nil {
		return -1
	}
	return majorVer
}
