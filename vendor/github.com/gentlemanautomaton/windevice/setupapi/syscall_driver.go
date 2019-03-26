package setupapi

import (
	"io"
	"syscall"
	"unsafe"
)

var (
	procSetupDiBuildDriverInfoList   = modsetupapi.NewProc("SetupDiBuildDriverInfoList")
	procSetupDiDestroyDriverInfoList = modsetupapi.NewProc("SetupDiDestroyDriverInfoList")
	procSetupDiEnumDriverInfo        = modsetupapi.NewProc("SetupDiEnumDriverInfoW")
)

// BuildDriverInfoList builds a driver information list that contains
// drivers of the requested driver type. It calls the SetupDiBuildDriverInfoList
// windows API function.
//
// The driver list will be affiliated with the device list identified by
// devices. Driver information can be retrieved from the list by calling the
// EnumDriverInfo function.
//
// The driver list membership is influenced by four factors:
//  1. The requested driver type
//  2. The setup class associated with the device list, if any
//  3. The installation and enumeration flags associated with the device list
//  4. The provided device
//
// It is the caller's responsibility to destroy the driver list when it is no
// longer needed by calling DestroyDriverInfoList with the same parameters.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdibuilddriverinfolist
func BuildDriverInfoList(devices syscall.Handle, device *DevInfoData, driverType uint32) error {
	r0, _, e := syscall.Syscall(
		procSetupDiBuildDriverInfoList.Addr(),
		3,
		uintptr(devices),
		uintptr(unsafe.Pointer(device)),
		uintptr(driverType))
	if r0 == 0 {
		if e != 0 {
			return syscall.Errno(e)
		}
		return syscall.EINVAL
	}
	return nil
}

// DestroyDriverInfoList destroys a device information list. It calls the
// SetupDiDestroyDriverInfoList windows API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdidestroydriverinfolist
func DestroyDriverInfoList(devices syscall.Handle, device *DevInfoData, driverType uint32) error {
	r0, _, e := syscall.Syscall(
		procSetupDiDestroyDriverInfoList.Addr(),
		3,
		uintptr(devices),
		uintptr(unsafe.Pointer(device)),
		uintptr(driverType))
	if r0 == 0 {
		if e != 0 {
			return syscall.Errno(e)
		}
		return syscall.EINVAL
	}
	return nil
}

// EnumDriverInfo returns information about a driver in a driver information
// list. It calls the SetupDiEnumDriverInfo windows API function.
//
// EnumDriverInfo returns io.EOF when there are no more members in the list.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdienumdriverinfow
func EnumDriverInfo(devices syscall.Handle, device *DevInfoData, driverType uint32, index uint32) (info DrvInfoData, err error) {
	const errNoMoreItems = 259

	info.Size = uint32(unsafe.Sizeof(info))

	r0, _, e := syscall.Syscall6(
		procSetupDiEnumDriverInfo.Addr(),
		5,
		uintptr(devices),
		uintptr(unsafe.Pointer(device)),
		uintptr(driverType),
		uintptr(index),
		uintptr(unsafe.Pointer(&info)),
		0)
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
