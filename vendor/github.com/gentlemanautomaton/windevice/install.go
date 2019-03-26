package windevice

import (
	"fmt"

	"github.com/gentlemanautomaton/windevice/devicecreation"
	"github.com/gentlemanautomaton/windevice/deviceid"
	"github.com/gentlemanautomaton/windevice/deviceregistry"
	"github.com/gentlemanautomaton/windevice/difunc"
	"github.com/gentlemanautomaton/windevice/infpath"
	"github.com/gentlemanautomaton/windevice/installflag"
	"github.com/gentlemanautomaton/windevice/newdevapi"
	"github.com/gentlemanautomaton/windevice/setupapi"
)

// Install attempts to install a device with the provided hardware identifier,
// setup information file and description.
func Install(id deviceid.Hardware, path, description string, flags installflag.Value) (instance deviceid.DeviceInstance, needReboot bool, err error) {
	// Make sure the supplied hardware ID is valid
	if err := id.Validate(); err != nil {
		return "", false, err
	}

	// Make sure the supplied INF file path is valid
	path, err = infpath.Prepare(path)
	if err != nil {
		return "", false, err
	}

	// Ask windows to retrieve the name and GUID from the INF file
	name, guid, err := setupapi.GetInfClass(path)
	if err != nil {
		return "", false, fmt.Errorf("failed to read inf file: %v", err)
	}

	// Prepare a fresh device information set
	devices, err := setupapi.CreateDeviceInfoList(&guid)
	if err != nil {
		return "", false, fmt.Errorf("failed to create device information set: %v", err)
	}
	defer setupapi.DestroyDeviceInfoList(devices)

	// Add a new device to the set
	device, err := setupapi.CreateDeviceInfo(devices, name, guid, description, devicecreation.GenerateID)
	if err != nil {
		return "", false, fmt.Errorf("failed to create device: %v", err)
	}

	// Set the hardware ID
	if err := setupapi.SetDeviceRegistryStrings(devices, device, deviceregistry.HardwareID, []string{string(id)}); err != nil {
		return "", false, fmt.Errorf("failed to set hardware ID: %v", err)
	}

	// Register the device
	if err := setupapi.CallClassInstaller(difunc.RegisterDevice, devices, &device); err != nil {
		return "", false, fmt.Errorf("failed to register device: %v", err)
	}

	// Update the device driver
	needReboot, err = newdevapi.UpdateDriverForPlugAndPlayDevices(id, path, flags)
	if err != nil {
		return "", needReboot, err
	}

	// Get the device instance ID of the new device
	instance, err = setupapi.GetDeviceInstanceID(devices, device)

	return instance, needReboot, err
}
