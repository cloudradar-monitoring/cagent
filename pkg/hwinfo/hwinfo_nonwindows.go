// +build !windows

package hwinfo

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/dmidecode"
	"github.com/jaypipes/ghw"
	"github.com/vcraescu/go-xrandr"
)

var lsusbLineRegexp = regexp.MustCompile(`[0-9|a-z|A-Z|.|/|-|:|\[|\]|_|+| ]+`)

func isCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command", "-v", name)
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

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

	dmiDecodeResults := retrieveInfoUsingDmiDecode(&errorCollector)
	if len(dmiDecodeResults) > 0 {
		res = mergeStringMaps(&res, &dmiDecodeResults)
	}

	return res, errorCollector.Combine()
}

func listPCIDevices(errs *common.ErrorCollector) []*pciDeviceInfo {
	pciInfo, err := ghw.PCI()

	if err != nil {
		errs.New(err)
		return nil
	}

	devices := pciInfo.ListDevices()
	if len(devices) == 0 {
		errs.New(err)
		return nil
	}

	result := make([]*pciDeviceInfo, 0, len(devices))
	for _, device := range devices {
		vendor := device.Vendor
		product := device.Product

		deviceType := device.Subclass.Name
		if deviceType == "unknown" {
			deviceType = ""
		}
		deviceClassName := device.Class.Name
		if deviceClassName == "unknown" {
			deviceClassName = ""
		}

		if deviceType == "" {
			deviceType = deviceClassName
		} else if deviceClassName != "" && deviceClassName != deviceType {
			deviceType = fmt.Sprintf("%s (%s)", deviceClassName, deviceType)
		}

		description := device.ProgrammingInterface.Name
		if description == "unknown" || description == "Normal decode" {
			description = ""
		}
		result = append(result, &pciDeviceInfo{
			DeviceType:  deviceType,
			Address:     device.Address,
			VendorName:  vendor.Name,
			ProductName: product.Name,
			Description: description,
		})
	}
	return result
}

func listUSBDevices(errs *common.ErrorCollector) []*usbDeviceInfo {
	results := make([]*usbDeviceInfo, 0)
	reg := regexp.MustCompile(`[^:]+`)
	var lines []string

	cmd := exec.Command("lsusb")
	buf := bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(&buf)
	if err := cmd.Run(); err != nil {
		errs.New(err)
		return nil
	}

	outBytes, err := ioutil.ReadAll(bufio.NewReader(&buf))
	if err != nil {
		errs.New(err)
		return nil
	}

	lines = strings.Split(string(outBytes), "\n")
	for i := 0; i < len(lines); i++ {
		tokens := strings.Split(lines[i], " ")
		sanitizedTokens := make([]string, 0)
		for _, t := range tokens {
			if t != "" && t != "\t" {
				sanitizedTokens = append(sanitizedTokens, t)
			}
		}
		sanitizedTokensCount := len(sanitizedTokens)
		if sanitizedTokensCount < 6 {
			if sanitizedTokensCount > 0 {
				errs.Addf("unexpected lsusb command output: got %d tokens in line: %s", sanitizedTokensCount, lines[i])
			}
			continue
		}
		var description string
		for i := 6; i < sanitizedTokensCount; i++ {
			if i == sanitizedTokensCount-1 {
				description += sanitizedTokens[i]
			} else {
				description += sanitizedTokens[i] + " "
			}
		}
		busNum, err := strconv.Atoi(sanitizedTokens[1])
		if err != nil {
			errs.Addf("error while parsing bus number: %s. line: %s", err.Error(), lines[i])
			continue
		}
		devNum, err := strconv.Atoi(reg.FindString(sanitizedTokens[3]))
		if err != nil {
			errs.Addf("error while parsing device number: %s. line: %s", err.Error(), lines[i])
			continue
		}
		devID := lsusbLineRegexp.FindString(sanitizedTokens[5])
		results = append(results, &usbDeviceInfo{busNum, devNum, devID, description})
	}
	return results
}

