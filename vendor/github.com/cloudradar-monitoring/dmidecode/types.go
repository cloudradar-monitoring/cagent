// +build !windows

package dmidecode

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ReqType uint8

const (
	TypeBIOSInfo ReqType = iota
	TypeSystem
	TypeBaseBoard
	TypeChassis
	TypeProcessor
	TypeMemoryController
	TypeMemoryModule
	TypeCache
	TypePortConnector
	TypeSystemSlots
	TypeOnBoardDevices
	TypeOEMStrings
	TypeSystemConfigurationOptions
	TypeBIOSLanguage
	TypeGroupAssociations
	TypeSystemEventLog
	TypePhysicalMemoryArray
	TypeMemoryDevice
	Type32BitMemoryError
	TypeMemoryArrayMappedAddress
	TypeMemoryDeviceMappedAddress
	TypeBuiltInPointingDevice
	TypePortableBattery
	TypeSystemReset
	TypeHardwareSecurity
	TypeSystemPowerControls
	TypeVoltageProbe
	TypeCoolingDevice
	TypeTemperatureProbe
	TypeElectricalCurrentProbe
	TypeOOBRemoteAccess
	TypeBootIntegrityServices
	TypeSystemBoot
	Type64BitMemoryError
	TypeManagementDevice
	TypeManagementDeviceComponent
	TypeManagementDeviceThresholdData
	TypeMemoryChannel
	TypeIPMIDevice
	TypePowerSupply
	TypeAdditionalInformation
	TypeOnBoardDevice
	TypeEndOfTable       = 127
	TypeVendorRangeBegin = 128
)

var (
	ErrPtrToSlice        = errors.New("dmidecode: parameter must be pointer to slice")
	ErrInvalidEntityType = errors.New("dmidecode: invalid entity type")
	ErrNotFound          = errors.New("dmidecode: DMI type not found")
)

type objType interface {
	ObjectType() ReqType
}

type unmarshaler interface {
	unmarshal(string) error
}

type SimpleCSV []string

type CPUSignature struct {
	raw      string
	Type     int
	Family   int
	Model    int
	Stepping int
}

type SizeType int64

var _ unmarshaler = (*SimpleCSV)(nil)
var _ unmarshaler = (*CPUSignature)(nil)
var _ unmarshaler = (*SizeType)(nil)

// ReqBiosInfo DMI type 0
type ReqBiosInfo struct {
	Vendor          string    `dmidecode:"Vendor"`
	Version         string    `dmidecode:"Version"`
	Address         string    `dmidecode:"Address"`
	BIOSRevision    string    `dmidecode:"BIOS Revision"`
	ReleaseDate     time.Time `dmidecode:"Release Date"`
	RuntimeSize     string    `dmidecode:"RuntimeSize"`
	RomSize         string    `dmidecode:"ROM Size"`
	Characteristics []string  `dmidecode:"Characteristics"`
}

// ReqSystem DMI type 1
type ReqSystem struct {
	Manufacturer string `dmidecode:"Manufacturer"`
	ProductName  string `dmidecode:"Product Name"`
	Version      string `dmidecode:"Version"`
	SerialNumber string `dmidecode:"Serial Number"`
	UUID         string `dmidecode:"UUID"`
	Family       string `dmidecode:"Family"`
}

// ReqBaseBoard DMI type 2
type ReqBaseBoard struct {
	Manufacturer           string   `dmidecode:"Manufacturer"`
	ProductName            string   `dmidecode:"Product Name"`
	SerialNumber           string   `dmidecode:"Serial Number"`
	Version                string   `dmidecode:"Version"`
	AssetTag               string   `dmidecode:"Asset Tag"`
	Features               []string `dmidecode:"Features"`
	LocationInChassis      string   `dmidecode:"Location In Chassis"`
	ChassisHandle          uint16   `dmidecode:"Chassis Handle" formatHint:"hex"`
	Type                   string   `dmidecode:"Type"`
	ContainedObjectHandles int      `dmidecode:"Contained Object Handles"`
}

