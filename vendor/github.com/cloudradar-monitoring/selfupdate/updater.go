package selfupdate

import "time"

const firstCheckSleepInterval = 2 * time.Minute

type Updater struct {
	interrupt chan struct{}
}

func newUpdater() *Updater {
	return &Updater{
		interrupt: make(chan struct{}, 1),
	}
}

func (u *Updater) start() {
	time.Sleep(firstCheckSleepInterval)
	for {
		updates, err := ListAvailableUpdates()
		if err != nil {
			log.WithError(err).Warn("while checking for updates")
		}

		if len(updates) > 0 {
			latestVersion := updates[len(updates)-1]
			log.Warnf("new update available %s. Triggering installation...", latestVersion.Version.Original())
			err = DownloadAndInstallUpdate(latestVersion)
			if err != nil {
				log.WithError(err).Warnf("while trying to install new version %s", latestVersion.Version.Original())
			} else {
				// stop checking
				return
			}
		} else {
			log.Debug("no new versions available")
		}

		select {
		case <-u.interrupt:
			return
		case <-time.After(config.CheckInterval):
			continue
		}
	}
}

func (u *Updater) Shutdown() {
	u.interrupt <- struct{}{}
}
