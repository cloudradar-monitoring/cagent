// +build windows

package winapi

import (
	"unsafe"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("package", "winapi")

func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}
