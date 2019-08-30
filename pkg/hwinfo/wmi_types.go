// +build windows

package hwinfo

type winMemoryType uint16

// https://docs.microsoft.com/en-us/windows/desktop/cimwin32prov/win32-physicalmemory
type win32_PhysicalMemory struct {
	Capacity   *uint64
	MemoryType *winMemoryType
}

// https://docs.microsoft.com/en-us/windows/desktop/cimwin32prov/win32-baseboard
type win32_BaseBoard struct {
	Manufacturer *string
	Product      *string
	Model        *string
	SerialNumber *string
}

// https://docs.microsoft.com/en-us/windows/desktop/cimwin32prov/win32-processor
type win32_Processor struct {
	Description               *string
	Name                      *string
	Manufacturer              *string
	NumberOfCores             *uint32
	NumberOfLogicalProcessors *uint32
}

const (
	monitorAvailabilityOther                 = 1
	monitorAvailabilityUnknown               = 2
	monitorAvailabilityRunningOrFullPower    = 3
	monitorAvailabilityWarning               = 4
	monitorAvailabilityInTest                = 5
	monitorAvailabilityNotApplicable         = 6
	monitorAvailabilityPowerOff              = 7
	monitorAvailabilityOffLine               = 8
	monitorAvailabilityOffDuty               = 9
	monitorAvailabilityDegraded              = 10
	monitorAvailabilityNotInstalled          = 11
	monitorAvailabilityInstallError          = 12
	monitorAvailabilityPowerSaveUnknown      = 13
	monitorAvailabilityPowerSaveLowPowerMode = 14
	monitorAvailabilityPowerSaveStandby      = 15
	monitorAvailabilityPowerCycle            = 16
	monitorAvailabilityPowerSaveWarning      = 17
	monitorAvailabilityPaused                = 18
	monitorAvailabilityNotReady              = 19
	monitorAvailabilityNotConfigured         = 20
	monitorAvailabilityQuiesced              = 21
)

// https://docs.microsoft.com/en-us/windows/desktop/cimwin32prov/win32-desktopmonitor
type win32_DesktopMonitor struct {
	Availability        *uint16
	Caption             *string
	Description         *string
	DeviceID            *string
	MonitorManufacturer *string
	Name                *string
	ScreenHeight        *uint32
	ScreenWidth         *uint32
	PNPDeviceID         *string
}

func (d *win32_DesktopMonitor) IsActive() bool {
	if d.Availability != nil {
		switch *d.Availability {
		case monitorAvailabilityRunningOrFullPower, monitorAvailabilityWarning, monitorAvailabilityInTest,
			monitorAvailabilityOffLine, monitorAvailabilityOffDuty, monitorAvailabilityDegraded,
			monitorAvailabilityPowerSaveUnknown, monitorAvailabilityPowerSaveLowPowerMode, monitorAvailabilityPowerSaveStandby,
			monitorAvailabilityPowerCycle, monitorAvailabilityPowerSaveWarning:
			return true
		default:
			return false
		}
	}
	return false
}

func (w winMemoryType) String() string {
	switch w {
	case 2:
		return "DRAM"
	case 3:
		return "Synchronous DRAM"
	case 4:
		return "Cache DRAM"
	case 5:
		return "EDO"
	case 6:
		return "EDRAM"
	case 7:
		return "VRAM"
	case 8:
		return "SRAM"
	case 9:
		return "RAM"
	case 10:
		return "ROM"
	case 11:
		return "FLASH"
	case 12:
		return "EEPROM"
	case 13:
		return "FEPROM"
	case 14:
		return "EPROM"
	case 15:
		return "CDRAM"
	case 16:
		return "3DRAM"
	case 17:
		return "SDRAM"
	case 18:
		return "SGRAM"
	case 19:
		return "RDRAM"
	case 20:
		return "DDR"
	case 21:
		return "DDR2"
	case 22:
		return "DDR2 FB-DIMM"
	case 24:
		return "DDR3"
	case 25:
		return "FBD2"
	default:
		return "unknown"
	}
}
