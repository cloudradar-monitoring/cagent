package deviceid

import (
	"errors"
	"fmt"
)

// Instance is an instance identifier. It is used distinguish multiple instances
// a device.
//
// https://docs.microsoft.com/en-us/windows-hardware/drivers/install/instance-ids
type Instance string

// Validate returns an error if the instance ID is not valid.
func (id Instance) Validate() error {
	if id == "" {
		return errors.New("an empty instance identifier was provided")
	}
	if len(id) > MaxLength {
		return fmt.Errorf("instance identifier exceeds maximum length of %d: %s", MaxLength, id)
	}
	return nil
}
