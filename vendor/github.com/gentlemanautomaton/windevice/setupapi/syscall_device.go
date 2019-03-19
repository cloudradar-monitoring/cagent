package setupapi

import (
	"io"
	"syscall"
	"unsafe"

	"github.com/gentlemanautomaton/windevice/deviceclass"
	"github.com/gentlemanautomaton/windevice/devicecreation"
	"github.com/gentlemanautomaton/windevice/deviceid"
	"golang.org/x/sys/windows"
)

var (
	procSetupDiGetClassDevsExW        = modsetupapi.NewProc("SetupDiGetClassDevsExW")
	procSetupDiCreateDeviceInfoList   = modsetupapi.NewProc("SetupDiCreateDeviceInfoList")
	procSetupDiDestroyDeviceInfoList  = modsetupapi.NewProc("SetupDiDestroyDeviceInfoList")
	procSetupDiEnumDeviceInfo         = modsetupapi.NewProc("SetupDiEnumDeviceInfo")
	procSetupDiGetDeviceInstallParams = modsetupapi.NewProc("SetupDiGetDeviceInstallParamsW")
	procSetupDiSetDeviceInstallParams = modsetupapi.NewProc("SetupDiSetDeviceInstallParamsW")
	procSetupDiGetDeviceInstanceID    = modsetupapi.NewProc("SetupDiGetDeviceInstanceIdW")
	procSetupDiCreateDeviceInfo       = modsetupapi.NewProc("SetupDiCreateDeviceInfoW")
)

