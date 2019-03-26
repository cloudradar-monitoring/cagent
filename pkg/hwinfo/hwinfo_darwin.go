// +build darwin

package hwinfo

import (
	"bufio"
	"bytes"
	"os/exec"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

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
