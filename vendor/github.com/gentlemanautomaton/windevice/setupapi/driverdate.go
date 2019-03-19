package setupapi

import (
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

// DriverDate holds a device driver release date as a filetime.
type DriverDate windows.Filetime

// Value returns the driver date as a time value.
func (date DriverDate) Value() time.Time {
	v := syscall.Filetime(date)
	return time.Unix(0, v.Nanoseconds())
}

// Value returns the driver date as a time value.
func (date DriverDate) String() string {
	return date.Value().String()
}
