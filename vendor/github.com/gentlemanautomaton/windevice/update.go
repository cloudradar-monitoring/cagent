package windevice

import (
	"github.com/gentlemanautomaton/windevice/deviceid"
	"github.com/gentlemanautomaton/windevice/infpath"
	"github.com/gentlemanautomaton/windevice/installflag"
	"github.com/gentlemanautomaton/windevice/newdevapi"
)

// Update attempts to update devices matching the given hardware identifier.
// The provided path must point to a valid INF file.
func Update(id deviceid.Hardware, path string, flags installflag.Value) (needReboot bool, err error) {
	// Make sure the supplied hardware ID is valid
	if err := id.Validate(); err != nil {
		return false, err
	}

	// Make sure the supplied INF file path is valid
	path, err = infpath.Prepare(path)
	if err != nil {
		return false, err
	}

	// Update the device driver
	return newdevapi.UpdateDriverForPlugAndPlayDevices(id, path, flags)
}
