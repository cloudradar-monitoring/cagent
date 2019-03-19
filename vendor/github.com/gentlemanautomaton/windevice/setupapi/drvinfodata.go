package setupapi

import (
	"github.com/gentlemanautomaton/windevice/drivertype"
	"github.com/gentlemanautomaton/windevice/driverversion"
)

// DrvInfoData holds driver information.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/ns-setupapi-_sp_drvinfo_data_v2_w
type DrvInfoData struct {
	Size         uint32
	Type         drivertype.Value
	reserved     uintptr
	Description  Line
	MfgName      Line
	ProviderName Line
	Date         DriverDate
	Version      driverversion.Value
}
