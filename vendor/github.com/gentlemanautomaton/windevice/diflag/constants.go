package diflag

// Windows device installation and enumeration flags.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/ns-setupapi-_sp_devinstall_params_w
const (
	ShowOEM                        = 0x00000001 // DI_SHOWOEM
	ShowCompat                     = 0x00000002 // DI_SHOWCOMPAT
	ShowClass                      = 0x00000004 // DI_SHOWCLASS
	ShowAll                        = 0x00000007 // DI_SHOWALL
	NoVCP                          = 0x00000008 // DI_NOVCP
	DidCompat                      = 0x00000010 // DI_DIDCOMPAT
	DidClass                       = 0x00000020 // DI_DIDCLASS
	AutoAssignResources            = 0x00000040 // DI_AUTOASSIGNRES
	NeedRestart                    = 0x00000080 // DI_NEEDRESTART
	NeedReboot                     = 0x00000100 // DI_NEEDREBOOT
	NoBrowse                       = 0x00000200 // DI_NOBROWSE
	MultipleManufacturers          = 0x00000400 // DI_MULTMFGS
	Disabled                       = 0x00000800 // DI_DISABLED
	GeneralPageAdded               = 0x00001000 // DI_GENERALPAGE_ADDED
	ResourcePageAdded              = 0x00002000 // DI_RESOURCEPAGE_ADDED
	PropertiesChange               = 0x00004000 // DI_PROPERTIES_CHANGE
	InfIsSorted                    = 0x00008000 // DI_INF_IS_SORTED
	EnumSingleInf                  = 0x00010000 // DI_ENUMSINGLEINF
	DoNotCallConfigManager         = 0x00020000 // DI_DONOTCALLCONFIGMG
	InstallDisabled                = 0x00040000 // DI_INSTALLDISABLED
	CompatFromClass                = 0x00080000 // DI_COMPAT_FROM_CLASS
	ClassInstallParams             = 0x00100000 // DI_CLASSINSTALLPARAMS
	NoInstallerDefaultAction       = 0x00200000 // DI_NODI_DEFAULTACTION
	QuietInstall                   = 0x00800000 // DI_QUIETINSTALL
	NoFileCopy                     = 0x01000000 // DI_NOFILECOPY
	ForceCopy                      = 0x02000000 // DI_FORCECOPY
	DriverPageAdded                = 0x04000000 // DI_DRIVERPAGE_ADDED
	UseClassInstallerSelectStrings = 0x08000000 // DI_USECI_SELECTSTRINGS
	OverrideInfFlags               = 0x10000000 // DI_OVERRIDE_INFFLAGS
	PropsNoChangeUsage             = 0x20000000 // DI_PROPS_NOCHANGEUSAGE, OBSOLETE
	NoSelectIcons                  = 0x40000000 // DI_NOSELECTICONS, OBSOLETE
	NoWriteIDs                     = 0x80000000 // DI_NOWRITE_IDS
)
