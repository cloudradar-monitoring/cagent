// +build !windows

package hwinfo

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cloudradar-monitoring/dmidecode"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

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

	pciDevices, err := listPCIDevices()
	errorCollector.Add(err)
	if len(pciDevices) > 0 {
		res["pci.list"] = pciDevices
	}

	usbDevices, err := listUSBDevices()
	errorCollector.Add(err)
	if len(usbDevices) > 0 {
		res["usb.list"] = usbDevices
	}

	displays, err := listDisplays()
	errorCollector.Add(err)
	if len(displays) > 0 {
		res["displays.list"] = displays
	}

	dmiDecodeResults, err := retrieveInfoUsingDmiDecode()
	errorCollector.Add(err)
	if len(dmiDecodeResults) > 0 {
		res = common.MergeStringMaps(res, dmiDecodeResults)
	}

	return res, errorCollector.Combine()
}

func retrieveInfoUsingDmiDecode() (map[string]interface{}, error) {
	if !isCommandAvailable("dmidecode") {
		log.Infof("[HWINFO] dmidecode is not present. Skipping retrieval of baseboard, CPU and RAM info...")
		return nil, nil
	}

	var dmidecodeCmd []string

	// run dmidecode as command argument to shell
	dmidecodeCmd = append(dmidecodeCmd, "-c")

	if runtime.GOOS != "darwin" {
		// expecting 'sudo' package is installed and /etc/sudoers.d/cagent-dmidecode is present
		dmidecodeCmd = append(dmidecodeCmd, "sudo")
	}
	dmidecodeCmd = append(dmidecodeCmd, "dmidecode")

	cmd := exec.Command("/bin/sh", dmidecodeCmd...)

	stdoutBuffer := bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(&stdoutBuffer)

	stderrBuffer := bytes.Buffer{}
	cmd.Stderr = bufio.NewWriter(&stderrBuffer)

	if err := cmd.Run(); err != nil {
		stderrBytes, _ := ioutil.ReadAll(bufio.NewReader(&stderrBuffer))
		stderr := string(stderrBytes)
		if strings.Contains(stderr, "/dev/mem: Operation not permitted") {
			log.Infof("[HWINFO] there was an error while executing '%s': %s\nProbably 'CONFIG_STRICT_DEVMEM' kernel configuration option is enabled. Please refer to kernel configuration manual.", dmidecodeCmd, stderr)
			return nil, nil
		}
		return nil, errors.Wrap(err, "execute dmidecode")
	}

	dmi, err := dmidecode.Unmarshal(bufio.NewReader(&stdoutBuffer))
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal dmi")
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
		log.WithError(err).Info("[HWINFO] failed fetching baseboard info")
	}

	var reqMem []dmidecode.ReqPhysicalMemoryArray
	if err = dmi.Get(&reqMem); err == nil {
		res["ram.number_of_modules"] = reqMem[0].NumberOfDevices
	} else if err != dmidecode.ErrNotFound {
		log.WithError(err).Info("[HWINFO] failed fetching memory array info")
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
		log.WithError(err).Info("[HWINFO] failed fetching memory device info")
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
		log.WithError(err).Info("[HWINFO] failed fetching cpu info")
	}

	return res, nil
}
