package setupapi

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/gentlemanautomaton/windevice/deviceregistry"
)

var (
	procSetupDiGetDeviceRegistryProperty = modsetupapi.NewProc("SetupDiGetDeviceRegistryPropertyW")
	procSetupDiSetDeviceRegistryProperty = modsetupapi.NewProc("SetupDiSetDeviceRegistryPropertyW")
)

// GetDeviceRegistryString retrieves a property from the registry as a string.
func GetDeviceRegistryString(devices syscall.Handle, device DevInfoData, property deviceregistry.Code) (value string, err error) {
	var buffer [1024 * 2]byte
	dataType, data, err := GetDeviceRegistryProperty(devices, device, property, buffer[:])
	if err != nil {
		return "", err
	}

	switch dataType {
	case syscall.REG_SZ, syscall.REG_EXPAND_SZ:
		return utf16BytesToString(data), nil
	default:
		return "", fmt.Errorf("expected REG_SZ registry type but received type %d", dataType)
	}
}

// GetDeviceRegistryStrings retrieves a property from the registry as a
// slice of strings.
func GetDeviceRegistryStrings(devices syscall.Handle, device DevInfoData, property deviceregistry.Code) (values []string, err error) {
	var buffer [1024 * 2]byte
	dataType, data, err := GetDeviceRegistryProperty(devices, device, property, buffer[:])
	if err != nil {
		return nil, err
	}

	switch dataType {
	case syscall.REG_SZ, syscall.REG_EXPAND_SZ:
		return []string{utf16BytesToString(data)}, nil
	case syscall.REG_MULTI_SZ:
		return utf16BytesToSplitString(data), nil
	default:
		return nil, fmt.Errorf("expected REG_MULTI_SZ registry type but received type %d", dataType)
	}
}

// GetDeviceRegistryUint32 retrieves a property from the registry as a uint32.
func GetDeviceRegistryUint32(devices syscall.Handle, device DevInfoData, property deviceregistry.Code) (value uint32, err error) {
	var buffer [4]byte
	dataType, data, err := GetDeviceRegistryProperty(devices, device, property, buffer[:])
	if err != nil {
		return 0, err
	}

	if len(data) != 4 {
		return 0, fmt.Errorf("expected 4-byte DWORD but received %d bytes", len(data))
	}

	switch dataType {
	case syscall.REG_DWORD_LITTLE_ENDIAN:
		return binary.LittleEndian.Uint32(data), nil
	case syscall.REG_DWORD_BIG_ENDIAN:
		return binary.BigEndian.Uint32(data), nil
	default:
		return 0, fmt.Errorf("expected REG_DWORD registry type but received type %d", dataType)
	}
}

// GetDeviceRegistryProperty retrieves a member property from a device
// information list. It calls the SetupDiGetDeviceRegistryProperty windows
// API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdigetdeviceregistrypropertyw
func GetDeviceRegistryProperty(devices syscall.Handle, device DevInfoData, property deviceregistry.Code, buffer []byte) (dataType uint32, data []byte, err error) {
	// Make up to 3 attempts to get the property data.
	const rounds = 3
	for i := 0; i < rounds; i++ {
		var length uint32
		length, dataType, err = getDeviceRegistryProperty(devices, device, property, buffer)
		if err == nil {
			data = buffer[:length]
			break
		}
		if err == syscall.ERROR_INSUFFICIENT_BUFFER && i < rounds {
			buffer = make([]byte, length)
		} else {
			return dataType, nil, err
		}
	}
	return dataType, data, err
}

func getDeviceRegistryProperty(devices syscall.Handle, device DevInfoData, property deviceregistry.Code, buffer []byte) (reqSize uint32, registryDataType uint32, err error) {
	if len(buffer) == 0 {
		return 0, 0, ErrEmptyBuffer
	}

	r0, _, e := syscall.Syscall9(
		procSetupDiGetDeviceRegistryProperty.Addr(),
		7,
		uintptr(devices),
		uintptr(unsafe.Pointer(&device)),
		uintptr(property),
		uintptr(unsafe.Pointer(&registryDataType)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
		uintptr(unsafe.Pointer(&reqSize)),
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

// SetDeviceRegistryString sets a device registry property to a string value.
//
// The value will be encoded in REG_SZ format.
func SetDeviceRegistryString(devices syscall.Handle, device DevInfoData, property deviceregistry.Code, value string) (err error) {
	buffer, err := utf16BytesFromString(value)
	if err != nil {
		return err
	}
	return SetDeviceRegistryProperty(devices, device, property, buffer)
}

// SetDeviceRegistryStrings sets a device registry property to a set of
// string values.
//
// The values will be encoded in REG_MULTI_SZ format.
func SetDeviceRegistryStrings(devices syscall.Handle, device DevInfoData, property deviceregistry.Code, values []string) (err error) {
	buffer, err := utf16BytesFromStrings(values)
	if err != nil {
		return err
	}
	return SetDeviceRegistryProperty(devices, device, property, buffer)
}

// SetDeviceRegistryProperty sets a device registry property.
// It calls the SetupDiSetDeviceRegistryProperty windows API function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdisetdeviceregistrypropertyw
func SetDeviceRegistryProperty(devices syscall.Handle, device DevInfoData, property deviceregistry.Code, buffer []byte) (err error) {
	var pb *byte
	if len(buffer) > 0 {
		pb = &buffer[0]
	}
	r0, _, e := syscall.Syscall6(
		procSetupDiSetDeviceRegistryProperty.Addr(),
		5,
		uintptr(devices),
		uintptr(unsafe.Pointer(&device)),
		uintptr(property),
		uintptr(unsafe.Pointer(pb)),
		uintptr(len(buffer)),
		0)

	if r0 == 0 {
		if e != 0 {
			return syscall.Errno(e)
		}
		return syscall.EINVAL
	}
	return nil
}
