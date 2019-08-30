// +build darwin

package hwinfo

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

type cpuInfo struct {
	manufacturer      string
	manufacturingInfo string
	description       string
	coreCount         string
	threadCount       string
}

func dmidecodeCommand() string {
	return "dmidecode"
}

func runSystemProfiler(dataType string) ([]byte, error) {
	cmd := exec.Command("system_profiler", "-xml", dataType)
	buf := bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(&buf)
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "could not execute system_profiler with dataType %s", dataType)
	}

	return buf.Bytes(), nil
}

func listPCIDevices() ([]*pciDeviceInfo, error) {
	xml, err := runSystemProfiler("SPPCIDataType")
	if err != nil {
		log.WithError(err).Info("[HWINFO] could not list PCI devices. Skipping...")
		return nil, nil
	}
	result, err := parseOutputToListOfPCIDevices(bytes.NewReader(xml))
	if err != nil {
		return nil, errors.Wrap(err, "could not parse PCI devices")
	}
	return result, nil
}

func listUSBDevices() ([]*usbDeviceInfo, error) {
	xml, err := runSystemProfiler("SPUSBDataType")
	if err != nil {
		log.WithError(err).Info("[HWINFO] could not list USB devices. Skipping...")
		return nil, nil
	}
	result, err := parseOutputToListOfUSBDevices(bytes.NewReader(xml))
	if err != nil {
		return nil, errors.Wrap(err, "could not parse USB devices")
	}
	return result, nil
}

func listDisplays() ([]*monitorInfo, error) {
	xml, err := runSystemProfiler("SPDisplaysDataType")
	if err != nil {
		log.WithError(err).Info("[HWINFO] could not list displays. Skipping...")
		return nil, nil
	}
	result, err := parseOutputToListOfDisplays(bytes.NewReader(xml))
	if err != nil {
		return nil, errors.Wrap(err, "could not parse displays list")
	}
	return result, nil
}

func listCPUs() (map[string]interface{}, error) {
	var parsedCPUs []cpuInfo
	sysctl, err := exec.LookPath("/usr/sbin/sysctl")
	if err != nil {
		return nil, err
	}

	out, err := common.RunCommandInBackground(sysctl, "machdep.cpu")
	if err != nil {
		return nil, err
	}

	c := cpuInfo{}
	for _, line := range strings.Split(string(out), "\n") {
		values := strings.Fields(line)
		if len(values) < 1 {
			continue
		}

		switch {
		case strings.HasPrefix(line, "machdep.cpu.vendor"):
			c.manufacturer = values[1]
		case strings.HasPrefix(line, "machdep.cpu.brand_string"):
			c.description = strings.Join(values[1:], " ")
		case strings.HasPrefix(line, "machdep.cpu.thread_count"):
			c.threadCount = values[1]
		case strings.HasPrefix(line, "machdep.cpu.cores_per_package"):
			c.coreCount = values[1]
		case strings.HasPrefix(line, "machdep.cpu.family"):
			if c.manufacturingInfo != "" {
				c.manufacturingInfo += " "
			}
			c.manufacturingInfo += "Family " + values[1]
		case strings.HasPrefix(line, "machdep.cpu.model"):
			if c.manufacturingInfo != "" {
				c.manufacturingInfo += " "
			}
			c.manufacturingInfo += "Model " + values[1]
		case strings.HasPrefix(line, "machdep.cpu.stepping"):
			if c.manufacturingInfo != "" {
				c.manufacturingInfo += " "
			}
			c.manufacturingInfo += "Stepping " + values[1]
		}
	}

	parsedCPUs = append(parsedCPUs, c)

	encodedCpus := make(map[string]interface{})

	for i := range parsedCPUs {
		encodedCpus[fmt.Sprintf("cpu.%d.manufacturer", i)] = parsedCPUs[i].manufacturer
		encodedCpus[fmt.Sprintf("cpu.%d.manufacturing_info", i)] = parsedCPUs[i].manufacturingInfo
		encodedCpus[fmt.Sprintf("cpu.%d.description", i)] = parsedCPUs[i].description
		encodedCpus[fmt.Sprintf("cpu.%d.core_count", i)] = parsedCPUs[i].coreCount
		encodedCpus[fmt.Sprintf("cpu.%d.thread_count", i)] = parsedCPUs[i].threadCount
	}

	return encodedCpus, nil
}