// ReqSystem DMI type 3
type ReqChassis struct {
	Manufacturer       string `dmidecode:"Manufacturer"`
	Type               string `dmidecode:"Type"`
	Lock               string `dmidecode:"Lock"`
	Version            string `dmidecode:"Version"`
	SerialNumber       string `dmidecode:"Serial Number"`
	AssetTag           string `dmidecode:"Asset Tag"`
	BootUpState        string `dmidecode:"Boot-up State"`
	PowerSupplyState   string `dmidecode:"Power Supply State"`
	ThermalState       string `dmidecode:"Thermal State"`
	SecurityStatus     string `dmidecode:"Security Status"`
	OEMInformation     string `dmidecode:"OEM Information"`
	Height             string `dmidecode:"Height"`
	NumberOfPowerCords string `dmidecode:"Number Of Power Cords"`
	ContainedElements  int    `dmidecode:"Contained Elements"`
	SKUNumber          string `dmidecode:"SKU Number"`
}

// ReqProcessor DMI type 4
type ReqProcessor struct {
	SocketDesignation string       `dmidecode:"Socket Designation"`
	Type              string       `dmidecode:"Type"`
	Family            string       `dmidecode:"Family"`
	Manufacturer      string       `dmidecode:"Manufacturer"`
	ID                string       `dmidecode:"ID"`
	Signature         CPUSignature `dmidecode:"Signature"`
	Flags             []string     `dmidecode:"Flags"`
	Version           string       `dmidecode:"Version"`
	Voltage           string       `dmidecode:"Voltage"`
	ExternalClock     string       `dmidecode:"External Clock"`
	MaxSpeed          string       `dmidecode:"Max Speed"`
	CurrentSpeed      string       `dmidecode:"Current Speed"`
	Status            string       `dmidecode:"Status"`
	Upgrade           string       `dmidecode:"Upgrade"`
	L1CacheHandle     uint16       `dmidecode:"L1 Cache Handle"`
	L2CacheHandle     uint16       `dmidecode:"L2 Cache Handle"`
	L3CacheHandle     uint16       `dmidecode:"L3 Cache Handle"`
	SerialNumber      string       `dmidecode:"Serial Number"`
	AssetTag          string       `dmidecode:"Asset Tag"`
	PartNumber        string       `dmidecode:"Part Number"`
	CoreCount         int          `dmidecode:"Core Count"`
	CoreEnabled       int          `dmidecode:"Core Enabled"`
	ThreadCount       int          `dmidecode:"Thread Count"`
	Characteristics   []string     `dmidecode:"Characteristics"`
}

// ReqMemoryController DMI type 5
type ReqMemoryController struct {
}

// ReqMemoryModule DMI type 6
type ReqMemoryModule struct {
}

// ReqCache DMI type 7
type ReqCache struct {
	SocketDesignation   string    `dmidecode:"Socket Designation"`
	Configuration       SimpleCSV `dmidecode:"Configuration"`
	OperationalMode     string    `dmidecode:"Operational Mode"`
	Location            string    `dmidecode:"Location"`
	InstalledSize       string    `dmidecode:"Installed Size"`
	MaximumSize         string    `dmidecode:"Maximum Size"`
	SupportedSRAMTypes  []string  `dmidecode:"Supported SRAM Types"`
	InstalledSRAMType   string    `dmidecode:"Installed SRAM Type"`
	Speed               string    `dmidecode:"Speed"`
	ErrorCorrectionType string    `dmidecode:"Error Correction Type"`
	SystemType          string    `dmidecode:"System Type"`
	Associativity       string    `dmidecode:"Associativity"`
}

// ReqPortConnector DMI type 8
type ReqPortConnector struct {
	InternalReferenceDesignator string `dmidecode:"Internal Reference Designator"`
	InternalConnectorType       string `dmidecode:"Internal Connector Type"`
	ExternalReferenceDesignator string `dmidecode:"External Reference Designator"`
	ExternalConnectorType       string `dmidecode:"External Connector Type"`
	PortType                    string `dmidecode:"Port Type"`
}

// ReqSystemSlots DMI type 9
type ReqSystemSlots struct {
	Designation     string   `dmidecode:"Designation"`
	Type            string   `dmidecode:"Type"`
	CurrentUsage    string   `dmidecode:"Current Usage"`
	Length          string   `dmidecode:"Length"`
	ID              int      `dmidecode:"ID"`
	Characteristics []string `dmidecode:"Characteristics"`
	BusAddress      string   `dmidecode:"Bus Address"`
}

// ReqOnBoardDevices DMI type 10
type ReqOnBoardDevices struct {
}

