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
	LastFetchedAt time.Time
	Available     int
	Pending       int
	Err           error

	ca *Cagent
}

func windowsUpdates() (available int, pending int, err error) {
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
		itemRaw, err := oleutil.GetProperty(uhist, "Item", i)
		if err != nil {
			log.Error("[Windows] Failed to fetch Windows Update history item: ", err.Error())
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

		if WindowsUpdateStatus(int(resultCode.Val)) == WUStatusPending {
			pending++
		}
	}
	return
}

func (ca *Cagent) WindowsUpdatesWatcher() *WindowsUpdateWatcher {
	wuw := &WindowsUpdateWatcher{ca: ca}

	go func() {
		for {
			available, pending, err := windowsUpdates()
			wuw.LastFetchedAt = time.Now()
			wuw.Err = err
			wuw.Available = available
			wuw.Pending = pending

			time.Sleep(time.Second * time.Duration(ca.WindowsUpdatesWatcherInterval))
		}
	}()

	return wuw
}

func (wuw *WindowsUpdateWatcher) WindowsUpdates() (MeasurementsMap, error) {
	results := MeasurementsMap{}
	if wuw.LastFetchedAt.IsZero() {
		results["updates_available"] = nil
		results["updates_pending"] = nil
		results["query_state"] = "pending"
		results["query_timestamp"] = nil

		return results, nil
	}

	log.Debugf("[Windows Updates] last time fetched: %.1f seconds ago", time.Since(wuw.LastFetchedAt).Seconds())
	results["query_timestamp"] = wuw.LastFetchedAt.Unix()

	if wuw.Err != nil {
		results["updates_available"] = nil
		results["updates_pending"] = nil
		results["query_state"] = "failed"
		results["query_message"] = wuw.Err.Error()

		return results, nil
	}

	results["updates_available"] = wuw.Available
	results["updates_pending"] = wuw.Pending
	results["query_state"] = "succeeded"

	return results, nil
}
