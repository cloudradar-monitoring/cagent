package setupapi

import (
	"syscall"

	"github.com/gentlemanautomaton/windevice/diflag"
	"github.com/gentlemanautomaton/windevice/diflagex"
)

// DevInstallParams holds device installation parameters. It implements the
// SP_DEVINSTALL_PARAMS_W windows API structure.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/ns-setupapi-_sp_devinstall_params_w
type DevInstallParams struct {
	Size                     uint32
	Flags                    diflag.Value
	FlagsEx                  diflagex.Value
	ParentWindow             syscall.Handle
	InstallMsgHandler        *func()
	InstallMsgHandlerContext *byte
	FileQueue                syscall.Handle
	ClassInstallReserved     uintptr
	reserved                 uint32
	DriverPath               Path
}