// ReqOEMStrings DMI type 11
type ReqOEMStrings struct {
}

// ReqSystemConfigurationOptions DMI type 12
type ReqSystemConfigurationOptions struct {
}

// ReqBIOSLanguage DMI type 13
type ReqBIOSLanguage struct {
}

// ReqGroupAssociations DMI type 14
type ReqGroupAssociations struct {
}

// ReqSystemEventLog DMI type 15
type ReqSystemEventLog struct {
}

// ReqPhysicalMemoryArray DMI type 16
type ReqPhysicalMemoryArray struct {
	Location               string `dmidecode:"Location"`
	Use                    string `dmidecode:"Use"`
	ErrorCorrection        string `dmidecode:"Error Correction"`
	MaximumCapacity        string `dmidecode:"Maximum Capacity"`
	ErrorInformationHandle string `dmidecode:"Error Information Handle"`
	NumberOfDevices        int    `dmidecode:"Number Of Devices"`
}

// ReqMemoryDevice DMI type 17
type ReqMemoryDevice struct {
	ArrayHandle            uint16   `dmidecode:"Array Handle"`
	ErrorInformationHandle string   `dmidecode:"Error Information Handle"`
	TotalWidth             string   `dmidecode:"Total Width"`
	DataWidth              string   `dmidecode:"Data Width"`
	Size                   SizeType `dmidecode:"Size" comment:"-1 means module not installed"`
	FormFactor             string   `dmidecode:"Form Factor"`
	Set                    string   `dmidecode:"Set"`
	Locator                string   `dmidecode:"Locator"`
	BankLocator            string   `dmidecode:"Bank Locator"`
	Type                   string   `dmidecode:"Type"`
	TypeDetail             string   `dmidecode:"Type Detail"`
	Speed                  string   `dmidecode:"Speed"`
	Manufacturer           string   `dmidecode:"Manufacturer"`
	SerialNumber           string   `dmidecode:"Serial Number"`
	AssetTag               string   `dmidecode:"Asset Tag"`
	PartNumber             string   `dmidecode:"Part Number"`
	Rank                   int      `dmidecode:"Rank"`
	ConfiguredClockSpeed   string   `dmidecode:"Configured Clock Speed"`
	MinimumVoltage         string   `dmidecode:"Minimum Voltage"`
	MaximumVoltage         string   `dmidecode:"Maximum Voltage"`
	ConfiguredVoltage      string   `dmidecode:"Configured Voltage"`
}

// Req32BitMemoryError DMI type 18
type Req32BitMemoryError struct {
}

// ReqMemoryArrayMappedAddress DMI type 19
type ReqMemoryArrayMappedAddress struct {
}

// ReqMemoryDeviceMappedAddress DMI type 20
type ReqMemoryDeviceMappedAddress struct {
}

// ReqBuiltInPointingDevice DMI type 21
type ReqBuiltInPointingDevice struct {
}

// ReqPortableBattery DMI type 22
type ReqPortableBattery struct {
}

// ReqSystemReset DMI type 23
type ReqSystemReset struct {
}

// ReqHardwareSecurity DMI type 24
type ReqHardwareSecurity struct {
}

// ReqSystemPowerControls DMI type 25
type ReqSystemPowerControls struct {
}

// ReqVoltageProbe DMI type 26
type ReqVoltageProbe struct {
}

// ReqCoolingDevice DMI type 27
type ReqCoolingDevice struct {
}

// ReqTemperatureProbe DMI type 28
type ReqTemperatureProbe struct {
}

// ReqElectricalCurrentProbe DMI type 29
type ReqElectricalCurrentProbe struct {
}

// ReqOOBRemoteAccess DMI type 30
type ReqOOBRemoteAccess struct {
}

// ReqBootIntegrityServices DMI type 31
type ReqBootIntegrityServices struct {
}

// ReqSystemBoot DMI type 32
type ReqSystemBoot struct {
}

// Req64BitMemoryError DMI type 33
type Req64BitMemoryError struct {
}

// ReqManagementDevice DMI type 34
type ReqManagementDevice struct {
}

// ReqManagementDeviceComponent DMI type 35
type ReqManagementDeviceComponent struct {
}

// ReqManagementDeviceThresholdData DMI type 36
type ReqManagementDeviceThresholdData struct {
}

// ReqMemoryChannel DMI type 37
type ReqMemoryChannel struct {
}

