package diflagex

// Format maps extended flags to their string representations.
type Format map[Value]string

// FormatGo maps values to Go-style constant strings.
var FormatGo = Format{
	UseOldInfSearch:              "UseOldInfSearch",
	AutoSelectRank0:              "AutoSelectRank0",
	ClassInstallerFailed:         "ClassInstallerFailed",
	DidInfoList:                  "DIDInfoList",
	DidCompatInfo:                "DIDCompatInfo",
	FilterClasses:                "FilterClasses",
	SetFailedInstall:             "SetFailedInstall",
	DeviceChange:                 "DeviceChange",
	AlwaysWriteIDs:               "AlwaysWriteIDs",
	PropertyChangePending:        "PropertyChangePending",
	AllowExcludedDrivers:         "AllowExcludedDrivers",
	NoUserInterfaceOnQueryRemove: "NoUserInterfaceOnQueryRemove",
	UseClassForCompat:            "UseClassForCompat",
	OldInfInClassList:            "OldInfInClassList",
	NoDriverRegistryModify:       "NoDriverRegistryModify",
	InSystemSetup:                "InSystemSetup",
	InetDriver:                   "InetDriver",
	AppendDriverList:             "AppendDriverList",
	PreInstallBackup:             "PreInstallBackup",
	BackupOnReplace:              "BackupOnReplace",
	DriverListFromURL:            "DriverListFromURL",
	Reserved1:                    "Reserved1",
	ExcludeOldInetDrivers:        "ExcludeOldInetDrivers",
	PowerPageAdded:               "PowerPageAdded",
	FilterSimilarDrivers:         "FilterSimilarDrivers",
	InstalledDriver:              "InstalledDriver",
	NoClassListNodeMerge:         "NoClassListNodeMerge",
	AltPlatformDriverSearch:      "AltPlatformDriverSearch",
	RestartDeviceOnly:            "RestartDeviceOnly",
	RecursiveSearch:              "RecursiveSearch",
	SearchPublishedInfs:          "SearchPublishedInfs",
}
