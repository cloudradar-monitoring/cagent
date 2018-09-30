// +build windows

package cagent

import (
	"fmt"
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

func (ca *Cagent) WindowsUpdates() (MeasurementsMap, error) {
	results := MeasurementsMap{}
	start := time.Now()
	err := ole.CoInitializeEx(0, 0)
	if err != nil {
		log.Error("OLE CoInitializeEx: ", err.Error())
	}
	defer ole.CoUninitialize()
	mus, err := oleutil.CreateObject("Microsoft.Update.Session")
	if err != nil {
		return nil, fmt.Errorf("Windows Updates: Failed to create Microsoft.Update.Session: %s", err.Error())
	}

	defer mus.Release()
	update, err := mus.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, fmt.Errorf("Windows Updates: Failed to create QueryInterface: %s", err.Error())
	}

	defer update.Release()
	oleutil.PutProperty(update, "ClientApplicationID", "Cagent")

	us, err := oleutil.CallMethod(update, "CreateUpdateSearcher")
	if err != nil {
		return nil, fmt.Errorf("Windows Updates: Failed CallMethod CreateUpdateSearcher: %s", err.Error())
	}

	usd := us.ToIDispatch()
	defer usd.Release()

	usr, err := oleutil.CallMethod(usd, "Search", "IsInstalled=0 and Type='Software' and IsHidden=0")
	if err != nil {
		return nil, fmt.Errorf("Windows Updates: Failed CallMethod Search: %s", err.Error())
	}
	log.Debugf("[Windows Updates] OLE query took %.1fs", time.Since(start).Seconds())

	usrd := usr.ToIDispatch()
	defer usrd.Release()

	upd, err := oleutil.GetProperty(usrd, "Updates")
	if err != nil {
		return nil, fmt.Errorf("Windows Updates: Failed to get Updates property: %s", err.Error())
	}

	updd := upd.ToIDispatch()
	defer updd.Release()

	updn, err := oleutil.GetProperty(updd, "Count")
	if err != nil {
		return nil, fmt.Errorf("Windows Updates: Failed to get Count property: %s", err.Error())
	}

	results["updates_available"] = updn.Val

	thc, err := oleutil.CallMethod(usd, "GetTotalHistoryCount")
	if err != nil {
		return nil, fmt.Errorf("Windows Updates: Failed CallMethod GetTotalHistoryCount: %s", err.Error())
	}

	thcn := int(thc.Val)

	uhistRaw, err := oleutil.CallMethod(usd, "QueryHistory", 0, thcn)
	if err != nil {
		return nil, fmt.Errorf("Windows Updates: Failed CallMethod QueryHistory: %s", err.Error())
	}

	uhist := uhistRaw.ToIDispatch()
	defer uhist.Release()

	countUhist, err := oleutil.GetProperty(uhist, "Count")
	if err != nil {
		return nil, fmt.Errorf("Windows Updates: Failed to get Count property: %s", err.Error())
	}

	pendingCount := 0
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
			pendingCount++
		}
	}

	results["updates_needs_reboot"] = pendingCount

	return results, nil
}
