package setupapi

import "syscall"

// Path is a utf16 string of a predefined maximum length.
type Path [syscall.MAX_PATH]uint16

// String returns a string representation of the path.
func (path Path) String() string {
	return syscall.UTF16ToString(path[:])
}
