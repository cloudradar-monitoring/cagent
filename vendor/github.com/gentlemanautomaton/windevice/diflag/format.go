package diflag

// Format maps flags to their string representations.
type Format map[Value]string

// FormatGo maps values to Go-style constant strings.
var FormatGo = Format{
	ShowOEM:                        "ShowOEM",
	ShowCompat:                     "ShowCompat",
	ShowClass:                      "ShowClass",
	ShowAll:                        "ShowAll",
	NoVCP:                          "NoVCP",
	DidCompat:                      "DidCompat",
	DidClass:                       "DidClass",
	AutoAssignResources:            "AutoAssignResources",
	NeedRestart:                    "NeedRestart",
	NeedReboot:                     "NeedReboot",
	NoBrowse:                       "NoBrowse",
	MultipleManufacturers:          "MultipleManufacturers",
	Disabled:                       "Disabled",
	GeneralPageAdded:               "GeneralPageAdded",
	ResourcePageAdded:              "ResourcePageAdded",
	PropertiesChange:               "PropertiesChange",
	InfIsSorted:                    "InfIsSorted",
	EnumSingleInf:                  "EnumSingleInf",
	DoNotCallConfigManager:         "DoNotCallConfigManager",
	InstallDisabled:                "InstallDisabled",
	CompatFromClass:                "CompatFromClass",
	ClassInstallParams:             "ClassInstallParams",
	NoInstallerDefaultAction:       "NoInstallerDefaultAction",
	QuietInstall:                   "QuietInstall",
	NoFileCopy:                     "NoFileCopy",
	ForceCopy:                      "ForceCopy",
	DriverPageAdded:                "DriverPageAdded",
	UseClassInstallerSelectStrings: "UseClassInstallerSelectStrings",
	OverrideInfFlags:               "OverrideInfFlags",
	PropsNoChangeUsage:             "PropsNoChangeUsage",
	NoSelectIcons:                  "NoSelectIcons",
	NoWriteIDs:                     "NoWriteIDs",
}
