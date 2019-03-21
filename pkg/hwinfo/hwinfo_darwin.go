// +build darwin

package hwinfo

import (
	"bufio"
	"bytes"
	"os/exec"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/pkg/errors"
)

func runSystemProfiler(dataType string) ([]byte, error) {
	cmd := exec.Command("system_profiler", "-xml", dataType)
	buf := bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(&buf)
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, "could not execute system_profiler")
	}

	return buf.Bytes(), nil
}

func listPCIDevices(errs *common.ErrorCollector) []*pciDeviceInfo {
	xml, err := runSystemProfiler("SPPCIDataType")
	if err != nil {
		errs.New(err)
		return nil
	}
	return parseOutputToListOfPCIDevices(bytes.NewReader(xml), errs)
}

func listUSBDevices(errs *common.ErrorCollector) []*usbDeviceInfo {
	xml, err := runSystemProfiler("SPUSBDataType")
	if err != nil {
		errs.New(err)
		return nil
	}
	return parseOutputToListOfUSBDevices(bytes.NewReader(xml), errs)
}

func listDisplays(errs *common.ErrorCollector) []*monitorInfo {
	xml, err := runSystemProfiler("SPDisplaysDataType")
	if err != nil {
		errs.New(err)
		return nil
	}
	return parseOutputToListOfDisplays(bytes.NewReader(xml), errs)
}
