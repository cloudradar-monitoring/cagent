package difunc

// ClassInstallHeader declares a device installation function. It implements the
// SP_CLASSINSTALL_HEADER windows API structure.
//
// ClassInstallHeader is typically embedded at the start of a params struct
// that is specific to the installation function.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/ns-setupapi-_sp_classinstall_header
type ClassInstallHeader struct {
	Size            uint32
	InstallFunction Function
}
