package windevice

import (
	"syscall"
	"time"

	"github.com/gentlemanautomaton/windevice/driverversion"
	"github.com/gentlemanautomaton/windevice/setupapi"
)

// Driver provides access to Windows driver information while executing a query.
// It can be copied by value.
//
// Driver stores a system handle internally and shouldn't be used outside of a
// query callback.
type Driver struct {
	devices syscall.Handle
	device  setupapi.DevInfoData
	data    setupapi.DrvInfoData
}

// Sys returns low-level information about the driver.
func (driver Driver) Sys() (devices syscall.Handle, device setupapi.DevInfoData, data setupapi.DrvInfoData) {
	return driver.devices, driver.device, driver.data
}

// Description returns the description of the driver.
func (driver Driver) Description() string {
	return driver.data.Description.String()
}

// ManufacturerName returns the name of the driver's manufacturer.
func (driver Driver) ManufacturerName() string {
	return driver.data.MfgName.String()
}

// ProviderName returns the name of the driver provider.
func (driver Driver) ProviderName() string {
	return driver.data.ProviderName.String()
}

// Date returns the release date of the driver.
func (driver Driver) Date() time.Time {
	return driver.data.Date.Value()
}

// Version returns the version of the driver.
func (driver Driver) Version() driverversion.Value {
	return driver.data.Version
}
