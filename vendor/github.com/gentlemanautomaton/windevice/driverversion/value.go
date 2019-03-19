package driverversion

import "fmt"

// https://docs.microsoft.com/en-us/windows/desktop/direct3d9/d3dadapter-identifier9

// Value holds a device driver version in an unsigned 64-bit integer.
type Value uint64

// Product returns the product number.
func (v Value) Product() uint16 {
	return uint16(v >> 48)
}

// Version returns the version number.
func (v Value) Version() uint16 {
	return uint16(v >> 32)
}

// SubVersion returns the subversion number.
func (v Value) SubVersion() uint16 {
	return uint16(v >> 16)
}

// Build returns the build number.
func (v Value) Build() uint16 {
	return uint16(v)
}

// String returns a string representation of v.
func (v Value) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Product(), v.Version(), v.SubVersion(), v.Build())
}
