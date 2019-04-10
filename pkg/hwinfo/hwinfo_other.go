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
	"strings"

	"github.com/jaypipes/ghw"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vcraescu/go-xrandr"
)

var lsusbLineRegexp = regexp.MustCompile(`[0-9|a-z|A-Z|.|/|-|:|\[|\]|_|+| ]+`)

func captureStderr(funcToExecute func()) (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	// replace default stderr Writer with our own to capture output:
	defaultStderrWriter := os.Stderr
	os.Stderr = w

	// in any case, the default writer will be set again:
	defer func() {
		os.Stderr = defaultStderrWriter
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

func listPCIDevices() ([]*pciDeviceInfo, error) {
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
		return nil, errors.Wrap(err, "could not capture stderr when retrieving PCI information using ghw")
	}
	if ghwErr != nil {
		return nil, errors.Wrap(ghwErr, "there were error while retrieving PCI information using ghw")
	}
	if len(stderrOutput) > 0 {
		// ghw only reports to stderr in case if system files are missing or there was failure while reading it
		log.Warnf("[HWINFO] got error output while retrieving PCI information using ghw: %s\nProbably the system is not supported or system files unreadable.", stderrOutput)
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
	return result, nil
}

func listUSBDevices() ([]*usbDeviceInfo, error) {
	results := make([]*usbDeviceInfo, 0)
	reg := regexp.MustCompile(`[^:]+`)
	var lines []string

	cmd := exec.Command("lsusb")
	buf := bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(&buf)
	if err := cmd.Run(); err != nil {
		log.Info("[HWINFO] lsusb command is not available. Skipping USB listing...")
		return nil, nil
	}

	outBytes, err := ioutil.ReadAll(bufio.NewReader(&buf))
	if err != nil {
		return nil, errors.Wrap(err, "could not read lsusb output")
	}

	lines = strings.Split(string(outBytes), "\n")

	// tokenize and parse command output line by line:
	const minExpectedTokensCount = 6
	for _, line := range lines {
		tokens := strings.Split(line, " ")
		sanitizedTokens := make([]string, 0)
		for _, t := range tokens {
			if t != "" && t != "\t" {
				sanitizedTokens = append(sanitizedTokens, t)
			}
		}
		sanitizedTokensCount := len(sanitizedTokens)
		if sanitizedTokensCount < minExpectedTokensCount {
			if sanitizedTokensCount > 0 {
				log.Warnf("[HWINFO] unexpected lsusb command output: got %d tokens in line: %s", sanitizedTokensCount, line)
			}
			continue
		}

		var description string
		if sanitizedTokensCount > minExpectedTokensCount {
			restTokens := sanitizedTokens[minExpectedTokensCount:]
			description = strings.Join(restTokens, " ")
		}

		address := fmt.Sprintf("bus %s device %s", sanitizedTokens[1], reg.FindString(sanitizedTokens[3]))
		devID := lsusbLineRegexp.FindString(sanitizedTokens[5])
		results = append(results, &usbDeviceInfo{
			Address:     address,
			VendorName:  "",
			DeviceID:    devID,
			Description: description,
		})
	}
	return results, nil
}

func listDisplays() ([]*monitorInfo, error) {
	results := make([]*monitorInfo, 0)
	screens, err := xrandr.GetScreens()
	if err != nil {
		log.WithError(err).Info("[HWINFO] xrandr not installed or returned not expected result. Skipping display listing...")
		return nil, nil
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

	return results, nil
}
