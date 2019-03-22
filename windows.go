// +build windows

package cagent

import (
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	log "github.com/sirupsen/logrus"
)

type WindowsUpdateStatus int

const (
	WUStatusPending = iota
	WUStatusInProgress
	WUStatusCompleted
	WUStatusCompletedWithErrors
	WUStatusFailed
	WUStatusAborted
)

type WindowsUpdateWatcher struct {
	LastFetchedAt     time.Time
	LastTimeUpdatedAt time.Time
	Available         int
	Pending           int
	Err               error

	ca *Cagent
}

func windowsUpdates() (available int, pending int, lastTimeUpdated time.Time, err error) {
	start := time.Now()
	err = ole.CoInitializeEx(0, 0)
	if err != nil {
		log.Error("[Windows Updates] OLE CoInitializeEx: ", err.Error())
	}

	defer ole.CoUninitialize()
	mus, err := oleutil.CreateObject("Microsoft.Update.Session")
	if err != nil {
		log.Errorf("[Windows Updates] Failed to create Microsoft.Update.Session: %s", err.Error())
		return
	}

	defer mus.Release()
	update, err := mus.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		log.Error("[Windows Updates] Failed to create QueryInterface:: ", err.Error())
		return
	}

	defer update.Release()
	oleutil.PutProperty(update, "ClientApplicationID", "Cagent")

	us, err := oleutil.CallMethod(update, "CreateUpdateSearcher")
	if err != nil {
		log.Error("[Windows Updates] Failed CallMethod CreateUpdateSearcher: ", err.Error())
		return
	}

	usd := us.ToIDispatch()
	defer usd.Release()

	usr, err := oleutil.CallMethod(usd, "Search", "IsInstalled=0 and Type='Software' and IsHidden=0")
	if err != nil {
		log.Error("[Windows Updates] Failed CallMethod Search: ", err.Error())
		return
	}
	log.Debugf("[Windows Updates] OLE query took %.1fs", time.Since(start).Seconds())

	usrd := usr.ToIDispatch()
	defer usrd.Release()

	upd, err := oleutil.GetProperty(usrd, "Updates")
	if err != nil {
		log.Error("[Windows Updates] Failed to get Updates property: ", err.Error())
		return
	}

	updd := upd.ToIDispatch()
	defer updd.Release()

	updn, err := oleutil.GetProperty(updd, "Count")
	if err != nil {
		log.Error("[Windows Updates] Failed to get Count property: ", err.Error())
		return
	}

	available = int(updn.Val)

	thc, err := oleutil.CallMethod(usd, "GetTotalHistoryCount")
	if err != nil {
		log.Error("[Windows Updates] Failed CallMethod GetTotalHistoryCount: ", err.Error())
		return
	}

	thcn := int(thc.Val)

	// uhistRaw is list of update event records on the computer in descending chronological order
	uhistRaw, err := oleutil.CallMethod(usd, "QueryHistory", 0, thcn)
	if err != nil {
		log.Error("[Windows Updates] Failed CallMethod QueryHistory: ", err.Error())
		return
	}

	uhist := uhistRaw.ToIDispatch()
	defer uhist.Release()

	countUhist, err := oleutil.GetProperty(uhist, "Count")
	if err != nil {
		log.Error("[Windows Updates] Failed to get Count property: ", err.Error())
		return
	}

	for i := 0; i < int(countUhist.Val); i++ {
		// other available properties can be found here:
		// https://docs.microsoft.com/en-us/previous-versions/windows/desktop/aa386472(v%3dvs.85)

		itemRaw, err := oleutil.GetProperty(uhist, "Item", i)
		if err != nil {
			log.Error("[Windows Updates] Failed to fetch Windows Update history item: ", err.Error())
			continue
		}

		item := itemRaw.ToIDispatch()
		defer item.Release()

		resultCode, err := oleutil.GetProperty(item, "ResultCode")
		if err != nil {
			// On Win10 machine returns "Exception occured." after 75 updates so it looks like some undocumented internal limit.
			// We only need the last ones to found "Pending" updates so just ignore this error
			continue
		}

		updateStatus := WindowsUpdateStatus(int(resultCode.Val))
		if updateStatus == WUStatusPending {
			pending++
		}

		if lastTimeUpdated.IsZero() && updateStatus == WUStatusCompleted {
			date, err := oleutil.GetProperty(item, "Date")
			if err != nil {
				log.Warn("[Windows Updates] Failed to get Date property: ", err.Error())
				continue
			}
			if updateDate, ok := date.Value().(time.Time); ok {
				lastTimeUpdated = updateDate
			}
		}
	}
	return
}

func (ca *Cagent) WindowsUpdatesWatcher() *WindowsUpdateWatcher {
	if ca.windowsUpdateWatcher != nil {
		return ca.windowsUpdateWatcher
	}

	ca.windowsUpdateWatcher = &WindowsUpdateWatcher{ca: ca}

	go func() {
		for {
			available, pending, lastTimeUpdated, err := windowsUpdates()
			ca.windowsUpdateWatcher.LastFetchedAt = time.Now()
			ca.windowsUpdateWatcher.Err = err
			ca.windowsUpdateWatcher.Available = available
			ca.windowsUpdateWatcher.LastTimeUpdatedAt = lastTimeUpdated
			ca.windowsUpdateWatcher.Pending = pending

			time.Sleep(time.Second * time.Duration(ca.Config.WindowsUpdatesWatcherInterval))
		}
	}()

	return ca.windowsUpdateWatcher
}

func (wuw *WindowsUpdateWatcher) WindowsUpdates() (MeasurementsMap, error) {
	results := MeasurementsMap{}
	if wuw.LastFetchedAt.IsZero() {
		results["updates_available"] = nil
		results["updates_pending"] = nil
		results["last_update_timestamp"] = nil
		results["query_state"] = "pending"
		results["query_timestamp"] = nil

		return results, nil
	}

	log.Debugf("[Windows Updates] last time fetched: %.1f seconds ago", time.Since(wuw.LastFetchedAt).Seconds())
	results["query_timestamp"] = wuw.LastFetchedAt.Unix()

	if wuw.Err != nil {
		results["updates_available"] = nil
		results["updates_pending"] = nil
		results["last_update_timestamp"] = nil
		results["query_state"] = "failed"
		results["query_message"] = wuw.Err.Error()

		return results, nil
	}

	results["updates_available"] = wuw.Available
	results["updates_pending"] = wuw.Pending
	results["query_state"] = "succeeded"
	if wuw.LastTimeUpdatedAt.IsZero() {
		results["last_update_timestamp"] = nil
	} else {
		results["last_update_timestamp"] = wuw.LastTimeUpdatedAt.Unix()
	}

	return results, nil
}
