// +build windows

package winapi

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32                       = syscall.MustLoadDLL("user32.dll")
	procEnumWindows              = user32.MustFindProc("EnumWindows")
	procGetWindowThreadProcessId = user32.MustFindProc("GetWindowThreadProcessId")
	procIsHungAppWindow          = user32.MustFindProc("IsHungAppWindow")
)

func EnumWindows(enumFunc uintptr, lparam uintptr) (err error) {
	r1, _, e1 := syscall.Syscall(procEnumWindows.Addr(), 2, uintptr(enumFunc), uintptr(lparam), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func GetWindowThreadProcessId(hwnd syscall.Handle, str *uint32) (len int32, err error) {
	r0, _, e1 := syscall.Syscall(procGetWindowThreadProcessId.Addr(), 2, uintptr(hwnd), uintptr(unsafe.Pointer(str)), 0)
	len = int32(r0)
	if len == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func WindowByProcessId() (map[uint32]syscall.Handle, error) {
	m := make(map[uint32]syscall.Handle)

	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		var windowProcessId uint32
		_, err := GetWindowThreadProcessId(h, &windowProcessId)
		if err != nil {
			// ignore the error
			return 1 // continue enumeration
		}
		m[windowProcessId] = h

		return 1 // continue enumeration
	})

	err := EnumWindows(cb, 0)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func IsHangWindow(hwnd syscall.Handle) (bool, error) {
	isHang, _, e1 := syscall.Syscall(procIsHungAppWindow.Addr(), 1, uintptr(hwnd), 0, 0)
	if e1 != 0 {
		return false, error(e1)
	}

	fmt.Printf("IsHangWindow %d", isHang)

	return isHang == 1, nil
}
