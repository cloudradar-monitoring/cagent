package setupapi

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procSetupDiGetInfClass = modsetupapi.NewProc("SetupDiGetINFClassW")
)

// GetInfClass returns the device class name and GUID from an INF file.
// It calls the SetupDiGetINFClassW windows API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdigetinfclassw
func GetInfClass(path string) (name string, guid windows.GUID, err error) {
	if len(path)+1 >= windows.MAX_PATH {
		return name, guid, fmt.Errorf("path length exceeds the %d character limit specified by MAX_PATH: %s", windows.MAX_PATH, path)
	}

	utf16Path, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return name, guid, err
	}

	var nameBuffer [MaxClassNameLength]uint16

	r0, _, e := syscall.Syscall6(
		procSetupDiGetInfClass.Addr(),
		5,
		uintptr(unsafe.Pointer(utf16Path)),
		uintptr(unsafe.Pointer(&guid)),
		uintptr(unsafe.Pointer(&nameBuffer[0])),
		uintptr(MaxClassNameLength),
		0,
		0)

	if r0 == 0 {
		if e != 0 {
			return name, guid, syscall.Errno(e)
		}
		return name, guid, syscall.EINVAL
	}

	return syscall.UTF16ToString(nameBuffer[:]), guid, nil
}