// GetClassDevsEx builds and returns a device information list that contains
// devices matching the given parameters. It calls the SetupDiGetClassDevsEx
// windows API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdigetclassdevsexw
func GetClassDevsEx(guid *windows.GUID, enumerator string, flags uint32, devices syscall.Handle, machineName string) (handle syscall.Handle, err error) {
	var ep *uint16
	if enumerator != "" {
		ep, err = syscall.UTF16PtrFromString(enumerator)
		if err != nil {
			return syscall.InvalidHandle, err
		}
	}

	var mnp *uint16
	if machineName != "" {
		mnp, err = syscall.UTF16PtrFromString(machineName)
		if err != nil {
			return syscall.InvalidHandle, err
		}
	}

	if guid == nil {
		flags |= deviceclass.AllClasses
	}

	r0, _, e := syscall.Syscall9(
		procSetupDiGetClassDevsExW.Addr(),
		7,
		uintptr(unsafe.Pointer(guid)),
		uintptr(unsafe.Pointer(ep)),
		0, // hwndParent
		uintptr(flags),
		uintptr(devices),
		uintptr(unsafe.Pointer(mnp)),
		0,
		0,
		0)
	handle = syscall.Handle(r0)
	if handle == syscall.InvalidHandle {
		if e != 0 {
			err = syscall.Errno(e)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// CreateDeviceInfoList creates an empty device information list. It calls the
// SetupDiCreateDeviceInfoList windows API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdicreatedeviceinfolist
func CreateDeviceInfoList(guid *windows.GUID) (handle syscall.Handle, err error) {
	r0, _, e := syscall.Syscall(
		procSetupDiCreateDeviceInfoList.Addr(),
		2,
		uintptr(unsafe.Pointer(guid)),
		0,
		0)
	handle = syscall.Handle(r0)
	if handle == syscall.InvalidHandle {
		if e != 0 {
			err = syscall.Errno(e)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// DestroyDeviceInfoList destroys a device information list. It calls the
// SetupDiDestroyDeviceInfoList windows API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdidestroydeviceinfolist
func DestroyDeviceInfoList(devices syscall.Handle) error {
	r0, _, e := syscall.Syscall(
		procSetupDiDestroyDeviceInfoList.Addr(),
		1,
		uintptr(devices),
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

// EnumDeviceInfo returns information about a device in a device information
// list. It calls the SetupDiEnumDeviceInfo windows API function.
//
// EnumDeviceInfo returns io.EOF when there are no more members in the list.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdienumdeviceinfo
func EnumDeviceInfo(devices syscall.Handle, index uint32) (info DevInfoData, err error) {
	const errNoMoreItems = 259

	info.Size = uint32(unsafe.Sizeof(info))

	r0, _, e := syscall.Syscall(
		procSetupDiEnumDeviceInfo.Addr(),
		3,
		uintptr(devices),
		uintptr(index),
		uintptr(unsafe.Pointer(&info)))
	if r0 == 0 {
		switch e {
		case 0:
			err = syscall.EINVAL
		case errNoMoreItems:
			err = io.EOF
		default:
			err = syscall.Errno(e)
		}
	}
	return
}

// CreateDeviceInfo creates a new device and adds it to the device
// information set. It calls the SetupDiCreateDeviceInfoW windows
// API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdicreatedeviceinfow
func CreateDeviceInfo(devices syscall.Handle, name string, class windows.GUID, description string, flags devicecreation.Flags) (device DevInfoData, err error) {
	utf16Name, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return DevInfoData{}, err
	}

	var utf16Description *uint16
	if description != "" {
		utf16Description, err = syscall.UTF16PtrFromString(description)
		if err != nil {
			return DevInfoData{}, err
		}
	}

	device.Size = uint32(unsafe.Sizeof(device))

	r0, _, e := syscall.Syscall9(
		procSetupDiCreateDeviceInfo.Addr(),
		7,
		uintptr(devices),
		uintptr(unsafe.Pointer(utf16Name)),
		uintptr(unsafe.Pointer(&class)),
		uintptr(unsafe.Pointer(utf16Description)),
		0,
		uintptr(flags),
		uintptr(unsafe.Pointer(&device)),
		0,
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

// GetDeviceInstanceID returns the device instance ID for a device.
// It calls the SetupDiGetDeviceInstanceIdW windows API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdigetdeviceinstanceidw
func GetDeviceInstanceID(devices syscall.Handle, device DevInfoData) (id deviceid.DeviceInstance, err error) {
	const maxLength = deviceid.MaxLength + 1 // Accomodate null terminator
	var (
		buffer [maxLength]uint16
		length uint32
	)

	r0, _, e := syscall.Syscall6(
		procSetupDiGetDeviceInstanceID.Addr(),
		5,
		uintptr(devices),
		uintptr(unsafe.Pointer(&device)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
		uintptr(unsafe.Pointer(&length)),
		0)

	if r0 == 0 {
		if e != 0 {
			return "", syscall.Errno(e)
		}
		return "", syscall.EINVAL
	}
	return deviceid.DeviceInstance(syscall.UTF16ToString(buffer[:length])), nil
}

// GetDeviceInstallParams returns the installation parameters for a device
// or device information set. It calls the SetupDiGetDeviceInstallParams
// windows API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdigetdeviceinstallparamsw
func GetDeviceInstallParams(devices syscall.Handle, device *DevInfoData) (params DevInstallParams, err error) {
	params.Size = uint32(unsafe.Sizeof(params))

	r0, _, e := syscall.Syscall(
		procSetupDiGetDeviceInstallParams.Addr(),
		3,
		uintptr(devices),
		uintptr(unsafe.Pointer(device)),
		uintptr(unsafe.Pointer(&params)))

	if r0 == 0 {
		if e != 0 {
			err = syscall.Errno(e)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// SetDeviceInstallParams updates the installation parameters for a device
// or device information set. It calls the SetupDiSetDeviceInstallParams
// windows API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdisetdeviceinstallparamsw
func SetDeviceInstallParams(devices syscall.Handle, device *DevInfoData, params DevInstallParams) error {
	params.Size = uint32(unsafe.Sizeof(params))

	r0, _, e := syscall.Syscall(
		procSetupDiSetDeviceInstallParams.Addr(),
		3,
		uintptr(devices),
		uintptr(unsafe.Pointer(device)),
		uintptr(unsafe.Pointer(&params)))

	if r0 == 0 {
		if e != 0 {
			return syscall.Errno(e)
		}
		return syscall.EINVAL
	}
	return nil
}
