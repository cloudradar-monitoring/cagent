package newdevapi

import (
	"syscall"
	"unsafe"

	"github.com/gentlemanautomaton/windevice/deviceid"

	"github.com/gentlemanautomaton/windevice/installflag"
	"golang.org/x/sys/windows"
)

var (
	modnewdev = windows.NewLazySystemDLL("newdev.dll")

	procUpdateDriverForPlugAndPlayDevices = modnewdev.NewProc("UpdateDriverForPlugAndPlayDevicesW")
)

// UpdateDriverForPlugAndPlayDevices updates a driver for a device using
// the provided INF file. It calls the UpdateDriverForPlugAndPlayDevices
// windows API function.
//
// The device to be updated is identified by the provided hardware ID.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/newdev/nf-newdev-updatedriverforplugandplaydevicesw
func UpdateDriverForPlugAndPlayDevices(id deviceid.Hardware, path string, flags installflag.Value) (needReboot bool, err error) {
	utf16HardwareID, err := syscall.UTF16PtrFromString(string(id))
	if err != nil {
		return
	}

	utf16Path, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return
	}

	r0, _, e := syscall.Syscall6(
		procUpdateDriverForPlugAndPlayDevices.Addr(),
		5,
		0,
		uintptr(unsafe.Pointer(utf16HardwareID)),
		uintptr(unsafe.Pointer(utf16Path)),
		uintptr(flags),
		uintptr(unsafe.Pointer(&needReboot)),
		0)

	if r0 == 0 {
		if e != 0 {
			err = syscall.Errno(e)
		}
		err = syscall.EINVAL
	}
	return
}
