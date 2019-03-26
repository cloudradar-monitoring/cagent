package setupapi

import (
	"syscall"
	"unsafe"

	"github.com/gentlemanautomaton/windevice/difunc"
	"golang.org/x/sys/windows"
)

var (
	procSetupDiClassGuidsFromNameEx  = modsetupapi.NewProc("SetupDiClassGuidsFromNameExW")
	procSetupDiSetClassInstallParams = modsetupapi.NewProc("SetupDiSetClassInstallParamsW")
	procSetupDiCallClassInstaller    = modsetupapi.NewProc("SetupDiCallClassInstaller")
)

// ClassGuidsFromNameEx returns the list of GUIDs associated with a class
// name. It calls the SetupDiClassGuidsFromNameEx windows API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdiclassguidsfromnameexw
func ClassGuidsFromNameEx(className, machine string) (guids []windows.GUID, err error) {
	cp, err := syscall.UTF16PtrFromString(className)
	if err != nil {
		return nil, err
	}

	var mp *uint16
	if machine != "" {
		mp, err = syscall.UTF16PtrFromString(machine)
		if err != nil {
			return nil, err
		}
	}

	guids = make([]windows.GUID, 1)

	// Make up to 3 attempts to get the class data.
	const rounds = 3
	for i := 0; i < rounds; i++ {
		var length uint32
		length, err = classGuidsFromNameEx(cp, mp, guids)
		if err == nil {
			if length == 0 {
				return nil, nil
			}
			return guids, nil
		}
		if err == syscall.ERROR_INSUFFICIENT_BUFFER && i < rounds {
			guids = make([]windows.GUID, length)
		} else {
			return nil, err
		}
	}

	return nil, syscall.ERROR_INSUFFICIENT_BUFFER
}

func classGuidsFromNameEx(className, machine *uint16, buffer []windows.GUID) (reqSize uint32, err error) {
	var gp *windows.GUID
	if len(buffer) > 0 {
		gp = &buffer[0]
	}
	r0, _, e := syscall.Syscall6(
		procSetupDiClassGuidsFromNameEx.Addr(),
		6,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(gp)),
		uintptr(len(buffer)),
		uintptr(unsafe.Pointer(&reqSize)),
		uintptr(unsafe.Pointer(machine)),
		0)
	if r0 == 0 {
		if e != 0 {
			err = syscall.Errno(e)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// SetClassInstallParams updates the class installation parameters for a
// device or device information set. It calls the SetupDiSetClassInstallParams
// windows API function.
//
// If a header is provided, it must be embedded within a params struct that
// is appropriate for the device installation function specified in the
// header.
//
// If a non-zero size is provided, it must be the size of the enclosing params
// struct that contains the header. If a size of 0 is provided the
// installation parameters of the class will be cleared.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdisetdeviceinstallparamsw
func SetClassInstallParams(devices syscall.Handle, device *DevInfoData, header *difunc.ClassInstallHeader, size uint32) error {
	header.Size = uint32(unsafe.Sizeof(*header))

	r0, _, e := syscall.Syscall6(
		procSetupDiSetClassInstallParams.Addr(),
		4,
		uintptr(devices),
		uintptr(unsafe.Pointer(device)),
		uintptr(unsafe.Pointer(header)),
		uintptr(size),
		0,
		0)

	if r0 == 0 {
		if e != 0 {
			return syscall.Errno(e)
		}
		return syscall.EINVAL
	}
	return nil
}

// CallClassInstaller invokes a class installer function for a device.
// It calls the SetupDiCallClassInstaller windows API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdicallclassinstaller
func CallClassInstaller(function difunc.Function, devices syscall.Handle, device *DevInfoData) error {
	r0, _, e := syscall.Syscall(
		procSetupDiCallClassInstaller.Addr(),
		3,
		uintptr(function),
		uintptr(devices),
		uintptr(unsafe.Pointer(device)))

	if r0 == 0 {
		if e != 0 {
			return syscall.Errno(e)
		}
		return syscall.EINVAL
	}
	return nil
}
