package deviceid

import (
	"errors"
	"fmt"
)

// DeviceInstance is a device instance identifier. It is assigned by the
// operating system to device nodes and is unique to the node. It is formed
// by combining a device identifier with an instance identifier.
//
// https://docs.microsoft.com/en-us/windows-hardware/drivers/install/device-instance-ids
type DeviceInstance string

// Validate returns an error if the device instance ID is not valid.
func (id DeviceInstance) Validate() error {
	if id == "" {
		return errors.New("an empty device instance identifier was provided")
	}
	if len(id) > MaxLength {
		return fmt.Errorf("device instance identifier exceeds maximum length of %d: %s", MaxLength, id)
	}
	return nil
}
