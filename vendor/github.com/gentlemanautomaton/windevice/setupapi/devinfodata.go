package setupapi

import "golang.org/x/sys/windows"

// DevInfoData holds device information.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/ns-setupapi-sp_devinfo_data
type DevInfoData struct {
	Size     uint32
	GUID     windows.GUID
	DevInst  uint32
	reserved uintptr
}
