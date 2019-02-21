// +build windows

package winapi

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
	"unsafe"
)

const (
	HundredNSToTick = 0.0000001

	// systemProcessorPerformanceInformationClass information class to query with NTQuerySystemInformation
	// https://processhacker.sourceforge.io/doc/ntexapi_8h.html#ad5d815b48e8f4da1ef2eb7a2f18a54e0
	systemProcessorPerformanceInformationClass = 8

	// size of systemProcessorPerformanceInfoSize in memory
	systemProcessorPerformanceInfoSize = uint32(unsafe.Sizeof(SystemProcessorPerformanceInformation{}))
)

var (
	ntDLL        = windows.NewLazySystemDLL("Ntdll.dll")
	ntDLLLoadErr = ntDLL.Load() // attempt to load the system dll and store the err for reference

	// Windows API Proc
	// https://docs.microsoft.com/en-us/windows/desktop/api/winternl/nf-winternl-ntquerysysteminformation
	procNtQuerySystemInformation        = ntDLL.NewProc("NtQuerySystemInformation")
	procNtQuerySystemInformationLoadErr = procNtQuerySystemInformation.Find() // attempt to find the proc and store the error if needed
)

// SYSTEM_PROCESSOR_PERFORMANCE_INFORMATION
// defined in windows api doc with the following
// https://docs.microsoft.com/en-us/windows/desktop/api/winternl/nf-winternl-ntquerysysteminformation#system_processor_performance_information
// additional fields documented here
// https://www.geoffchappell.com/studies/windows/km/ntoskrnl/api/ex/sysinfo/processor_performance.htm
type SystemProcessorPerformanceInformation struct {
	IdleTime       int64 // idle time in 100ns (this is not a filetime).
	KernelTime     int64 // kernel time in 100ns.  kernel time includes idle time. (this is not a filetime).
	UserTime       int64 // usertime in 100ns (this is not a filetime).
	DpcTime        int64 // dpc time in 100ns (this is not a filetime).
	InterruptTime  int64 // interrupt time in 100ns
	InterruptCount uint32
}

func GetSystemProcessorPerformanceInformation() ([]SystemProcessorPerformanceInformation, error) {
	if ntDLLLoadErr != nil {
		return nil, errors.Wrap(ntDLLLoadErr, "winapi: can't load dll Ntdll.dll")
	}

	if procNtQuerySystemInformationLoadErr != nil {
		return nil, errors.Wrap(procNtQuerySystemInformationLoadErr, "winapi: can't get procedure NtQuerySystemInformation")
	}

	// Make maxResults large for safety.
	// We can't invoke the api call with a results array that's too small.
	// If we have more than 2056 cores on a single host, then it's probably the future.
	maxBuffer := 2056
	// buffer for results from the windows proc
	resultBuffer := make([]SystemProcessorPerformanceInformation, maxBuffer)
	// size of the buffer in memory
	bufferSize := uintptr(systemProcessorPerformanceInfoSize) * uintptr(maxBuffer)
	// size of the returned response
	var retSize uint32

	retCode, _, err := procNtQuerySystemInformation.Call(
		systemProcessorPerformanceInformationClass, // System Information Class -> SystemProcessorPerformanceInformation
		uintptr(unsafe.Pointer(&resultBuffer[0])),  // pointer to first element in result buffer
		bufferSize,                        // size of the buffer in memory
		uintptr(unsafe.Pointer(&retSize)), // pointer to the size of the returned results the windows proc will set this
	)

	if err != nil {
		return nil, err
	}

	// check return code for errors
	if retCode != 0 {
		return nil, errors.New(fmt.Sprintf("return code %d != 0, but error was not supplied", retCode))
	}

	// calculate the number of returned elements based on the returned size
	numReturnedElements := retSize / systemProcessorPerformanceInfoSize

	// trim results to the number of returned elements
	resultBuffer = resultBuffer[:numReturnedElements]
	return resultBuffer, nil
}