// ReqIPMIDevice DMI type 38
type ReqIPMIDevice struct {
}

// ReqPowerSupply DMI type 39
type ReqPowerSupply struct {
}

// ReqAdditionalInformation DMI type 40
type ReqAdditionalInformation struct {
}

// ReqOnBoardDevice DMI type 41
type ReqOnBoardDevice struct {
}

// ReqOemSpecificType
type ReqOemSpecificType struct {
	HeaderAndData []byte   `dmidecode:"Header and Data"`
	Strings       []string `dmidecode:"Strings"`
}

// ReqOnBoardDevice DMI type 127
type ReqEndOfTable struct {
}

// not that much time spent on this part actually. Multiline selection is the trick

var _ objType = (*ReqBiosInfo)(nil)
var _ objType = (*ReqSystem)(nil)
var _ objType = (*ReqBaseBoard)(nil)
var _ objType = (*ReqChassis)(nil)
var _ objType = (*ReqProcessor)(nil)
var _ objType = (*ReqMemoryController)(nil)
var _ objType = (*ReqMemoryModule)(nil)
var _ objType = (*ReqCache)(nil)
var _ objType = (*ReqPortConnector)(nil)
var _ objType = (*ReqSystemSlots)(nil)
var _ objType = (*ReqOnBoardDevices)(nil)
var _ objType = (*ReqOEMStrings)(nil)
var _ objType = (*ReqSystemConfigurationOptions)(nil)
var _ objType = (*ReqBIOSLanguage)(nil)
var _ objType = (*ReqGroupAssociations)(nil)
var _ objType = (*ReqSystemEventLog)(nil)
var _ objType = (*ReqPhysicalMemoryArray)(nil)
var _ objType = (*ReqMemoryDevice)(nil)
var _ objType = (*Req32BitMemoryError)(nil)
var _ objType = (*ReqMemoryArrayMappedAddress)(nil)
var _ objType = (*ReqMemoryDeviceMappedAddress)(nil)
var _ objType = (*ReqBuiltInPointingDevice)(nil)
var _ objType = (*ReqPortableBattery)(nil)
var _ objType = (*ReqSystemReset)(nil)
var _ objType = (*ReqHardwareSecurity)(nil)
var _ objType = (*ReqSystemPowerControls)(nil)
var _ objType = (*ReqVoltageProbe)(nil)
var _ objType = (*ReqCoolingDevice)(nil)
var _ objType = (*ReqTemperatureProbe)(nil)
var _ objType = (*ReqElectricalCurrentProbe)(nil)
var _ objType = (*ReqOOBRemoteAccess)(nil)
var _ objType = (*ReqBootIntegrityServices)(nil)
var _ objType = (*ReqSystemBoot)(nil)
var _ objType = (*Req64BitMemoryError)(nil)
var _ objType = (*ReqManagementDevice)(nil)
var _ objType = (*ReqManagementDeviceComponent)(nil)
var _ objType = (*ReqManagementDeviceThresholdData)(nil)
var _ objType = (*ReqMemoryChannel)(nil)
var _ objType = (*ReqIPMIDevice)(nil)
var _ objType = (*ReqPowerSupply)(nil)
var _ objType = (*ReqAdditionalInformation)(nil)
var _ objType = (*ReqOnBoardDevice)(nil)
var _ objType = (*ReqOemSpecificType)(nil)
var _ objType = (*ReqEndOfTable)(nil)

