// +build !windows

package hwinfo

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"

	"github.com/cloudradar-monitoring/dmidecode"
	"github.com/sirupsen/logrus"
)

func isCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command", "-v", name)
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func fetchInventory() (map[string]interface{}, error) {
	// check dmidecode is present in the system
	if !isCommandAvailable("dmidecode") {
		return nil, ErrNotPresent
	}

	// expecting /etc/sudoers.d/cagent-dmidecode is present
	cmd := exec.Command("/bin/sh", "-c", "sudo dmidecode")

	buf := bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(&buf)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("hwinfo: execute dmidecode %s", err.Error())
	}

	dmi, err := dmidecode.Unmarshal(bufio.NewReader(&buf))
	if err != nil {
		return nil, fmt.Errorf("hwinfo: unmarshal dmi %s", err.Error())
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
		logrus.Errorf("hwinfo: failed fetching baseboard info, %s", err.Error())
	}

	var reqMem []dmidecode.ReqPhysicalMemoryArray
	if err = dmi.Get(&reqMem); err == nil {
		res["ram.number_of_modules"] = reqMem[0].NumberOfDevices
	} else if err != dmidecode.ErrNotFound {
		logrus.Errorf("hwinfo: failed fetching memory array info, %s", err.Error())
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
		logrus.Errorf("hwinfo: failed fetching memory device info, %s", err.Error())
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
		logrus.Errorf("hwinfo: failed fetching cpu info, %s", err.Error())
	}

	return res, nil
}
