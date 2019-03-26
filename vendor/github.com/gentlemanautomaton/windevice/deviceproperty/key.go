package deviceproperty

import (
	"strconv"

	"github.com/gentlemanautomaton/winguid"
	"golang.org/x/sys/windows"
)

// Key is a device property key that identifies a particular device property.
type Key struct {
	Category   windows.GUID
	PropertyID uint32
}

// String returns a string representation of the key.
func (k Key) String() string {
	if name := k.Name(); name != "" {
		return name
	}
	return winguid.String(k.Category) + "." + strconv.Itoa(int(k.PropertyID))
}

// Name returns the name of the device property if it's known.
func (k Key) Name() string {
	return KnownKeys[k]
}
