package difuncremove

import (
	"github.com/gentlemanautomaton/windevice/difunc"
	"github.com/gentlemanautomaton/windevice/hwprofile"
)

// Params holds device removal parameters. It implements the
// SP_REMOVEDEVICE_PARAMS windows API structure.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/ns-setupapi-_sp_removedevice_params
type Params struct {
	Header  difunc.ClassInstallHeader
	Scope   hwprofile.Scope
	Profile hwprofile.ID
}