func listDisplays(errs *common.ErrorCollector) []*monitorInfo {
	results := make([]*monitorInfo, 0)
	screens, err := xrandr.GetScreens()
	if err != nil {
		errs.New(err)
		return nil
	}
	for _, s := range screens {
		for _, m := range s.Monitors {
			physicalSizeStr := fmt.Sprintf("%dmm x %dmm", int(m.Size.Width), int(m.Size.Height))
			resolutionStr := fmt.Sprintf("%dx%d", int(m.Resolution.Width), int(m.Resolution.Height))
			results = append(results, &monitorInfo{
				ID:         m.ID,
				IsPrimary:  m.Primary,
				Size:       physicalSizeStr,
				Resolution: resolutionStr,
			})
		}
	}

	return results
}

func retrieveInfoUsingDmiDecode(errs *common.ErrorCollector) map[string]interface{} {
	if !isCommandAvailable("dmidecode") {
		errs.Add("dmidecode not present")
		return nil
	}

	// expecting /etc/sudoers.d/cagent-dmidecode is present
	cmd := exec.Command("/bin/sh", "-c", "sudo dmidecode")

	buf := bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(&buf)
	if err := cmd.Run(); err != nil {
		errs.Addf("execute dmidecode %s", err.Error())
		return nil
	}

	dmi, err := dmidecode.Unmarshal(bufio.NewReader(&buf))
	if err != nil {
		errs.Addf("unmarshal dmi %s", err.Error())
		return nil
	}

	res := make(map[string]interface{})

	// all below requests are based on parsed data returned by dmidecode.Unmarshal
	// refer to doc dmidecode.Get to get description of function behavior
	var reqSys []dmidecode.ReqBaseBoard
	if err = dmi.Get(&reqSys); err == nil {
		res["baseboard.manufacturer"] = reqSys[0].Manufacturer
		res["baseboard.model"] = reqSys[0].Version
		res["baseboard.serial_number"] = reqSys[0].SerialNumber
	} else if err != dmidecode.ErrNotFound {
		errs.Addf("failed fetching baseboard info, %s", err.Error())
	}

	var reqMem []dmidecode.ReqPhysicalMemoryArray
	if err = dmi.Get(&reqMem); err == nil {
		res["ram.number_of_modules"] = reqMem[0].NumberOfDevices
	} else if err != dmidecode.ErrNotFound {
		errs.Addf("failed fetching memory array info, %s", err.Error())
	}

	var reqMemDevs []dmidecode.ReqMemoryDevice
	if err = dmi.Get(&reqMemDevs); err == nil {
		for i := range reqMemDevs {
			if reqMemDevs[i].Size == -1 {
				continue
			}
			res[fmt.Sprintf("ram.%d.size_B", i)] = reqMemDevs[i].Size
			res[fmt.Sprintf("ram.%d.type", i)] = reqMemDevs[i].Type
		}
	} else if err != dmidecode.ErrNotFound {
		errs.Addf("failed fetching memory device info, %s", err.Error())
	}

	var reqCPU []dmidecode.ReqProcessor
	if err = dmi.Get(&reqCPU); err == nil {
		for i := range reqCPU {
			res[fmt.Sprintf("cpu.%d.manufacturer", i)] = reqCPU[i].Manufacturer
			res[fmt.Sprintf("cpu.%d.manufacturing_info", i)] = reqCPU[i].Signature.String()
			res[fmt.Sprintf("cpu.%d.description", i)] = reqCPU[i].Version
			res[fmt.Sprintf("cpu.%d.core_count", i)] = reqCPU[i].CoreCount
			res[fmt.Sprintf("cpu.%d.core_enabled", i)] = reqCPU[i].CoreEnabled
			res[fmt.Sprintf("cpu.%d.thread_count", i)] = reqCPU[i].ThreadCount
		}
	} else if err != dmidecode.ErrNotFound {
		errs.Addf("failed fetching cpu info, %s", err.Error())
	}

	return res
}
