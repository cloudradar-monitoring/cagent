package deviceregistry

// Windows device registry property codes.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/nf-setupapi-setupdigetdeviceregistrypropertyw
const (
	Description              = 0  // SPDRP_DEVICEDESC
	HardwareID               = 1  // SPDRP_HARDWAREID
	CompatibleID             = 2  // SPDRP_COMPATIBLEIDS
	_                        = 3  // SPDRP_UNUSED0
	Service                  = 4  // SPDRP_SERVICE
	_                        = 5  // SPDRP_UNUSED1
	_                        = 6  // SPDRP_UNUSED2
	Class                    = 7  // SPDRP_CLASS
	ClassGUID                = 8  // SPDRP_CLASSGUID
	DriverRegName            = 9  // SPDRP_DRIVER
	ConfigFlags              = 10 // SPDRP_CONFIGFLAGS
	Manufacturer             = 11 // SPDRP_MFG
	FriendlyName             = 12 // SPDRP_FRIENDLYNAME
	LocationInformation      = 13 // SPDRP_LOCATION_INFORMATION
	PhysicalDeviceObjectName = 14 // SPDRP_PHYSICAL_DEVICE_OBJECT_NAME
	Capabilities             = 15 // SPDRP_CAPABILITIES
	UINumber                 = 16 // SPDRP_UI_NUMBER
	UpperFilters             = 17 // SPDRP_UPPERFILTERS
	LowerFilters             = 18 // SPDRP_LOWERFILTERS
	BusTypeGUID              = 19 // SPDRP_BUSTYPEGUID
	LegacyBusType            = 20 // SPDRP_LEGACYBUSTYPE
	BusNumber                = 21 // SPDRP_BUSNUMBER
	EnumeratorName           = 22 // SPDRP_ENUMERATOR_NAME
	Security                 = 23 // SPDRP_SECURITY
	SecuritySDS              = 24 // SPDRP_SECURITY_SDS
	DevType                  = 25 // SPDRP_DEVTYPE
	Exclusive                = 26 // SPDRP_EXCLUSIVE
	Characteristics          = 27 // SPDRP_CHARACTERISTICS
	Address                  = 28 // SPDRP_ADDRESS
	UINumberDescFormat       = 29 // SPDRP_UI_NUMBER_DESC_FORMAT
	DevicePowerData          = 30 // SPDRP_DEVICE_POWER_DATA
	RemovalPolicy            = 31 // SPDRP_REMOVAL_POLICY
	RemovalPolicyHWDefault   = 32 // SPDRP_REMOVAL_POLICY_HW_DEFAULT
	RemovalPolicyOverride    = 33 // SPDRP_REMOVAL_POLICY_OVERRIDE
	InstallState             = 34 // SPDRP_INSTALL_STATE
	LocationPaths            = 35 // SPDRP_LOCATION_PATHS
)

// Code identifies a device registry property.
type Code uint32
