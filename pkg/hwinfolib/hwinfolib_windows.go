// +build windows

package hwinfolib

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
)

// Flag definitions for Init() call
const (
	HWiFlagDebugModeEnable = (1 << 0) // Enable Debug Mode
	HWiFlagSWSMIEnable     = (1 << 1) // Enable SW SMI

	HWiFlagIDESafe = (0 << 2) // Use Safe IDE Mode (recommended)
	HWiFlagIDEAdr1 = (1 << 2) // Enable Low-level IDE scan up-to Device #1
	HWiFlagIDEAdr2 = (2 << 2) // Enable Low-level IDE scan up-to Device #2
	HWiFlagIDEAdr4 = (4 << 2) // Enable Low-level IDE scan up-to Device #4

	HWiFlagProbPCIScan      = (1 << 5)  // Check Problematic PCI devices (0=skip them)
	HWiFlagSMBUSEnable      = (1 << 6)  // Enable SMBus support
	HWiFlagECDisable        = (1 << 7)  // Disable Embedded Controller support
	HWiFlagHPETDisable      = (1 << 8)  // Disable using of High-Precision Event Timers
	HWiFlagGPUI2CDisable    = (1 << 10) // Disable support of GPU I2C
	HWiFlagIOCTLKernel      = (1 << 11) // Enable routing IOCTLs via kernel
	HWiFlagPersistentDriver = (1 << 12) // Enable Persistent Driver
	HWiFlagDriveScanDisable = (1 << 13) // Disable entire Drive Scan
	HWiFlagPCIDirect        = (1 << 14) // Enable Low-level PCI Scan
	HWiFlagCSMISASDisable   = (1 << 15) // Disable CSMI-SAS support
	HWiFlagIMEDisable       = (1 << 16) // Disable Intel Management Engine Support
	HWiFlagGPUWakeExt       = (1 << 17) // Perform Extended GPU wake-up method for disabled secondary GPUs
	HWiFlagPreferADL        = (1 << 18) // 1=report certain AMD GPU parameters (e.g. clocks, utilization) from AMD GPU driver (ADL) instead of direct-access

	// This is the recommended default setting
	HWiFlagDefault = (HWiFlagSMBUSEnable | HWiFlagPCIDirect) // Enable SMBus support + Low-level PCI Scan
)

const (
	errNoOK                    = 0
	errNoHwinfoLibNotFound     = 0xFFFF01
	errNoHwinfoLibProcNotFound = 0xFFFF02
)

var libHandle *syscall.Handle

// TryLoadLibrary returns true if the library was successfully found, loaded and initialized
func TryLoadLibrary() (bool, error) {
	if libHandle != nil {
		// already initialized
		return true, nil
	}

	dll, err := syscall.LoadLibrary("hwinfo.dll")
	if err != nil {
		return false, err
	}

	libHandle = &dll

	initFn, err := syscall.GetProcAddress(*libHandle, "Init")
	if err != nil {
		return false, err
	}

	resultCode, _, _ := syscall.Syscall(initFn, 1, uintptr(HWiFlagDefault), 0, 0)
	return resultCode == errNoOK, mapReturnCodeToError(resultCode, false)
}

// DeInit Call this function to de-initialize HWiNFO SDK, when there will be no further calls issued.
// This function will unload the HWiNFO SDK driver as well.
func DeInit() error {
	if libHandle == nil {
		return nil
	}

	defer func() {
		_ = syscall.FreeLibrary(*libHandle)
		libHandle = nil
	}()

	fn, err := syscall.GetProcAddress(*libHandle, "DeInit")
	if err != nil {
		return err
	}

	resultCode, _, _ := syscall.Syscall(fn, 0, 0, 0, 0)
	return mapReturnCodeToError(resultCode, false)
}

