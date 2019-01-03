// +build !windows

package hwinfo

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"

	"github.com/troian/dmidecode"
)

func fetchInventory() (map[string]interface{}, error) {
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

	var sysArray []dmidecode.ReqBaseBoard
	if err = dmi.Get(&sysArray); err != nil {
		return nil, err
	}

	var memArray []dmidecode.ReqPhysicalMemoryArray
	if err = dmi.Get(&memArray); err != nil {
		return nil, err
	}

	var memDevicesArr []dmidecode.ReqMemoryDevice
	if err = dmi.Get(&memDevicesArr); err != nil {
		return nil, err
	}

	var processorArr []dmidecode.ReqProcessor
	if err = dmi.Get(&processorArr); err != nil {
		return nil, err
	}

	res["baseboard.manufacturer"] = sysArray[0].Manufacturer
	res["baseboard.model"] = sysArray[0].Version
	res["baseboard.serial_number"] = sysArray[0].SerialNumber

	res["ram.number_of_modules"] = memArray[0].NumberOfDevices
	for i := range memDevicesArr {
		if memDevicesArr[i].Size == -1 {
			continue
		}
		res[fmt.Sprintf("ram.%d.size_B", i)] = memDevicesArr[i].Size
		res[fmt.Sprintf("ram.%d.type", i)] = memDevicesArr[i].Type
	}

	for i := range processorArr {
		res[fmt.Sprintf("cpu.%d.manufacturer", i)] = processorArr[i].Manufacturer
		res[fmt.Sprintf("cpu.%d.manufacturing_info", i)] = processorArr[i].Signature.String()
		res[fmt.Sprintf("cpu.%d.description", i)] = processorArr[i].Version
		res[fmt.Sprintf("cpu.%d.core_count", i)] = processorArr[i].CoreCount
		res[fmt.Sprintf("cpu.%d.core_enabled", i)] = processorArr[i].CoreEnabled
		res[fmt.Sprintf("cpu.%d.thread_count", i)] = processorArr[i].ThreadCount
	}

	return res, nil
}
