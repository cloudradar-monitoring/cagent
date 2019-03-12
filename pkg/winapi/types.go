// +build windows

package winapi

import (
	"reflect"
	"unicode/utf16"
	"unsafe"
)

const (
	HundredNSToTick = 0.0000001

	// systemProcessorPerformanceInformationClass information class to query with NTQuerySystemInformation()
	// https://processhacker.sourceforge.io/doc/ntexapi_8h.html#ad5d815b48e8f4da1ef2eb7a2f18a54e0
	systemProcessorPerformanceInformationClass = 8
	systemProcessorPerformanceInfoSize         = unsafe.Sizeof(SystemProcessorPerformanceInformation{})

	// systemProcessInformationClass class to query with NTQuerySystemInformation()
	// https://docs.microsoft.com/en-us/windows/desktop/api/winternl/nf-winternl-ntquerysysteminformation#system_process_information
	systemProcessInformationClass = 5
	systemProcessInfoSize         = unsafe.Sizeof(SystemProcessInformation{})
	systemThreadInfoSize          = unsafe.Sizeof(systemThreadInformation{})

	// systemProcessAllAccess class to query with OpenProcess()
	// https://docs.microsoft.com/ru-ru/windows/desktop/ProcThread/process-security-and-access-rights
	systemProcessAllAccess = 0x1F0FFF

	// systemProcessBasicInformationClass class to query with NtQueryInformationProcess
	// returns PROCESS_BASIC_INFORMATION struct
	systemProcessBasicInformationClass = 0
	systemProcessBasicInformationSize  = unsafe.Sizeof(processBasicInformation{})
)

// SYSTEM_PROCESSOR_PERFORMANCE_INFORMATION
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
	BufferAddr    *uint16
}

func (u unicodeString) String() string {
	var s []uint16
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	hdr.Data = uintptr(unsafe.Pointer(u.BufferAddr))
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

// KWAIT_REASON
type kWaitReason int32

// CLIENT_ID
type clientID struct {
	UniqueProcess uintptr // HANDLE
	UniqueThread  uintptr // HANDLE
}

// SYSTEM_THREAD_INFORMATION
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

// PROCESS_BASIC_INFORMATION
type processBasicInformation struct {
	ExitStatus                   uintptr
	PebBaseAddress               uintptr
	AffinityMask                 uintptr
	BasePriority                 int32
	UniqueProcessID              uintptr
	InheritedFromUniqueProcessId uintptr
}
