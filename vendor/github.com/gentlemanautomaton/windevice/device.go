package windevice

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/gentlemanautomaton/windevice/deviceid"
	"github.com/gentlemanautomaton/windevice/deviceproperty"
	"github.com/gentlemanautomaton/windevice/deviceregistry"
	"github.com/gentlemanautomaton/windevice/diflag"
	"github.com/gentlemanautomaton/windevice/diflagex"
	"github.com/gentlemanautomaton/windevice/difunc"
	"github.com/gentlemanautomaton/windevice/difuncremove"
	"github.com/gentlemanautomaton/windevice/drivertype"
	"github.com/gentlemanautomaton/windevice/hwprofile"
	"github.com/gentlemanautomaton/windevice/installstate"
	"github.com/gentlemanautomaton/windevice/setupapi"
)

// Device provides access to Windows device information while executing a query.
// It can be copied by value.
//
// Device stores a system handle internally and shouldn't be used outside of a
// query callback.
type Device struct {
	devices syscall.Handle
	data    setupapi.DevInfoData
}

// Sys returns low-level information about the device.
func (device Device) Sys() (devices syscall.Handle, data setupapi.DevInfoData) {
	return device.devices, device.data
}

// Drivers returns a driver set that contains drivers affiliated with the
// device.
func (device Device) Drivers(q DriverQuery) DriverSet {
	return DriverSet{
		devices: device.devices,
		device:  device.data, // TODO: Clone the data first?
		query:   q,
	}
}

// InstalledDriver returns a driver set that contains the device's currently
// installed driver.
func (device Device) InstalledDriver() DriverSet {
	return DriverSet{
		devices: device.devices,
		device:  device.data, // TODO: Clone the data first?
		query: DriverQuery{
			Type:    drivertype.ClassDriver,
			FlagsEx: diflagex.InstalledDriver | diflagex.AllowExcludedDrivers,
		},
	}
}

// Remove removes the device.
//
// When called with a global scope, all hardware profiles will be affected.
//
// When called with a config-specific scope, only the given hardware
// profile will be affected.
//
// A hardware profile of zero indicates the current hardware profile.
func (device Device) Remove(scope hwprofile.Scope, profile hwprofile.ID) (needReboot bool, err error) {
	// Prepare the removal function parameters
	difParams := difuncremove.Params{
		Header: difunc.ClassInstallHeader{
			InstallFunction: difunc.Remove,
		},
		Scope:   scope,
		Profile: profile,
	}

	// Set parameters for the class installation function call
	if err := setupapi.SetClassInstallParams(device.devices, &device.data, &difParams.Header, uint32(unsafe.Sizeof(difParams))); err != nil {
		return false, err
	}

	// Perform the removal
	if err := setupapi.CallClassInstaller(difunc.Remove, device.devices, &device.data); err != nil {
		return false, err
	}

	// Check to see whether a reboot is needed
	devParams, err := setupapi.GetDeviceInstallParams(device.devices, &device.data)
	if err != nil {
		// The device was removed but we failed to check whether a reboot is
		// needed
		return false, nil // Return success anyway because the removal succeeded
	}
	if devParams.Flags.Match(diflag.NeedReboot) || devParams.Flags.Match(diflag.NeedRestart) {
		return true, nil
	}
	return false, nil
}

// DeviceInstanceID returns the device instance ID of the device.
func (device Device) DeviceInstanceID() (deviceid.DeviceInstance, error) {
	return setupapi.GetDeviceInstanceID(device.devices, device.data)
}

// Properties returns all of the properties of the device instance.
func (device Device) Properties() ([]deviceproperty.Property, error) {
	keys, err := setupapi.GetDevicePropertyKeys(device.devices, device.data)
	if err != nil {
		return nil, err
	}

	props := make([]deviceproperty.Property, 0, len(keys))
	for i, key := range keys {
		value, err := setupapi.GetDeviceProperty(device.devices, device.data, key)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve device property %d: %v", i, err)
		}
		props = append(props, deviceproperty.Property{
			Key:   key,
			Value: value,
		})
	}

	return props, nil
}

// Description returns the description of the device.
func (device Device) Description() (string, error) {
	return setupapi.GetDeviceRegistryString(device.devices, device.data, deviceregistry.Description)
}

