// +build windows

package hwinfo

import (
	"fmt"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/gentlemanautomaton/windevice"
	"github.com/gentlemanautomaton/windevice/deviceclass"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/wmi"
)

const wmiQueryTimeout = time.Second * 10

func fetchInventory() (map[string]interface{}, error) {
	res := make(map[string]interface{})

	errorCollector := common.ErrorCollector{}
	pciDevices := listPCIDevices(&errorCollector)
	if len(pciDevices) > 0 {
		res["pci.list"] = pciDevices
	}

	usbDevices := listUSBDevices(&errorCollector)
	if len(usbDevices) > 0 {
		res["usb.list"] = usbDevices
	}

	displays := listDisplays(&errorCollector)
	if len(displays) > 0 {
		res["display.list"] = displays
	}

	cpuInfo := getCPUInfo(&errorCollector)
	if len(cpuInfo) > 0 {
		res = common.MergeStringMaps(res, cpuInfo)
	}

	baseboardInfo := getBaseboardInfo(&errorCollector)
	if len(baseboardInfo) > 0 {
		res = common.MergeStringMaps(res, baseboardInfo)
	}

	ramInfo := getRamInfo(&errorCollector)
	if len(ramInfo) > 0 {
		res = common.MergeStringMaps(res, ramInfo)
	}

	return res, errorCollector.Combine()
}

func listPCIDevices(errs *common.ErrorCollector) []*pciDeviceInfo {
	query := windevice.DeviceQuery{
		Enumerator: "PCI",
		Flags:      deviceclass.Present,
	}

	result := make([]*pciDeviceInfo, 0)
	var index int
	err := query.Each(func(device windevice.Device) {
		fullDisplayingName, err := device.Description()
		if err != nil {
			errs.AddNewf("could not get PCI device description %s", err.Error())
			index++
			return
		}
		friendlyName, _ := device.FriendlyName()

		class, err := device.Class()
		if err != nil {
			errs.AddNewf("could not get PCI device class %s", err.Error())
		}
		location, err := device.LocationInformation()
		if err != nil {
			errs.AddNewf("could not get PCI device location %s", err.Error())
		}
		vendor, err := device.Manufacturer()
		if err != nil {
			errs.AddNewf("could not get PCI device vendor %s", err.Error())
		}

		description := class
		if friendlyName != "" {
			description = fmt.Sprintf("%s (%s)", friendlyName, class)
		}

		result = append(result, &pciDeviceInfo{
			Address:     location,
			DeviceType:  "",
			VendorName:  vendor,
			ProductName: fullDisplayingName,
			Description: description,
		})

		index++
	})

	if err != nil {
		errs.Add(err)
	}

	return result
}

func listUSBDevices(errs *common.ErrorCollector) []*usbDeviceInfo {
	enumerator := "USB"

	query := windevice.DeviceQuery{
		Enumerator: enumerator,
		Flags:      deviceclass.Present,
	}

	result := make([]*usbDeviceInfo, 0)
	var index int
	err := query.Each(func(device windevice.Device) {
		fullDisplayingName, err := device.Description()
		if err != nil {
			errs.AddNewf("could not get USB device description %s", err.Error())
			index++
			return
		}
		deviceInstanceID, _ := device.DeviceInstanceID()
		friendlyName, _ := device.FriendlyName()
		location, _ := device.LocationInformation()
		vendor, err := device.Manufacturer()
		if err != nil {
			errs.AddNewf("could not get USB device vendor %s", err.Error())
		}
		class, err := device.Class()
		if err != nil {
			errs.AddNewf("could not get USB device class %s", err.Error())
		}
		if class == "USB" {
			class = ""
		}

		descriptionParts := []string{fullDisplayingName, friendlyName, class}
		description := ""
		for _, part := range descriptionParts {
			if part != "" {
				if description != "" {
					description += " "
				}
				description += part
			}
		}

		result = append(result, &usbDeviceInfo{
			Address:     location,
			VendorName:  vendor,
			DeviceID:    string(deviceInstanceID),
			Description: description,
		})

		index++
	})

	if err != nil {
		errs.Add(err)
	}

	return result
}

