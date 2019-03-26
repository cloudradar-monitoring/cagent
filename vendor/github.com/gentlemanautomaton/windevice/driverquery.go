package windevice

import (
	"github.com/gentlemanautomaton/windevice/diflagex"
	"github.com/gentlemanautomaton/windevice/drivertype"
)

// DriverQuery holds driver query information. Its zero value is not a
// valid query.
//
// For the query to be valid, its type must specify either
// drivertype.ClassDriver or drivertype.CompatDriver.
type DriverQuery struct {
	Type    drivertype.Value
	FlagsEx diflagex.Value
}
