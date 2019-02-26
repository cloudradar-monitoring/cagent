// +build windows

package hwinfo

import (
	"fmt"
	"time"

	"github.com/StackExchange/wmi"

	wmiutil "github.com/cloudradar-monitoring/cagent/pkg/wmi"
)

type winMemoryType uint16

const reqTimeout = time.Second * 10

type win32_PhysicalMemory struct {
	Capacity   *uint64
	MemoryType *winMemoryType

	/*
		Possible fields so we could use them later if we will need

		BankLabel     *string
		DataWidth     *uint16
		DeviceLocator *string
		InstallDate   *time.Time
		Manufacturer  *string
		MaxVoltage    *uint32
		MinVoltage    *uint32
		Model         *string
		PartNumber    *string
		SerialNumber  *string
		Speed         *uint32
		Status        *string
		TotalWidth    *uint16
	*/
}

type win32_BaseBoard struct {
	Manufacturer *string
	Product      *string
	Model        *string
	SerialNumber *string
}

type win32_Processor struct {
	Description               *string
	Name                      *string
	Manufacturer              *string
	NumberOfCores             *uint32
	NumberOfEnabledCore       *uint32
	NumberOfLogicalProcessors *uint32

	/*
		Possible fields so we could use them later if we will need

		DeviceID                  *string
		SerialNumber              *string
		SocketDesignation         *string
		MaxClockSpeed             *uint32
		ProcessorType             *uint16
	*/
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

func fetchInventory() (map[string]interface{}, error) {
	res := make(map[string]interface{})

	var errs []string

	var cpus []win32_Processor
	query := wmi.CreateQuery(&cpus, "")
	err := wmiutil.QueryWithTimeout(reqTimeout, query, &cpus)
	if err != nil {
		return nil, fmt.Errorf("hwinfo: request cpus info %s", err.Error())
	}

	for i := range cpus {
		res[fmt.Sprintf("cpu.%d.manufacturer", i)] = cpus[i].Manufacturer
		res[fmt.Sprintf("cpu.%d.manufacturing_info", i)] = cpus[i].Description
		res[fmt.Sprintf("cpu.%d.description", i)] = cpus[i].Name
		res[fmt.Sprintf("cpu.%d.core_count", i)] = cpus[i].NumberOfCores
		res[fmt.Sprintf("cpu.%d.core_enabled", i)] = cpus[i].NumberOfEnabledCore
		res[fmt.Sprintf("cpu.%d.thread_count", i)] = cpus[i].NumberOfLogicalProcessors
	}

	var baseBoard []win32_BaseBoard
	query = wmi.CreateQuery(&baseBoard, "")
	if err = wmiutil.QueryWithTimeout(reqTimeout, query, &baseBoard); err != nil {
		errs = append(errs, "request baseboard info: ", err.Error())
	}

	if len(baseBoard) > 0 {
		res["baseboard.manufacturer"] = baseBoard[0].Manufacturer
		res["baseboard.serial_number"] = baseBoard[0].SerialNumber
		res["baseboard.model"] = baseBoard[0].Product
	}

	var ram []win32_PhysicalMemory
	query = wmi.CreateQuery(&ram, "")
	if err = wmi.Query(query, &ram); err != nil {
		errs = append(errs, "request ram info: ", err.Error())
	}

	res["ram.number_of_modules"] = len(ram)
	for i := range ram {
		res[fmt.Sprintf("ram.%d.size_B", i)] = ram[i].Capacity
		memoryType := ram[i].MemoryType
		if memoryType == nil {
			res[fmt.Sprintf("ram.%d.type", i)] = nil
		} else {
			res[fmt.Sprintf("ram.%d.type", i)] = (*memoryType).String()
		}
	}

	if len(errs) > 0 {
		errString := "hwinfo: "

		for _, s := range errs {
			errString += s + ";"
		}

		return res, err
	}

	return res, nil
}
