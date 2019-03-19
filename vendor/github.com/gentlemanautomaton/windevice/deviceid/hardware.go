package deviceid

import (
	"errors"
	"fmt"
)

// Hardware is a hardware identifier. It is used to match devices to
// setup information files. A device can have more than one hardware ID.
//
// https://docs.microsoft.com/en-us/windows-hardware/drivers/install/hardware-ids
type Hardware string

// Validate returns an error if the hardware ID is not valid.
func (id Hardware) Validate() error {
	if id == "" {
		return errors.New("an empty hardware identifier was provided")
	}
	if len(id) > MaxLength {
		return fmt.Errorf("hardware identifier exceeds maximum length of %d: %s", MaxLength, id)
	}
	return nil
}
