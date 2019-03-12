// +build windows

package winapi

import (
	"fmt"
	"unsafe"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

var (
	ntDLL        = windows.NewLazySystemDLL("Ntdll.dll")
	ntDLLLoadErr = ntDLL.Load()

	kernel32DLL    = windows.NewLazySystemDLL("kernel32.dll")
	kernel32DLLErr = kernel32DLL.Load()

	// https://docs.microsoft.com/en-us/windows/desktop/api/winternl/nf-winternl-ntquerysysteminformation
	procNtQuerySystemInformation        = ntDLL.NewProc("NtQuerySystemInformation")
	procNtQuerySystemInformationLoadErr = procNtQuerySystemInformation.Find()

	// https://docs.microsoft.com/ru-ru/windows/desktop/api/winternl/nf-winternl-ntqueryinformationprocess
	procNtQueryInformationProcess        = ntDLL.NewProc("NtQueryInformationProcess")
	procNtQueryInformationProcessLoadErr = procNtQueryInformationProcess.Find()

	// https://docs.microsoft.com/en-us/windows/desktop/api/memoryapi/nf-memoryapi-readprocessmemory
	procReadProcessMemory        = kernel32DLL.NewProc("ReadProcessMemory")
	procReadProcessMemoryLoadErr = procReadProcessMemory.Find()
)

func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func checkNtDLLProceduresAvailable() error {
	if ntDLLLoadErr != nil {
		return errors.Wrap(ntDLLLoadErr, "winapi: can't load dll Ntdll.dll")
	}
	if procNtQuerySystemInformationLoadErr != nil {
		return errors.Wrap(procNtQuerySystemInformationLoadErr, "winapi: can't get procedure NtQuerySystemInformation")
	}
	if procNtQueryInformationProcessLoadErr != nil {
		return errors.Wrap(procNtQueryInformationProcessLoadErr, "winapi: can't get procedure NtQueryInformationProcess")
	}
	return nil
}

func checkKernel32ProceduresAvailable() error {
	if kernel32DLLErr != nil {
		return errors.Wrap(kernel32DLLErr, "winapi: can't load dll kernel32.dll")
	}
	if procReadProcessMemoryLoadErr != nil {
		return errors.Wrap(procReadProcessMemoryLoadErr, "winapi: can't get procedure ReadProcessMemory")
	}
	return nil
}

func GetSystemProcessorPerformanceInformation() ([]SystemProcessorPerformanceInformation, error) {
	if err := checkNtDLLProceduresAvailable(); err != nil {
		return nil, err
	}

	// Make maxResults large for safety.
	// We can't invoke the api call with a results array that's too small.
	// If we have more than 2056 cores on a single host, then it's probably the future.
	maxBuffer := 2056
	// buffer for results from the windows proc
	resultBuffer := make([]SystemProcessorPerformanceInformation, maxBuffer)
	// size of the buffer in memory
	bufferSize := systemProcessorPerformanceInfoSize * uintptr(maxBuffer)
	// size of the returned response
	var retSize uint32

	retCode, _, err := procNtQuerySystemInformation.Call(
		systemProcessorPerformanceInformationClass, // System Information Class -> SystemProcessorPerformanceInformation
		uintptr(unsafe.Pointer(&resultBuffer[0])),  // pointer to first element in result buffer
		bufferSize,                        // size of the buffer in memory
		uintptr(unsafe.Pointer(&retSize)), // pointer to the size of the returned results the windows proc will set this
	)

	if retCode != 0 {
		return nil, fmt.Errorf("winapi call to NtQuerySystemInformation returned %d. err: %s", retCode, err.Error())
	}

	// calculate the number of returned elements based on the returned size
	numReturnedElements := retSize / uint32(systemProcessorPerformanceInfoSize)

	// trim results to the number of returned elements
	resultBuffer = resultBuffer[:numReturnedElements]
	return resultBuffer, nil
}

func GetSystemProcessInformation() (map[uint32]*SystemProcessInformation, error) {
	if err := checkNtDLLProceduresAvailable(); err != nil {
		return nil, err
	}

	var p *SystemProcessInformation
	var retSize uint32
	var retCode uintptr
	var err error

	callWithBufferSize := func(size uintptr) {
		buffer := make([]byte, size)
		p = (*SystemProcessInformation)(unsafe.Pointer(&buffer[0]))
		retCode, _, err = procNtQuerySystemInformation.Call(
			systemProcessInformationClass,
			uintptr(unsafe.Pointer(p)),
			uintptr(size),
			uintptr(unsafe.Pointer(&retSize)),
		)
	}

	const maxProcs = 300
	const maxThreadsPerProc = 100
	// technically, windows can have more processes and more threads per process, but we try to calculate average value
	var currBufferSize = (systemProcessInfoSize + systemThreadInfoSize*uintptr(maxThreadsPerProc)) * uintptr(maxProcs)
	callWithBufferSize(currBufferSize)

	if retCode != 0 {
		log.Debugf(
			"winapi call to NtQuerySystemInformation returned code: %d. required buffer size: %d. actual size: %d",
			retCode, retSize, currBufferSize,
		)

		if uintptr(retSize) > currBufferSize {
			log.Debugf("winapi: trying to call again with increased buffer size")
			currBufferSize = uintptr(retSize)
			callWithBufferSize(currBufferSize)
		}

		if retCode != 0 {
			return nil, errors.Wrapf(
				err,
				"winapi call to NtQuerySystemInformation returned code: %d. required buffer size: %d. actual size: %d",
				retCode, retSize, currBufferSize,
			)
		}
	}

	var counter int
	result := make(map[uint32]*SystemProcessInformation)
	for {
		result[uint32((*p).UniqueProcessId)] = p
		counter++
		if p.NextEntryOffset == 0 {
			break
		}
		p = (*SystemProcessInformation)(add(unsafe.Pointer(p), uintptr(p.NextEntryOffset)))
	}

	if len(result) != counter {
		log.Warnf("winapi: parsing information failed: Returned %d processes, saved %d", counter, len(result))
	}

	return result, nil
}

// returns address for ProcessEnvironmentBlock struct
func GetProcessBasicInformation(processHandle windows.Handle) (*processBasicInformation, error) {
	if err := checkNtDLLProceduresAvailable(); err != nil {
		return nil, err
	}

	var pbi processBasicInformation
	var retSize int
	retCode, _, err := procNtQueryInformationProcess.Call(
		uintptr(processHandle),
		systemProcessBasicInformationClass,
		uintptr(unsafe.Pointer(&pbi)),
		systemProcessBasicInformationSize,
		uintptr(unsafe.Pointer(&retSize)),
	)

	if retCode != 0 {
		return nil, fmt.Errorf("winapi call to NtQueryInformationProcess returned %d. err: %s", retCode, err.Error())
	}

	return &pbi, nil
}

func ReadProcessMemory(processHandle windows.Handle, srcAddr uintptr, dstAddr uintptr, size uintptr) (int, error) {
	var nBytesRead int

	if err := checkKernel32ProceduresAvailable(); err != nil {
		return nBytesRead, err
	}

	retCode, _, err := procReadProcessMemory.Call(
		uintptr(processHandle),
		srcAddr,
		dstAddr,
		size,
		uintptr(unsafe.Pointer(&nBytesRead)),
	)

	// ReadProcessMemory function returns 0 in the case of failure
	if retCode == 0 {
		return nBytesRead, fmt.Errorf("winapi call to ReadProcessMemory returned %d. err: %s", retCode, err.Error())
	}

	return nBytesRead, nil
}
