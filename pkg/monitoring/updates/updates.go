package updates

import (
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("package", "updates")

type Watcher struct {
	fetchTimeout    time.Duration
	checkInterval   time.Duration
	lastFetchedInfo map[string]interface{}

	lastError     error
	interruptChan chan struct{}
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

func tryDetectIfSystemRestartRequired() bool {
	_, err := os.Stat("/var/run/reboot-required")
	if err != nil && os.IsNotExist(err) {
		return false
	} else if err != nil {
		log.WithError(err).Warn("while checking if system restart required")
		return false
	}
	return true
}
