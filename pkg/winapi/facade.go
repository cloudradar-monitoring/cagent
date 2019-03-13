// +build windows

package winapi

import (
	"fmt"
	"reflect"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

const (
	// http://terminus.rewolf.pl/terminus/structures/ntdll/_PEB_combined.html
	processParametersAddrOffsetInPEB64bit = uintptr(0x20)

	// http://terminus.rewolf.pl/terminus/structures/ntdll/_RTL_USER_PROCESS_PARAMETERS_combined.html
	commandLineAddrOffsetInRTLUserProcessParameters64bit = uintptr(0x70)
)

// GetProcessCommandLine to read process memory structure called PEB
// https://docs.microsoft.com/en-us/windows/desktop/api/winternl/ns-winternl-_peb
// it refers to RTL_USER_PROCESS_PARAMETERS structure, which contains command line string for the process
// https://docs.microsoft.com/ru-ru/windows/desktop/api/winternl/ns-winternl-_rtl_user_process_parameters
func GetProcessCommandLine(pid uint32) (string, error) {
	handle, err := windows.OpenProcess(systemProcessAllAccess, false, pid)
	if err != nil {
		return "", err
	}
	defer func() {
		err := windows.CloseHandle(handle)
		if err != nil {
			log.Warnf("winapi: there was error closing process handle: %s", err)
		}
	}()

	pbi, err := GetProcessBasicInformation(handle)
	if err != nil {
		return "", err
	}

	if pbi.PebBaseAddress == 0 {
		// it means that we are running as 32-bit process under WOW64 and pid belongs to the 64-bit process
		return "", nil
	}

	var rtlUserProcessParametersAddr uintptr
	_, err = ReadProcessMemory(
		handle,
		uintptr(pbi.PebBaseAddress)+processParametersAddrOffsetInPEB64bit,
		uintptr(unsafe.Pointer(&rtlUserProcessParametersAddr)),
		unsafe.Sizeof(rtlUserProcessParametersAddr),
	)
	if err != nil {
		return "", err
	}

	if rtlUserProcessParametersAddr == 0 {
		return "", fmt.Errorf("winapi: error while reading [%d] process memory: RTL_USER_PROCESS_PARAMETERS addr was not found", pid)
	}

	var externalCommandLine unicodeString
	_, err = ReadProcessMemory(
		handle,
		rtlUserProcessParametersAddr+commandLineAddrOffsetInRTLUserProcessParameters64bit,
		uintptr(unsafe.Pointer(&externalCommandLine)),
		unsafe.Sizeof(externalCommandLine),
	)
	if err != nil {
		return "", err
	}

	if externalCommandLine.Length == 0 || externalCommandLine.MaximumLength == 0 {
		return "", fmt.Errorf("winapi: error while reading [%d] process memory: command line UNICODE_STRING addr was not found", pid)
	}

	buffer := make([]uint16, externalCommandLine.Length, externalCommandLine.MaximumLength)
	_, err = ReadProcessMemory(
		handle,
		uintptr(unsafe.Pointer(externalCommandLine.BufferAddr)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(externalCommandLine.Length),
	)
	if err != nil {
		return "", err
	}

	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&buffer))
	hdr.Len = int(externalCommandLine.Length / 2)
	hdr.Cap = int(externalCommandLine.MaximumLength / 2)

	return string(utf16.Decode(buffer)), nil
}

func GetIsServiceHaveDelayedAutoStartFlag(serviceHandle windows.Handle) (bool, error) {
	var resultBuffer []byte
	currentBufferSize := uint32(128)
	for {
		b := make([]byte, currentBufferSize)
		err := windows.QueryServiceConfig2(serviceHandle, systemServiceConfigDelayedAutoStartInfoClass, &b[0], currentBufferSize, &currentBufferSize)
		if err == nil {
			resultBuffer = b
			break
		}
		if err.(syscall.Errno) != syscall.ERROR_INSUFFICIENT_BUFFER {
			return false, errors.Wrapf(err, "winapi call to QueryServiceConfig2 returned %s", err)
		}
		if currentBufferSize <= uint32(len(b)) {
			return false, errors.Wrapf(err, "winapi call to QueryServiceConfig2 returned unexpected value for buffer size: %d", currentBufferSize)
		}
	}
	infoStructure := (*serviceDelayedAutoStartInfo)(unsafe.Pointer(&resultBuffer[0]))
	return infoStructure.DelayedAutoStart, nil
}
