package deviceregistry

// Windows device registry key types.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdiopendevregkey
const (
	Device = 0x00000001 // DIREG_DEV
	Driver = 0x00000002 // DIREG_DRV
	Both   = 0x00000004 // DIREG_BOTH
)

// KeyType specifies a type of device registry key.
type KeyType uint32
