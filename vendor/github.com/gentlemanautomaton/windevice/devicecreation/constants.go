package devicecreation

// Windows device creation flags.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdicreatedeviceinfow
const (
	GenerateID          = 0x00000001 // DICD_GENERATE_ID
	InheritClassDrivers = 0x00000002 // DICD_INHERIT_CLASSDRVS
)
