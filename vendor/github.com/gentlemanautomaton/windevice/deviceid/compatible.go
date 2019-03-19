package deviceid

import (
	"errors"
	"fmt"
)

// Compatible is a device compatibility identifier. It has the same format
// and role as a hardware identifier, but is only used if a hardware ID
// match fails. A device can have more than one compatible ID.
//
// https://docs.microsoft.com/en-us/windows-hardware/drivers/install/compatible-ids
type Compatible string

// Validate returns an error if the compatible ID is not valid.
func (id Compatible) Validate() error {
	if id == "" {
		return errors.New("an empty compatibility identifier was provided")
	}
	if len(id) > MaxLength {
		return fmt.Errorf("compatibility identifier exceeds maximum length of %d: %s", MaxLength, id)
	}
	return nil
}
