// +build !darwin,!windows

package hwinfo

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/jaypipes/ghw"
	"github.com/vcraescu/go-xrandr"
)

var lsusbLineRegexp = regexp.MustCompile(`[0-9|a-z|A-Z|.|/|-|:|\[|\]|_|+| ]+`)

func captureStderr(funcToExecute func()) (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	stderr := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = stderr
	}()

	funcToExecute()
	err = w.Close()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func listPCIDevices(errs *common.ErrorCollector) []*pciDeviceInfo {
	var ghwErr error
	var devices []*ghw.PCIDevice

	// unfortunately, jaypipes/ghw library sometime writes error message directly to os.Stderr instead of returning from function call.
	// that's why we will try to capture stderr output and handle it:
	stderrOutput, err := captureStderr(func() {
		pciInfo, ghwErr := ghw.PCI()
		if ghwErr == nil {
			devices = pciInfo.ListDevices()
		}
	})
	if err != nil {
		errs.Addf("could not capture stderr when retrieving PCI information using ghw: %s", err.Error())
		return nil
	}
	if ghwErr != nil {
		errs.Addf("there were error while retrieving PCI information using ghw: %s", ghwErr.Error())
		return nil
	}
	if len(stderrOutput) > 0 {
		errs.Addf("there were error output while retrieving PCI information using ghw: %s", stderrOutput)
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
		address := fmt.Sprintf("bus %d device %d", busNum, devNum)
		devID := lsusbLineRegexp.FindString(sanitizedTokens[5])
		results = append(results, &usbDeviceInfo{
			address,
			"",
			devID,
			description,
		})
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
				ID:          m.ID,
				Size:        physicalSizeStr,
				Resolution:  resolutionStr,
				Description: "",
				VendorName:  "",
			})
		}
	}

	return results
}
