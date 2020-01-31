package updates

import "time"

const (
	pkgMgrNameApt = "apt-get"
	pkgMgrNameYUM = "yum"
	pkgMgrNameDNF = "dnf"
)

type pkgMgr interface {
	GetBinaryPath() string
	FetchUpdates(timeout time.Duration) error
	GetAvailableUpdatesCount() (int, *int, error)
}

func newPkgMgr(pkgMgrName string) pkgMgr {
	switch pkgMgrName {
	case pkgMgrNameApt:
		return &pkgMgrApt{}
	case pkgMgrNameYUM:
		return &pkgMgrYUM{BinaryPath: "/usr/bin/yum"}
	case pkgMgrNameDNF:
		// DNF basically just an extended edition of YUM
		// its commands and output are compatible so we will reuse the implementation:
		return &pkgMgrYUM{BinaryPath: "/usr/bin/dnf"}
	}
	panic("unsupported pkg mgr")
}
