// +build windows

package winapi

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"reflect"
	"unicode/utf16"
	"unsafe"
)

const (
	HundredNSToTick = 0.0000001

	// systemProcessorPerformanceInformationClass information class to query with NTQuerySystemInformation
	// https://processhacker.sourceforge.io/doc/ntexapi_8h.html#ad5d815b48e8f4da1ef2eb7a2f18a54e0
	systemProcessorPerformanceInformationClass = 8
	systemProcessorPerformanceInfoSize         = uint32(unsafe.Sizeof(SystemProcessorPerformanceInformation{}))

	// systemProcessInformationClass class to query with NTQuerySystemInformation
	// https://docs.microsoft.com/en-us/windows/desktop/api/winternl/nf-winternl-ntquerysysteminformation#system_process_information
	systemProcessInformationClass = 5
	systemProcessInfoSize         = uint32(unsafe.Sizeof(SystemProcessInformation{}))
	systemThreadInfoSize          = uint32(unsafe.Sizeof(systemThreadInformation{}))
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

// KPRIORITY
type kPriority int32

// UNICODE_STRING
type unicodeString struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}

func (u unicodeString) String() string {
	var s []uint16
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	hdr.Data = uintptr(unsafe.Pointer(u.Buffer))
	hdr.Len = int(u.Length / 2)
	hdr.Cap = int(u.MaximumLength / 2)
	return string(utf16.Decode(s))
}

// SYSTEM_PROCESS_INFORMATION
type SystemProcessInformation struct {
	NextEntryOffset              uint32        // ULONG
	NumberOfThreads              uint32        // ULONG
	WorkingSetPrivateSize        int64         // LARGE_INTEGER
	HardFaultCount               uint32        // ULONG
	NumberOfThreadsHighWatermark uint32        // ULONG
	CycleTime                    uint64        // ULONGLONG
	CreateTime                   int64         // LARGE_INTEGER
	UserTime                     int64         // LARGE_INTEGER
	KernelTime                   int64         // LARGE_INTEGER
	ImageName                    unicodeString // UNICODE_STRING
	BasePriority                 kPriority     // KPRIORITY
	UniqueProcessId              uintptr       // HANDLE
	InheritedFromUniqueProcessId uintptr       // HANDLE
	HandleCount                  uint32        // ULONG
	SessionId                    uint32        // ULONG
	UniqueProcessKey             *uint32       // ULONG_PTR
	PeakVirtualSize              uintptr       // SIZE_T
	VirtualSize                  uintptr       // SIZE_T
	PageFaultCount               uint32        // ULONG
	PeakWorkingSetSize           uintptr       // SIZE_T
	WorkingSetSize               uintptr       // SIZE_T
	QuotaPeakPagedPoolUsage      uintptr       // SIZE_T
	QuotaPagedPoolUsage          uintptr       // SIZE_T
	QuotaPeakNonPagedPoolUsage   uintptr       // SIZE_T
	QuotaNonPagedPoolUsage       uintptr       // SIZE_T
	PagefileUsage                uintptr       // SIZE_T
	PeakPagefileUsage            uintptr       // SIZE_T
	PrivatePageCount             uintptr       // SIZE_T
	ReadOperationCount           int64         // LARGE_INTEGER
	WriteOperationCount          int64         // LARGE_INTEGER
	OtherOperationCount          int64         // LARGE_INTEGER
	ReadTransferCount            int64         // LARGE_INTEGER
	WriteTransferCount           int64         // LARGE_INTEGER
	OtherTransferCount           int64         // LARGE_INTEGER
}

type kWaitReason int32

type clientID struct {
	UniqueProcess uintptr // HANDLE
	UniqueThread  uintptr // HANDLE
}

type systemThreadInformation struct {
	KernelTime      int64       // LARGE_INTEGER
	UserTime        int64       // LARGE_INTEGER
	CreateTime      int64       // LARGE_INTEGER
	WaitTime        uint32      // ULONG
	StartAddress    uintptr     // PVOID
	ClientId        clientID    // CLIENT_ID
	Priority        kPriority   // KPRIORITY
	BasePriority    int32       // LONG
	ContextSwitches uint32      // ULONG
	ThreadState     uint32      // ULONG
	WaitReason      kWaitReason // KWAIT_REASON
}

func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func checkProcNtQuerySystemInformationAvailable() error {
	if ntDLLLoadErr != nil {
		return errors.Wrap(ntDLLLoadErr, "winapi: can't load dll Ntdll.dll")
	}
	if procNtQuerySystemInformationLoadErr != nil {
		return errors.Wrap(procNtQuerySystemInformationLoadErr, "winapi: can't get procedure NtQuerySystemInformation")
	}
	return nil
}

func GetSystemProcessorPerformanceInformation() ([]SystemProcessorPerformanceInformation, error) {
	if err := checkProcNtQuerySystemInformationAvailable(); err != nil {
		return nil, err
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

	if retCode != 0 {
		return nil, fmt.Errorf("winapi call to NtQuerySystemInformation returned %d. err: %s", retCode, err.Error())
	}

	// calculate the number of returned elements based on the returned size
	numReturnedElements := retSize / systemProcessorPerformanceInfoSize

	// trim results to the number of returned elements
	resultBuffer = resultBuffer[:numReturnedElements]
	return resultBuffer, nil
}

func GetSystemProcessInformation() (map[uint32]*SystemProcessInformation, error) {
	if err := checkProcNtQuerySystemInformationAvailable(); err != nil {
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
	var currBufferSize = (uintptr(systemProcessInfoSize) + uintptr(systemThreadInfoSize)*uintptr(maxThreadsPerProc)) * uintptr(maxProcs)
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
