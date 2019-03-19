package drivertype

import "fmt"

// Value identifies a type of driver.
type Value uint32

// String returns a string representation of the driver type.
func (t Value) String() string {
	switch t {
	case ClassDriver:
		return "ClassDriver"
	case CompatDriver:
		return "CompatDriver"
	case NoDriver:
		return "NoDriver"
	default:
		return fmt.Sprintf("UnknownDriverType %d", t)
	}
}