func newReqBiosInfo() interface{}                      { return &ReqBiosInfo{} }
func newReqSystem() interface{}                        { return &ReqSystem{} }
func newReqBaseBoard() interface{}                     { return &ReqBaseBoard{} }
func newReqChassis() interface{}                       { return &ReqChassis{} }
func newReqProcessor() interface{}                     { return &ReqProcessor{} }
func newReqMemoryController() interface{}              { return &ReqMemoryController{} }
func newReqMemoryModule() interface{}                  { return &ReqMemoryModule{} }
func newReqCache() interface{}                         { return &ReqCache{} }
func newReqPortConnector() interface{}                 { return &ReqPortConnector{} }
func newReqSystemSlots() interface{}                   { return &ReqSystemSlots{} }
func newReqOnBoardDevices() interface{}                { return &ReqOnBoardDevices{} }
func newReqOEMStrings() interface{}                    { return &ReqOEMStrings{} }
func newReqSystemConfigurationOptions() interface{}    { return &ReqSystemConfigurationOptions{} }
func newReqBIOSLanguage() interface{}                  { return &ReqBIOSLanguage{} }
func newReqGroupAssociations() interface{}             { return &ReqGroupAssociations{} }
func newReqSystemEventLog() interface{}                { return &ReqSystemEventLog{} }
func newReqPhysicalMemoryArray() interface{}           { return &ReqPhysicalMemoryArray{} }
func newReqMemoryDevice() interface{}                  { return &ReqMemoryDevice{} }
func newReq32BitMemoryError() interface{}              { return &Req32BitMemoryError{} }
func newReqMemoryArrayMappedAddress() interface{}      { return &ReqMemoryArrayMappedAddress{} }
func newReqMemoryDeviceMappedAddress() interface{}     { return &ReqMemoryDeviceMappedAddress{} }
func newReqBuiltInPointingDevice() interface{}         { return &ReqBuiltInPointingDevice{} }
func newReqPortableBattery() interface{}               { return &ReqPortableBattery{} }
func newReqSystemReset() interface{}                   { return &ReqSystemReset{} }
func newReqHardwareSecurity() interface{}              { return &ReqHardwareSecurity{} }
func newReqSystemPowerControls() interface{}           { return &ReqSystemPowerControls{} }
func newReqVoltageProbe() interface{}                  { return &ReqVoltageProbe{} }
func newReqCoolingDevice() interface{}                 { return &ReqCoolingDevice{} }
func newReqTemperatureProbe() interface{}              { return &ReqTemperatureProbe{} }
func newReqElectricalCurrentProbe() interface{}        { return &ReqElectricalCurrentProbe{} }
func newReqOOBRemoteAccess() interface{}               { return &ReqOOBRemoteAccess{} }
func newReqBootIntegrityServices() interface{}         { return &ReqBootIntegrityServices{} }
func newReqSystemBoot() interface{}                    { return &ReqSystemBoot{} }
func newReq64BitMemoryError() interface{}              { return &Req64BitMemoryError{} }
func newReqManagementDevice() interface{}              { return &ReqManagementDevice{} }
func newReqManagementDeviceComponent() interface{}     { return &ReqManagementDeviceComponent{} }
func newReqManagementDeviceThresholdData() interface{} { return &ReqManagementDeviceThresholdData{} }
func newReqMemoryChannel() interface{}                 { return &ReqMemoryChannel{} }
func newReqIPMIDevice() interface{}                    { return &ReqIPMIDevice{} }
func newReqPowerSupply() interface{}                   { return &ReqPowerSupply{} }
func newReqAdditionalInformation() interface{}         { return &ReqAdditionalInformation{} }
func newReqOnBoardDevice() interface{}                 { return &ReqOnBoardDevice{} }
func newReqOemSpecificType() interface{}               { return &ReqOemSpecificType{} }
func newReqEndOfTable() interface{}                    { return &ReqEndOfTable{} }

