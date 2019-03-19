package deviceid

import (
	"errors"
	"fmt"
)

// Device is a device identifier. It is assigned by a device enumerator
// to a device.
//
// https://docs.microsoft.com/en-us/windows-hardware/drivers/install/device-ids
type Device string

// Validate returns an error if the device ID is not valid.
func (id Device) Validate() error {
	if id == "" {
		return errors.New("an empty device identifier was provided")
	}
	if len(id) > MaxLength {
		return fmt.Errorf("device identifier exceeds maximum length of %d: %s", MaxLength, id)
	}
	return nil
}
