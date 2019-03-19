package deviceregistry

// Windows device registry key access masks.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdiopendevregkey
const (
	QueryValue       = 0x00000001 // KEY_QUERY_VALUE
	SetValue         = 0x00000002 // KEY_SET_VALUE
	CreateSubKey     = 0x00000004 // KEY_CREATE_SUB_KEY
	EnumerateSubKeys = 0x00000008 // KEY_ENUMERATE_SUB_KEYS
	Notify           = 0x00000010 // KEY_NOTIFY
	CreateLink       = 0x00000020 // KEY_CREATE_LINK
	Win32            = 0x00000200 // KEY_WOW64_32KEY
	Win64            = 0x00000100 // KEY_WOW64_64KEY
	Wrote            = 0x00020006 // KEY_WRITE
	Execute          = 0x00020019 // KEY_EXECUTE
	Read             = 0x00020019 // KEY_READ
	AllAccess        = 0x000F003F // KEY_ALL_ACCESS
)

// AccessMask defines a set of registry access rights.
type AccessMask uint32