func listDisplays(errs *common.ErrorCollector) []*monitorInfo {
	var monitors []win32_DesktopMonitor
	query := wmi.CreateQuery(&monitors, "")
	err := wmiutil.QueryWithTimeout(wmiQueryTimeout, query, &monitors)
	if err != nil {
		errs.AddNewf("request cpus info failed: %s", err.Error())
		return nil
	}

	result := make([]*monitorInfo, 0)
	for _, m := range monitors {
		if !m.IsActive() {
			continue
		}

		monitorID := ""
		if m.PNPDeviceID != nil {
			monitorID = *m.PNPDeviceID
		}

		description := ""
		if m.DeviceID != nil && *m.DeviceID != "" {
			description = *m.DeviceID
		}

		if m.Name != nil && *m.Name != "" {
			if description != "" {
				description += " - "
			}
			description += *m.Name
		}

		vendor := ""
		if m.MonitorManufacturer != nil {
			vendor = *m.MonitorManufacturer
		}

		resolutionStr := ""
		if m.ScreenWidth != nil && m.ScreenHeight != nil && *m.ScreenWidth > 0 {
			resolutionStr = fmt.Sprintf("%dx%d", *m.ScreenWidth, *m.ScreenHeight)
		}

		result = append(result, &monitorInfo{
			ID:          monitorID,
			Description: description,
			VendorName:  vendor,
			Size:        "",
			Resolution:  resolutionStr,
		})
	}
	return result
}

func getCPUInfo(errs *common.ErrorCollector) map[string]interface{} {
	res := make(map[string]interface{})

	var cpus []win32_Processor
	query := wmi.CreateQuery(&cpus, "")
	err := wmiutil.QueryWithTimeout(wmiQueryTimeout, query, &cpus)
	if err != nil {
		errs.AddNewf("request cpus info failed: %s", err.Error())
		return nil
	}

	for i := range cpus {
		res[fmt.Sprintf("cpu.%d.manufacturer", i)] = cpus[i].Manufacturer
		res[fmt.Sprintf("cpu.%d.manufacturing_info", i)] = cpus[i].Description
		res[fmt.Sprintf("cpu.%d.description", i)] = cpus[i].Name
		res[fmt.Sprintf("cpu.%d.core_count", i)] = cpus[i].NumberOfCores
		res[fmt.Sprintf("cpu.%d.core_enabled", i)] = cpus[i].NumberOfEnabledCore
		res[fmt.Sprintf("cpu.%d.thread_count", i)] = cpus[i].NumberOfLogicalProcessors
	}

	return res
}

func getBaseboardInfo(errs *common.ErrorCollector) map[string]interface{} {
	var baseBoard []win32_BaseBoard
	query := wmi.CreateQuery(&baseBoard, "")
	if err := wmiutil.QueryWithTimeout(wmiQueryTimeout, query, &baseBoard); err != nil {
		errs.AddNewf("request baseboard info failed: %s", err.Error())
		return nil
	}

	res := make(map[string]interface{})
	if len(baseBoard) > 0 {
		res["baseboard.manufacturer"] = baseBoard[0].Manufacturer
		res["baseboard.serial_number"] = baseBoard[0].SerialNumber
		res["baseboard.model"] = baseBoard[0].Product
	}

	return res
}

func getRamInfo(errs *common.ErrorCollector) map[string]interface{} {
	var ram []win32_PhysicalMemory
	query := wmi.CreateQuery(&ram, "")
	if err := wmi.Query(query, &ram); err != nil {
		errs.AddNewf("request ram info failed: %s", err.Error())
	}

	res := make(map[string]interface{})

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

	return res
}
