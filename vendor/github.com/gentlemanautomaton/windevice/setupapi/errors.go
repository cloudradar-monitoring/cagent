package setupapi

import (
	"errors"
	"syscall"
)

var (
	// ErrEmptyBuffer is returned when a nil or zero-sized buffer is provided
	// to a system call.
	ErrEmptyBuffer = errors.New("nil or empty buffer provided")

	// ErrInvalidRegistry is returned when an unexpected registry value type
	// is encountered.
	//ErrInvalidRegistry = errors.New("invalid registry type")

	// ErrInvalidData is returned when a property isn't present or isn't valid.
	ErrInvalidData = syscall.Errno(13) // ERROR_INVALID_DATA

	// ErrInvalidClass indicates that an invalid class was specified.
	ErrInvalidClass = syscall.Errno(0xE0000209) // ERROR_INVALID_CLASS

	// ErrNoDriverSelected is returned when a device doesn't have a driver
	// affiliated with it.
	//ErrNoDriverSelected = syscall.Errno(3758096899)
)