// HardwareID returns the set of hardware IDs associated with the device.
func (device Device) HardwareID() ([]deviceid.Hardware, error) {
	ids, err := setupapi.GetDeviceRegistryStrings(device.devices, device.data, deviceregistry.HardwareID)
	if err != nil {
		return nil, err
	}
	hids := make([]deviceid.Hardware, 0, len(ids))
	for _, id := range ids {
		hids = append(hids, deviceid.Hardware(id))
	}
	return hids, nil
}

// CompatibleID returns the set of compatible IDs associated with the device.
func (device Device) CompatibleID() ([]deviceid.Compatible, error) {
	ids, err := setupapi.GetDeviceRegistryStrings(device.devices, device.data, deviceregistry.CompatibleID)
	if err != nil {
		return nil, err
	}
	cids := make([]deviceid.Compatible, 0, len(ids))
	for _, id := range ids {
		cids = append(cids, deviceid.Compatible(id))
	}
	return cids, nil
}

// Service returns the service for the device.
func (device Device) Service() (string, error) {
	return setupapi.GetDeviceRegistryString(device.devices, device.data, deviceregistry.Service)
}

// Class returns the class name of the device.
func (device Device) Class() (string, error) {
	return setupapi.GetDeviceRegistryString(device.devices, device.data, deviceregistry.Class)
}

// ClassGUID returns a string representation of the globally unique identifier
// of the device's class.
func (device Device) ClassGUID() (string, error) {
	return setupapi.GetDeviceRegistryString(device.devices, device.data, deviceregistry.ClassGUID)
}

// ConfigFlags returns the configuration flags for the device.
func (device Device) ConfigFlags() (uint32, error) {
	return setupapi.GetDeviceRegistryUint32(device.devices, device.data, deviceregistry.ConfigFlags)
}

// DriverRegName returns the registry name of the device's driver.
func (device Device) DriverRegName() (string, error) {
	return setupapi.GetDeviceRegistryString(device.devices, device.data, deviceregistry.DriverRegName)
}

// Manufacturer returns the manufacturer of the device.
func (device Device) Manufacturer() (string, error) {
	return setupapi.GetDeviceRegistryString(device.devices, device.data, deviceregistry.Manufacturer)
}

// FriendlyName returns the friendly name of the device.
func (device Device) FriendlyName() (string, error) {
	return setupapi.GetDeviceRegistryString(device.devices, device.data, deviceregistry.FriendlyName)
}

// LocationInformation returns the location information for the device.
func (device Device) LocationInformation() (string, error) {
	return setupapi.GetDeviceRegistryString(device.devices, device.data, deviceregistry.LocationInformation)
}

// PhysicalDeviceObjectName returns the physical object name of the device.
func (device Device) PhysicalDeviceObjectName() (string, error) {
	return setupapi.GetDeviceRegistryString(device.devices, device.data, deviceregistry.PhysicalDeviceObjectName)
}

// EnumeratorName returns the name of the device's enumerator.
func (device Device) EnumeratorName() (string, error) {
	return setupapi.GetDeviceRegistryString(device.devices, device.data, deviceregistry.EnumeratorName)
}

// DevType returns the type of the device.
func (device Device) DevType() (uint32, error) {
	return setupapi.GetDeviceRegistryUint32(device.devices, device.data, deviceregistry.DevType)
}

// Characteristics returns the characteristics of the device.
func (device Device) Characteristics() (uint32, error) {
	return setupapi.GetDeviceRegistryUint32(device.devices, device.data, deviceregistry.Characteristics)
}

// InstallState returns the installation state of the device.
func (device Device) InstallState() (installstate.State, error) {
	state, err := setupapi.GetDeviceRegistryUint32(device.devices, device.data, deviceregistry.InstallState)
	return installstate.State(state), err
}

// NetCfgInstance returns the NetCfgInstance of the device.
func (device Device) NetCfgInstance() (id string, err error) {
	key, err := setupapi.OpenDevRegKey(device.devices, device.data, hwprofile.Global, 0, deviceregistry.Driver, deviceregistry.Read)
	if err != nil {
		return "", err
	}
	defer key.Close()

	id, _, err = key.GetStringValue("NetCfgInstanceId")
	return
}
