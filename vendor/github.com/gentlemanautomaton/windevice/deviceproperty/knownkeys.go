package deviceproperty

import "github.com/gentlemanautomaton/winguid"

// KnownKeys maps well known keys to names.
var KnownKeys = map[Key]string{
	Key{Category: winguid.New("80D81EA6-7473-4B0C-8216-EFC11A2C4C8B"), PropertyID: 2}:     "System.Devices.ModelId",
	Key{Category: winguid.New("656A3BB3-ECC0-43FD-8477-4AE0404A96CD"), PropertyID: 8194}:  "System.Devices.ModelName",
	Key{Category: winguid.New("656A3BB3-ECC0-43FD-8477-4AE0404A96CD"), PropertyID: 8195}:  "System.Devices.ModelNumber",
	Key{Category: winguid.New("656A3BB3-ECC0-43FD-8477-4AE0404A96CD"), PropertyID: 12288}: "System.Devices.FriendlyName",
	Key{Category: winguid.New("B725F130-47EF-101A-A5F1-02608C9EEBAC"), PropertyID: 10}:    "System.ItemNameDisplay",

	// Device status
	Key{Category: winguid.New("4340A6C5-93FA-4706-972C-7B648008A5A7"), PropertyID: 2}: "System.Devices.DevNodeStatus",
	Key{Category: winguid.New("4340A6C5-93FA-4706-972C-7B648008A5A7"), PropertyID: 3}: "System.Devices.ProblemCode",

	// Device relations
	Key{Category: winguid.New("4340A6C5-93FA-4706-972C-7B648008A5A7"), PropertyID: 4}:  "System.Devices.EjectionRelations",
	Key{Category: winguid.New("4340A6C5-93FA-4706-972C-7B648008A5A7"), PropertyID: 5}:  "System.Devices.RemovalRelations",
	Key{Category: winguid.New("4340A6C5-93FA-4706-972C-7B648008A5A7"), PropertyID: 6}:  "System.Devices.PowerRelations",
	Key{Category: winguid.New("4340A6C5-93FA-4706-972C-7B648008A5A7"), PropertyID: 7}:  "System.Devices.BusRelations",
	Key{Category: winguid.New("4340A6C5-93FA-4706-972C-7B648008A5A7"), PropertyID: 8}:  "System.Devices.Parent",
	Key{Category: winguid.New("4340A6C5-93FA-4706-972C-7B648008A5A7"), PropertyID: 9}:  "System.Devices.Children",
	Key{Category: winguid.New("4340A6C5-93FA-4706-972C-7B648008A5A7"), PropertyID: 10}: "System.Devices.Siblings",
	Key{Category: winguid.New("4340A6C5-93FA-4706-972C-7B648008A5A7"), PropertyID: 11}: "System.Devices.TransportRelations",

	// Network devices
	Key{Category: winguid.New("49CD1F76-5626-4B17-A4E8-18B4AA1A2213"), PropertyID: 7}: "System.Devices.NetworkName",
	Key{Category: winguid.New("49CD1F76-5626-4B17-A4E8-18B4AA1A2213"), PropertyID: 8}: "System.Devices.NetworkType",

	// Device model
	Key{Category: winguid.New("78C34FC8-104A-4ACA-9EA4-524D52996E57"), PropertyID: 39}: "System.Devices.Model",

	// Device instance
	Key{Category: winguid.New("78C34FC8-104A-4ACA-9EA4-524D52996E57"), PropertyID: 256}: "System.Devices.InstanceId",

	// Standard device properties (aka registry properties)
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 2}:  "System.Devices.Description",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 3}:  "System.Devices.HardwareIds",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 4}:  "System.Devices.CompatibleIds",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 6}:  "System.Devices.Service",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 9}:  "System.Devices.Class",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 10}: "System.Devices.ClassGuid",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 11}: "System.Devices.Driver",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 12}: "System.Devices.ConfigFlags",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 13}: "System.Devices.Manufacturer",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 14}: "System.Devices.FriendlyName",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 15}: "System.Devices.LocationInfo",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 16}: "System.Devices.PDOName",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 17}: "System.Devices.Capabilities",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 18}: "System.Devices.UINumber",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 19}: "System.Devices.UpperFilters",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 20}: "System.Devices.LowerFilters",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 21}: "System.Devices.BusTypeGuid",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 22}: "System.Devices.LegacyBusType",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 23}: "System.Devices.BusNumber",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 24}: "System.Devices.EnumeratorName",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 25}: "System.Devices.Security",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 26}: "System.Devices.SecuritySDS",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 27}: "System.Devices.DevType",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 28}: "System.Devices.Exclusive",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 29}: "System.Devices.Characteristics",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 30}: "System.Devices.Address",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 31}: "System.Devices.UINumberDescFormat",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 32}: "System.Devices.PowerData",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 33}: "System.Devices.RemovalPolicy",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 34}: "System.Devices.RemovalPolicyDefault",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 35}: "System.Devices.RemovalPolicyOverride",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 36}: "System.Devices.InstallState",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 37}: "System.Devices.LocationPaths",
	Key{Category: winguid.New("A45C254E-DF1C-4EFD-8020-67D146A850E0"), PropertyID: 38}: "System.Devices.BaseContainerId",

	// Device driver properties
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 2}:  "System.Drivers.AssemblyDate",
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 3}:  "System.Drivers.Version",
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 4}:  "System.Drivers.Description",
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 5}:  "System.Drivers.InfPath",
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 6}:  "System.Drivers.InfSection",
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 7}:  "System.Drivers.InfSectionExt",
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 8}:  "System.Drivers.MatchingDeviceId",
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 9}:  "System.Drivers.Provider",
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 10}: "System.Drivers.PropPageProvider",
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 11}: "System.Drivers.CoInstallers",
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 12}: "System.Drivers.ResourcePickerTags",
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 13}: "System.Drivers.ResourcePickerExceptions",
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 14}: "System.Drivers.Rank",
	Key{Category: winguid.New("A8B865DD-2E3D-4094-AD97-E593A70C75D6"), PropertyID: 15}: "System.Drivers.LogoLevel",

	// Other device properties
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 2}:  "System.Devices.NumaProximityDomain",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 3}:  "System.Devices.DHPRebalancePolicy",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 4}:  "System.Devices.NumaNode",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 5}:  "System.Devices.BusReportedDeviceDesc",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 5}:  "System.Devices.IsPresent",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 6}:  "System.Devices.HasProblem",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 7}:  "System.Devices.ConfigurationId",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 8}:  "System.Devices.ReportedDeviceIdsHash",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 9}:  "System.Devices.PhysicalDeviceLocation",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 10}: "System.Devices.BiosDeviceName",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 11}: "System.Devices.DriverProblemDesc",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 12}: "System.Devices.DebuggerSafe",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 13}: "System.Devices.PostInstallInProgress",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 14}: "System.Devices.Stack",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 15}: "System.Devices.ExtendedConfigurationIds",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 16}: "System.Devices.IsRebootRequired",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 17}: "System.Devices.FirmwareDate",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 18}: "System.Devices.FirmwareVersion",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 19}: "System.Devices.FirmwareRevision",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 20}: "System.Devices.DependencyProviders",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 21}: "System.Devices.DependencyDependents",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 22}: "System.Devices.SoftRestartSupported",
	Key{Category: winguid.New("540B947E-8B40-45BC-A8A2-6A0B894CBDA2"), PropertyID: 23}: "System.Devices.ExtendedAddress",
	Key{Category: winguid.New("83DA6326-97A6-4088-9453-A1923F573B29"), PropertyID: 6}:  "System.Devices.SessionId",

	// Device activity properties
	Key{Category: winguid.New("83DA6326-97A6-4088-9453-A1923F573B29"), PropertyID: 100}: "System.Devices.InstallDate",
	Key{Category: winguid.New("83DA6326-97A6-4088-9453-A1923F573B29"), PropertyID: 101}: "System.Devices.FirstInstallDate",
	Key{Category: winguid.New("83DA6326-97A6-4088-9453-A1923F573B29"), PropertyID: 102}: "System.Devices.LastArrivalDate",
	Key{Category: winguid.New("83DA6326-97A6-4088-9453-A1923F573B29"), PropertyID: 103}: "System.Devices.LastRemovalDate",
}
