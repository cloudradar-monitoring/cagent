package setupapi

import (
	"syscall"
	"unsafe"

	"github.com/gentlemanautomaton/windevice/deviceproperty"
)

var (
	procSetupDiGetDevicePropertyKeys = modsetupapi.NewProc("SetupDiGetDevicePropertyKeys")
	procSetupDiGetDeviceProperty     = modsetupapi.NewProc("SetupDiGetDevicePropertyW")
)

// GetDevicePropertyKeys returns all of the property keys for a device
// instance. It calls the SetupDiGetDevicePropertyW windows API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdigetdevicepropertykeys
func GetDevicePropertyKeys(devices syscall.Handle, device DevInfoData) (keys []deviceproperty.Key, err error) {
	// Make up to 3 attempts to get the property data
	const rounds = 3
	for i := 0; i < rounds; i++ {
		var length uint32
		length, err = getDevicePropertyKeys(devices, device, keys)
		if err == nil {
			keys = keys[:length]
			return
		}
		if err == syscall.ERROR_INSUFFICIENT_BUFFER && i < rounds {
			keys = make([]deviceproperty.Key, length)
		} else {
			return
		}
	}
	return
}

func getDevicePropertyKeys(devices syscall.Handle, device DevInfoData, buffer []deviceproperty.Key) (reqSize uint32, err error) {
	var pk *deviceproperty.Key
	if len(buffer) > 0 {
		pk = &buffer[0]
	}

	r0, _, e := syscall.Syscall6(
		procSetupDiGetDevicePropertyKeys.Addr(),
		6,
		uintptr(devices),
		uintptr(unsafe.Pointer(&device)),
		uintptr(unsafe.Pointer(pk)),
		uintptr(len(buffer)),
		uintptr(unsafe.Pointer(&reqSize)),
		0) // Flags are always zero
	if r0 == 0 {
		if e != 0 {
			err = syscall.Errno(e)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// GetDeviceProperty returns a device property for a device instance.
// It calls the SetupDiGetDevicePropertyW windows API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdigetdevicepropertyw
func GetDeviceProperty(devices syscall.Handle, device DevInfoData, key deviceproperty.Key) (value deviceproperty.Value, err error) {
	// Allocate a buffer on the stack
	var b [1024]byte
	buffer := b[:]

	// Make up to 3 attempts to get the property data
	const rounds = 3
	for i := 0; i < rounds; i++ {
		var length uint32
		var dataType deviceproperty.Type
		length, dataType, err = getDeviceProperty(devices, device, key, buffer)
		if err == nil {
			value = deviceproperty.NewValue(dataType, buffer[:length])
			return
		}
		if err == syscall.ERROR_INSUFFICIENT_BUFFER && i < rounds {
			buffer = make([]byte, length)
		} else {
			return
		}
	}
	return
}

func getDeviceProperty(devices syscall.Handle, device DevInfoData, key deviceproperty.Key, buffer []byte) (reqSize uint32, dataType deviceproperty.Type, err error) {
	var pb *byte
	if len(buffer) > 0 {
		pb = &buffer[0]
	}

	r0, _, e := syscall.Syscall9(
		procSetupDiGetDeviceProperty.Addr(),
		8,
		uintptr(devices),
		uintptr(unsafe.Pointer(&device)),
		uintptr(unsafe.Pointer(&key)),
		uintptr(unsafe.Pointer(&dataType)),
		uintptr(unsafe.Pointer(pb)),
		uintptr(len(buffer)),
		uintptr(unsafe.Pointer(&reqSize)),
		0, // Flags are always zero
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
