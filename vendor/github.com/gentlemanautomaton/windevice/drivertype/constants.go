package drivertype

// Windows driver types.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/ns-setupapi-_sp_drvinfo_data_v2_w
const (
	NoDriver     = 0 // SPDIT_NODRIVER
	ClassDriver  = 1 // SPDIT_CLASSDRIVER
	CompatDriver = 2 // SPDIT_COMPATDRIVER
)
