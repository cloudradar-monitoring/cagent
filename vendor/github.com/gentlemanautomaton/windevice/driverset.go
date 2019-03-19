package windevice

import (
	"io"
	"syscall"

	"github.com/gentlemanautomaton/windevice/setupapi"
)

// DriverSet accesses driver information while executing a device query.
// It can be copied by value.
//
// DriverSet stores a system handle internally and shouldn't be used outside
// of a device query callback.
type DriverSet struct {
	devices syscall.Handle
	device  setupapi.DevInfoData
	query   DriverQuery
}

// Count returns the number of devices in the set.
func (ds DriverSet) Count() (int, error) {
	var total int
	err := ds.Each(func(Driver) {
		total++
	})
	return total, err
}

// Each performs an action on each driver in the driver set.
func (ds DriverSet) Each(action DriverActor) error {
	if ds.query.FlagsEx != 0 {
		// Retrieve the parameters for the device information set
		params, err := setupapi.GetDeviceInstallParams(ds.devices, &ds.device)
		if err != nil {
			return err
		}

		// Apply the flags from the query
		params.FlagsEx |= ds.query.FlagsEx

		// Update the parameters for the device information set
		if err := setupapi.SetDeviceInstallParams(ds.devices, &ds.device, params); err != nil {
			return err
		}
	}

	err := setupapi.BuildDriverInfoList(ds.devices, &ds.device, uint32(ds.query.Type))
	if err != nil {
		return err
	}
	defer setupapi.DestroyDriverInfoList(ds.devices, &ds.device, uint32(ds.query.Type))

	i := uint32(0)
	for {
		driver, err := setupapi.EnumDriverInfo(ds.devices, &ds.device, uint32(ds.query.Type), i)
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return err
		}

		i++

		drv := Driver{
			devices: ds.devices,
			device:  ds.device,
			data:    driver,
		}

		action(drv)
	}
}
