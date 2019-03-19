package setupapi

import (
	"syscall"
	"unsafe"

	"github.com/gentlemanautomaton/windevice/deviceregistry"
	"github.com/gentlemanautomaton/windevice/hwprofile"
	"golang.org/x/sys/windows/registry"
)

var (
	procSetupDiOpenDevRegKey = modsetupapi.NewProc("SetupDiOpenDevRegKey")
)

// OpenDevRegKey opens a device registry key and returns a handle to it.
// It calls the SetupDiOpenDevRegKey windows API function.
//
// It is the caller's respnosibility to close the registry key handle when
// finished with it.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdiopendevregkey
func OpenDevRegKey(devices syscall.Handle, device DevInfoData, scope hwprofile.Scope, profile hwprofile.ID, keyType deviceregistry.KeyType, access deviceregistry.AccessMask) (key registry.Key, err error) {
	r0, _, e := syscall.Syscall6(
		procSetupDiOpenDevRegKey.Addr(),
		6,
		uintptr(devices),
		uintptr(unsafe.Pointer(&device)),
		uintptr(scope),
		uintptr(profile),
		uintptr(keyType),
		uintptr(access))
	key = registry.Key(r0)
	if key == registry.Key(syscall.InvalidHandle) {
		if e != 0 {
			err = syscall.Errno(e)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
