// +build windows

package updates

import (
	"fmt"
	"sync"
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

var ErrorDisabledOnHost = fmt.Errorf("windows updates disabled on the host")
var ErrorPreviousQueryStillInProcess = fmt.Errorf("previous request still in process")
type WindowsUpdateStatus int

const (
	WUStatusPending = iota
	WUStatusInProgress
	WUStatusCompleted
	WUStatusCompletedWithErrors
	WUStatusFailed
	WUStatusAborted
)

type SecondsAgo struct {
	since time.Time
}

func (t SecondsAgo) MarshalJSON() ([]byte, error) {
	return t.MarshalText()
}

func (t SecondsAgo) MarshalText() ([]byte, error) {
	if t.since.IsZero() {
		return nil, fmt.Errorf("since time not set")
	}

	return []byte(fmt.Sprintf("%.0f", time.Since(t.since).Seconds())), nil
}

var queryInProccess bool
var queryInProccessMutex sync.Mutex

func (w *Watcher) tryFetchAndParseUpdatesInfo() (results map[string]interface{}, err error) {
	available, pending, updated, err := w.query()
	if err != nil {
		if err == ErrorPreviousQueryStillInProcess {
			// do not return any error in this case
			// because we can't control the timeout on Windows
			// so lets skip the new request
			return nil, nil
		}

		return map[string]interface{}{
			"updates_available":              nil,
			"updates_pending":                nil,
			"last_updates_installed_ago_sec": nil,
			"last_query_age_sec":             nil,
			"query_state":                    "failed",
		}, err
	}

	return map[string]interface{}{
		"updates_available":              available,
		"updates_pending":                pending,
		"last_updates_installed_ago_sec": SecondsAgo{updated},
		"last_query_age_sec":             SecondsAgo{time.Now()},
		"query_state":                    "succeeded",
	}, nil
}

func (w *Watcher) query() (available int, pending int, lastTimeUpdated time.Time, err error) {
	queryInProccessMutex.Lock()
	if queryInProccess {
		queryInProccessMutex.Unlock()
		err = ErrorPreviousQueryStillInProcess
		return
	}
	start := time.Now()

	queryInProccess = true
	defer func() {
		queryInProccess = false
	}()
	queryInProccessMutex.Unlock()

	err = ole.CoInitializeEx(0, 0)
	if err != nil {
		log.Error("[Windows Updates] OLE CoInitializeEx: ", err.Error())
		return
	}

	defer ole.CoUninitialize()
	mus, err2 := oleutil.CreateObject("Microsoft.Update.Session")
	if err2 != nil {
		err = err2
		log.Errorf("[Windows Updates] Failed to create Microsoft.Update.Session: %s", err.Error())
		return
	}

	defer mus.Release()
	update, err2 := mus.QueryInterface(ole.IID_IDispatch)
	if err2 != nil {
		err = err2
		log.Error("[Windows Updates] Failed to create QueryInterface:: ", err.Error())
		return
	}

	defer update.Release()
	oleutil.PutProperty(update, "ClientApplicationID", "Cagent")

	us, err2 := oleutil.CallMethod(update, "CreateUpdateSearcher")
	if err2 != nil {
		err = err2
		log.Error("[Windows Updates] Failed CallMethod CreateUpdateSearcher: ", err.Error())
		return
	}

	ush := us.ToIDispatch()
	defer ush.Release()
	// lets use the fast local-only query to check if WindowsUpdates service is enabled on the host
	_, err2 = oleutil.CallMethod(ush, "GetTotalHistoryCount")
	if err2 != nil {
		// that means Windows Updates service is disabled
		err = ErrorDisabledOnHost
		log.Warningf("[Windows Updates] Windows Updates service is disabled: ", err.Error())
		return
	}

	usd := us.ToIDispatch()
	defer usd.Release()
	usr, err2 := oleutil.CallMethod(usd, "Search", "IsInstalled=0 and Type='Software' and IsHidden=0")
	if err2 != nil {
		err = err2
		log.Error("[Windows Updates] Failed to query Windows Updates: ", err.Error())
		return
	}
	log.Debugf("[Windows Updates] OLE query took %.1fs", time.Since(start).Seconds())

	usrd := usr.ToIDispatch()
	defer usrd.Release()

	upd, err2 := oleutil.GetProperty(usrd, "Updates")
	if err2 != nil {
		err = err2
		log.Error("[Windows Updates] Failed to get Updates property: ", err.Error())
		return
	}

	updd := upd.ToIDispatch()
	defer updd.Release()

	updn, err2 := oleutil.GetProperty(updd, "Count")
	if err2 != nil {
		err = err2
		log.Error("[Windows Updates] Failed to get Count property: ", err.Error())
		return
	}

	available = int(updn.Val)
	pending = 0

	thc, err2 := oleutil.CallMethod(usd, "GetTotalHistoryCount")
	if err2 != nil {
		err = err2
		log.Error("[Windows Updates] Failed CallMethod GetTotalHistoryCount: ", err.Error())
		return
	}

	thcn := int(thc.Val)

	// uhistRaw is list of update event records on the computer in descending chronological order
	uhistRaw, err2 := oleutil.CallMethod(usd, "QueryHistory", 0, thcn)
	if err2 != nil {
		err = err2
		log.Error("[Windows Updates] Failed CallMethod QueryHistory: ", err.Error())
		return
	}

	uhist := uhistRaw.ToIDispatch()
	defer uhist.Release()

	countUhist, err2 := oleutil.GetProperty(uhist, "Count")
	if err2 != nil {
		err = err2
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
			// On Win10 machine returns "Exception occurred." after 75 updates so it looks like some undocumented internal limit.
			// We only need the last ones to found "Pending" updates so just ignore this error
			continue
		}

		updateStatus := WindowsUpdateStatus(int(resultCode.Val))
		if updateStatus == WUStatusPending {
			pending++
		}

		if updateStatus == WUStatusCompleted {
			date, err := oleutil.GetProperty(item, "Date")
			if err != nil {
				log.Warn("[Windows Updates] Failed to get Date property: ", err.Error())
				continue
			}
			if updateDate, ok := date.Value().(time.Time); ok {
				if lastTimeUpdated.IsZero() || updateDate.After(lastTimeUpdated) {
					lastTimeUpdated = updateDate
				}
			}
		}
	}

	return
}
