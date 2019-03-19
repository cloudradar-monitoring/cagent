package difunc

// Windows device installation functions.
//
// https://docs.microsoft.com/en-us/windows-hardware/drivers/install/using-device-installation-functions
const (
	SelectDevice                  = 0x00000001 // DIF_SELECTDEVICE
	InstallDevice                 = 0x00000002 // DIF_INSTALLDEVICE
	AssignResources               = 0x00000003 // DIF_ASSIGNRESOURCES
	Properties                    = 0x00000004 // DIF_PROPERTIES
	Remove                        = 0x00000005 // DIF_REMOVE
	FirstTimeSetup                = 0x00000006 // DIF_FIRSTTIMESETUP
	FoundDevice                   = 0x00000007 // DIF_FOUNDDEVICE
	SelectClassDrivers            = 0x00000008 // DIF_SELECTCLASSDRIVERS
	ValidateClassDrivers          = 0x00000009 // DIF_VALIDATECLASSDRIVERS
	InstallClassDrivers           = 0x0000000A // DIF_INSTALLCLASSDRIVERS
	CalcDiskSpace                 = 0x0000000B // DIF_CALCDISKSPACE
	DestroyPrivateData            = 0x0000000C // DIF_DESTROYPRIVATEDATA
	ValidateDriver                = 0x0000000D // DIF_VALIDATEDRIVER
	Detect                        = 0x0000000F // DIF_DETECT
	InstallWizard                 = 0x00000010 // DIF_INSTALLWIZARD
	DestroyWizardData             = 0x00000011 // DIF_DESTROYWIZARDDATA
	PropertyChange                = 0x00000012 // DIF_PROPERTYCHANGE
	EnableClass                   = 0x00000013 // DIF_ENABLECLASS
	DetectVerify                  = 0x00000014 // DIF_DETECTVERIFY
	InstallDeviceFiles            = 0x00000015 // DIF_INSTALLDEVICEFILES
	Unremove                      = 0x00000016 // DIF_UNREMOVE
	SelectBestCompatDrv           = 0x00000017 // DIF_SELECTBESTCOMPATDRV
	AllowInstall                  = 0x00000018 // DIF_ALLOW_INSTALL
	RegisterDevice                = 0x00000019 // DIF_REGISTERDEVICE
	NewDeviceWizardPreselect      = 0x0000001A // DIF_NEWDEVICEWIZARD_PRESELECT
	NewDeviceWizardSelect         = 0x0000001B // DIF_NEWDEVICEWIZARD_SELECT
	NewDeviceWizardPreanalyze     = 0x0000001C // DIF_NEWDEVICEWIZARD_PREANALYZE
	NewDeviceWizardPostanalyze    = 0x0000001D // DIF_NEWDEVICEWIZARD_POSTANALYZE
	NewDeviceWizardFinishInstall  = 0x0000001E // DIF_NEWDEVICEWIZARD_FINISHINSTALL
	Unused1                       = 0x0000001F // DIF_UNUSED1
	InstallInterfaces             = 0x00000020 // DIF_INSTALLINTERFACES
	DetectCancel                  = 0x00000021 // DIF_DETECTCANCEL
	RegisterCoInstallers          = 0x00000022 // DIF_REGISTER_COINSTALLERS
	AddPropertyPageAdvanced       = 0x00000023 // DIF_ADDPROPERTYPAGE_ADVANCED
	AddPropertyPageBasic          = 0x00000024 // DIF_ADDPROPERTYPAGE_BASIC
	Reserved1                     = 0x00000025 // DIF_RESERVED1
	Troubleshooter                = 0x00000026 // DIF_TROUBLESHOOTER
	PowerMessageWake              = 0x00000027 // DIF_POWERMESSAGEWAKE
	AddRemotePropertyPageAdvanced = 0x00000028 // DIF_ADDREMOTEPROPERTYPAGE_ADVANCED
	UpdateDriverUI                = 0x00000029 // DIF_UPDATEDRIVER_UI
	FinishInstallAction           = 0x0000002A // DIF_FINISHINSTALL_ACTION
	Reserved2                     = 0x00000030 // DIF_RESERVED2
)
