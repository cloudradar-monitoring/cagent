package setupapi

import "syscall"

// Line is a utf16 string of a predefined maximum length.
type Line [LineLength]uint16

// String returns a string representation of the line.
func (line Line) String() string {
	return syscall.UTF16ToString(line[:])
}
