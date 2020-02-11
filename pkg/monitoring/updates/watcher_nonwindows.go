// +build !windows

package updates

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/host"
)

var ErrorDisabledOnHost = fmt.Errorf("updates disabled on the host")

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

	systemRestartRequired := tryDetectIfSystemRestartRequired()
	var systemRestartRequiredInt int
	if systemRestartRequired {
		systemRestartRequiredInt = 1
	}

	results := map[string]interface{}{
		"updates_available":          totalUpgrades,
		"security_updates_available": securityUpgrades,
		"system_restart_required":    systemRestartRequiredInt,
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
