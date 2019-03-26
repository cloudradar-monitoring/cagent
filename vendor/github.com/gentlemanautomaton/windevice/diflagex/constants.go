package diflagex

// Windows device installation and enumeration flags.
//
// https://docs.microsoft.com/en-us/windows/desktop/api/setupapi/ns-setupapi-_sp_devinstall_params_w
const (
	UseOldInfSearch              = 0x00000001 // DI_FLAGSEX_USEOLDINFSEARCH, OBSOLETE
	AutoSelectRank0              = 0x00000002 // DI_FLAGSEX_AUTOSELECTRANK0, OBSOLETE
	ClassInstallerFailed         = 0x00000004 // DI_FLAGSEX_CI_FAILED
	FinishInstallAction          = 0x00000008 // DI_FLAGSEX_FINISHINSTALL_ACTION
	DidInfoList                  = 0x00000010 // DI_FLAGSEX_DIDINFOLIST
	DidCompatInfo                = 0x00000020 // DI_FLAGSEX_DIDCOMPATINFO
	FilterClasses                = 0x00000040 // DI_FLAGSEX_FILTERCLASSES
	SetFailedInstall             = 0x00000080 // DI_FLAGSEX_SETFAILEDINSTALL
	DeviceChange                 = 0x00000100 // DI_FLAGSEX_DEVICECHANGE
	AlwaysWriteIDs               = 0x00000200 // DI_FLAGSEX_ALWAYSWRITEIDS
	PropertyChangePending        = 0x00000400 // DI_FLAGSEX_PROPCHANGE_PENDING
	AllowExcludedDrivers         = 0x00000800 // DI_FLAGSEX_ALLOWEXCLUDEDDRVS
	NoUserInterfaceOnQueryRemove = 0x00001000 // DI_FLAGSEX_NOUIONQUERYREMOVE
	UseClassForCompat            = 0x00002000 // DI_FLAGSEX_USECLASSFORCOMPAT
	OldInfInClassList            = 0x00004000 // DI_FLAGSEX_OLDINF_IN_CLASSLIST, OBSOLETE
	NoDriverRegistryModify       = 0x00008000 // DI_FLAGSEX_NO_DRVREG_MODIFY
	InSystemSetup                = 0x00010000 // DI_FLAGSEX_IN_SYSTEM_SETUP
	InetDriver                   = 0x00020000 // DI_FLAGSEX_INET_DRIVER
	AppendDriverList             = 0x00040000 // DI_FLAGSEX_APPENDDRIVERLIST
	PreInstallBackup             = 0x00080000 // DI_FLAGSEX_PREINSTALLBACKUP
	BackupOnReplace              = 0x00100000 // DI_FLAGSEX_BACKUPONREPLACE
	DriverListFromURL            = 0x00200000 // DI_FLAGSEX_DRIVERLIST_FROM_URL
	Reserved1                    = 0x00400000 // DI_FLAGSEX_RESERVED1
	ExcludeOldInetDrivers        = 0x00800000 // DI_FLAGSEX_EXCLUDE_OLD_INET_DRIVERS
	PowerPageAdded               = 0x01000000 // DI_FLAGSEX_POWERPAGE_ADDED
	FilterSimilarDrivers         = 0x02000000 // DI_FLAGSEX_FILTERSIMILARDRIVERS
	InstalledDriver              = 0x04000000 // DI_FLAGSEX_INSTALLEDDRIVER
	NoClassListNodeMerge         = 0x08000000 // DI_FLAGSEX_NO_CLASSLIST_NODE_MERGE
	AltPlatformDriverSearch      = 0x10000000 // DI_FLAGSEX_ALTPLATFORM_DRVSEARCH
	RestartDeviceOnly            = 0x20000000 // DI_FLAGSEX_RESTART_DEVICE_ONLY
	RecursiveSearch              = 0x40000000 // DI_FLAGSEX_RECURSIVESEARCH
	SearchPublishedInfs          = 0x80000000 // DI_FLAGSEX_SEARCH_PUBLISHED_INFS
)