func (req *ReqBiosInfo) ObjectType() ReqType                   { return TypeBIOSInfo }
func (req *ReqSystem) ObjectType() ReqType                     { return TypeSystem }
func (req *ReqBaseBoard) ObjectType() ReqType                  { return TypeBaseBoard }
func (req *ReqChassis) ObjectType() ReqType                    { return TypeChassis }
func (req *ReqProcessor) ObjectType() ReqType                  { return TypeProcessor }
func (req *ReqMemoryController) ObjectType() ReqType           { return TypeMemoryController }
func (req *ReqMemoryModule) ObjectType() ReqType               { return TypeMemoryModule }
func (req *ReqCache) ObjectType() ReqType                      { return TypeCache }
func (req *ReqPortConnector) ObjectType() ReqType              { return TypePortConnector }
func (req *ReqSystemSlots) ObjectType() ReqType                { return TypeSystemSlots }
func (req *ReqOnBoardDevices) ObjectType() ReqType             { return TypeOnBoardDevices }
func (req *ReqOEMStrings) ObjectType() ReqType                 { return TypeOEMStrings }
func (req *ReqSystemConfigurationOptions) ObjectType() ReqType { return TypeSystemConfigurationOptions }
func (req *ReqBIOSLanguage) ObjectType() ReqType               { return TypeBIOSLanguage }
func (req *ReqGroupAssociations) ObjectType() ReqType          { return TypeGroupAssociations }
func (req *ReqSystemEventLog) ObjectType() ReqType             { return TypeSystemEventLog }
func (req *ReqPhysicalMemoryArray) ObjectType() ReqType        { return TypePhysicalMemoryArray }
func (req *ReqMemoryDevice) ObjectType() ReqType               { return TypeMemoryDevice }
func (req *Req32BitMemoryError) ObjectType() ReqType           { return Type32BitMemoryError }
func (req *ReqMemoryArrayMappedAddress) ObjectType() ReqType   { return TypeMemoryArrayMappedAddress }
func (req *ReqMemoryDeviceMappedAddress) ObjectType() ReqType  { return TypeMemoryDeviceMappedAddress }
func (req *ReqBuiltInPointingDevice) ObjectType() ReqType      { return TypeBuiltInPointingDevice }
func (req *ReqPortableBattery) ObjectType() ReqType            { return TypePortableBattery }
func (req *ReqSystemReset) ObjectType() ReqType                { return TypeSystemReset }
func (req *ReqHardwareSecurity) ObjectType() ReqType           { return TypeHardwareSecurity }
func (req *ReqSystemPowerControls) ObjectType() ReqType        { return TypeSystemPowerControls }
func (req *ReqVoltageProbe) ObjectType() ReqType               { return TypeVoltageProbe }
func (req *ReqCoolingDevice) ObjectType() ReqType              { return TypeCoolingDevice }
func (req *ReqTemperatureProbe) ObjectType() ReqType           { return TypeTemperatureProbe }
func (req *ReqElectricalCurrentProbe) ObjectType() ReqType     { return TypeElectricalCurrentProbe }
func (req *ReqOOBRemoteAccess) ObjectType() ReqType            { return TypeOOBRemoteAccess }
func (req *ReqBootIntegrityServices) ObjectType() ReqType      { return TypeBootIntegrityServices }
func (req *ReqSystemBoot) ObjectType() ReqType                 { return TypeSystemBoot }
func (req *Req64BitMemoryError) ObjectType() ReqType           { return Type64BitMemoryError }
func (req *ReqManagementDevice) ObjectType() ReqType           { return TypeManagementDevice }
func (req *ReqManagementDeviceComponent) ObjectType() ReqType  { return TypeManagementDeviceComponent }
func (req *ReqManagementDeviceThresholdData) ObjectType() ReqType {
	return TypeManagementDeviceThresholdData
}
func (req *ReqMemoryChannel) ObjectType() ReqType         { return TypeMemoryChannel }
func (req *ReqIPMIDevice) ObjectType() ReqType            { return TypeIPMIDevice }
func (req *ReqPowerSupply) ObjectType() ReqType           { return TypePowerSupply }
func (req *ReqAdditionalInformation) ObjectType() ReqType { return TypeAdditionalInformation }
func (req *ReqOnBoardDevice) ObjectType() ReqType         { return TypeOnBoardDevice }
func (req *ReqOemSpecificType) ObjectType() ReqType       { return TypeVendorRangeBegin }
func (req *ReqEndOfTable) ObjectType() ReqType            { return TypeEndOfTable }

func (un *CPUSignature) unmarshal(data string) error {
	un.raw = data
	return nil
}

func (un *CPUSignature) String() string {
	return un.raw
}

func (un *SimpleCSV) unmarshal(data string) error {
	vals := strings.Split(data, ",")
	for _, val := range vals {
		*un = append(*un, strings.Trim(val, " "))
	}
	return nil
}

func (un *SizeType) unmarshal(data string) error {
	if data == "No Module Installed" {
		*un = -1
		return nil
	}
	vals := sizeTypeRegex.FindStringSubmatch(data)
	if len(vals) != 3 {
		return errors.New("dmidecode: invalid size type")
	}

	val, e := strconv.ParseUint(vals[1], 10, 64)
	if e != nil {
		return fmt.Errorf("dmidecode: parse size value %s", e.Error())
	}

	switch vals[2] {
	case "GB":
		val *= 1024
		fallthrough
	case "MB":
		val *= 1024
		fallthrough
	case "KB":
		val *= 1024
		fallthrough
	case "B":
	default:
		return errors.New("dmidecode: invalid size type")
	}

	*un = SizeType(val)
	return nil
}