// GetNumberOfDetectedSensors Call this function to get the number of sensors detected
// Returns Number of sensors detected by the HWiNFO SDK. Use this value as the maximum number of iSensor value passed to subsequent functions.
func GetNumberOfDetectedSensors() (int, error) {
	fn, err := syscall.GetProcAddress(*libHandle, "GetNumberOfDetectedSensors")
	if err != nil {
		return 0, err
	}

	resultCode, _, _ := syscall.Syscall(fn, 0, 0, 0, 0)
	return int(resultCode), mapReturnCodeToError(resultCode, true)
}

// ReadDataFromSensor Call this function (periodically) to have the specified sensor read all values from hardware (temperatures, voltages, fans, etc.).
// Use this function whenever it is required to (re-)read all sensor values from hardware. Subsequent functions return particular readings that have been scanned by calling this function.
// iSensor is the sensor number index (0 to Max value returned by the GetNumberOfDetectedSensors() function)
func ReadDataFromSensor(iSensor int) error {
	fn, err := syscall.GetProcAddress(*libHandle, "ReadDataFromSensor")
	if err != nil {
		return err
	}

	resultCode, _, _ := syscall.Syscall(fn, 1, uintptr(iSensor), 0, 0)
	return mapReturnCodeToError(resultCode, false)
}

// GetSensorName Call this function to get the name of sensor specified.
func GetSensorName(iSensor int) (string, error) {
	fn, err := syscall.GetProcAddress(*libHandle, "GetSensorName")
	if err != nil {
		return "", err
	}

	const bufferSize = 256
	buffer := make([]byte, bufferSize)
	var outInfoResultSize int
	resultCode, _, _ := syscall.Syscall6(
		fn,
		4,
		uintptr(iSensor),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(&outInfoResultSize)),
		0,
		0,
	)

	err = mapReturnCodeToError(resultCode, false)
	if err != nil {
		return "", err
	}

	buffer = buffer[:outInfoResultSize]
	return string(buffer), nil
}

// GetTemperature Call this function to get the sensor temperature value.
// This function when called doesn’t read the value from hardware directly.
// Use the ReadDataFromSensor function to read from hardware and then get the particular reading value using this function.
// iSensor: Sensor number index (0 to Max value returned by the HWi32_GetNumberOfDetectedSensors() function)
// index: Index of the temperature reading for a particular sensor (0 - 128). The reading description is returned in the supplied buffer.
// outBuffer: Pointer to a buffer to return the description of the indexed temperature reading.
// outBufferSize: Size of the buffer supplied.
// Returns the temperature value in degrees of Celsius.
// 0 = Not Available (Sensor input not connected, not present or value = 0)
func GetTemperature(iSensor int, index int) (string, float64, error) {
	fn, err := syscall.GetProcAddress(*libHandle, "GetTemperature")
	if err != nil {
		return "", 0.0, err
	}

	const bufferSize = 256
	buffer := make([]byte, bufferSize)
	var temperature float64
	var outInfoResultSize int
	returnValue, _, _ := syscall.Syscall6(
		fn,
		6,
		uintptr(iSensor),
		uintptr(index),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(&outInfoResultSize)),
		uintptr(unsafe.Pointer(&temperature)),
	)

	err = mapReturnCodeToError(returnValue, false)
	if err != nil {
		return "", temperature, err
	}

	buffer = buffer[:outInfoResultSize]
	return string(buffer), temperature, nil
}

func mapReturnCodeToError(code uintptr, ignoreUnknownErrors bool) error {
	switch int(code) {
	case errNoHwinfoLibNotFound:
		return errors.WithStack(errors.New("HWiNFO library wasn't found"))
	case errNoHwinfoLibProcNotFound:
		return errors.WithStack(errors.New("HWiNFO procedure missing"))
	case errNoOK:
		return nil
	default:
		if ignoreUnknownErrors {
			return nil
		}
		return errors.WithStack(fmt.Errorf("hwinfo library call returned %d. Check that you have the administrative rights", code))
	}
}
