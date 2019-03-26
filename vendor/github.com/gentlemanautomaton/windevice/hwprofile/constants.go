package hwprofile

// Windows hardware profile scope flags.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/ns-setupapi-_sp_removedevice_params
const (
	Global         = 0x00000001 // DI_REMOVEDEVICE_GLOBAL, DICS_FLAG_GLOBAL
	ConfigSpecific = 0x00000002 // DI_REMOVEDEVICE_CONFIGSPECIFIC, DICS_FLAG_CONFIGSPECIFIC
	ConfigGeneral  = 0x00000004 // DICS_FLAG_CONFIGGENERAL
)
