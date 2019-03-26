package hwinfo

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func helperLoadSystemProfilerXML(t *testing.T, xmlFileName string) []byte {
	path := filepath.Join("testdata", xmlFileName)
	result, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestParseOutputToListOfPCIDevices(t *testing.T) {
	xml := helperLoadSystemProfilerXML(t, "pci.xml")

	pciDevicesList, err := parseOutputToListOfPCIDevices(bytes.NewReader(xml))
	assert.NoError(t, err)

	expectedPCIDevicesList := []*pciDeviceInfo{
		{
			Address:     "Thunderbolt@191,0,0",
			DeviceType:  "ieee1394openhci",
			VendorName:  "0x11c1",
			ProductName: "pci11c1,5901",
			Description: "",
		},
		{
			Address:     "Thunderbolt@192,0,0",
			DeviceType:  "ethernet",
			VendorName:  "0x14e4",
			ProductName: "Apple 57761-B0",
			Description: "ethernet",
		},
		{
			Address:     "Thunderbolt@195,0,0",
			DeviceType:  "usbopenhost",
			VendorName:  "0x12d8",
			ProductName: "pci12d8,400e",
			Description: "",
		},
		{
			Address:     "Thunderbolt@195,0,1",
			DeviceType:  "usbopenhost",
			VendorName:  "0x12d8",
			ProductName: "pci12d8,400e",
			Description: "",
		},
		{
			Address:     "Thunderbolt@195,0,2",
			DeviceType:  "usbehci",
			VendorName:  "0x12d8",
			ProductName: "pci12d8,400f",
			Description: "",
		},
	}

	assert.EqualValues(t, expectedPCIDevicesList, pciDevicesList)
}

func TestParseOutputToListOfUSBDevices(t *testing.T) {
	xml := helperLoadSystemProfilerXML(t, "usb.xml")

	usbDevicesList, err := parseOutputToListOfUSBDevices(bytes.NewReader(xml))
	assert.NoError(t, err)

	expectedUSBDevicesList := []*usbDeviceInfo{
		{
			Address:     "0x1e2d",
			VendorName:  "",
			DeviceID:    "",
			Description: "USB20Bus AppleUSBEHCIPCI",
		},
		{
			Address:     "0x1a100000 / 1",
			VendorName:  "0x8087 (Intel Corporation)",
			DeviceID:    "0x0024",
			Description: "hub_device",
		},
		{
			Address:     "0x1e26",
			VendorName:  "",
			DeviceID:    "",
			Description: "USB20Bus AppleUSBEHCIPCI",
		},
		{
			Address:     "0x1d100000 / 1",
			VendorName:  "0x8087 (Intel Corporation)",
			DeviceID:    "0x0024",
			Description: "hub_device",
		},
		{
			Address:     "0x1d180000 / 2",
			VendorName:  "0x0424 (SMSC)",
			DeviceID:    "0x2512",
			Description: "hub_device",
		},
		{
			Address:     "0x1d182000 / 3",
			VendorName:  "Apple, Inc.",
			DeviceID:    "0x8242",
			Description: "IR Receiver",
		},
		{
			Address:     "0x1d181000 / 4",
			VendorName:  "Apple Inc. 0x0a5c (Broadcom Corp.)",
			DeviceID:    "0x4500",
			Description: "BRCM20702 Hub",
		},
		{
			Address:     "0x1d181300 / 7",
			VendorName:  "Apple Inc.",
			DeviceID:    "0x828a",
			Description: "Bluetooth USB Host Controller",
		},
		{
			Address:     "0x1e31",
			VendorName:  "",
			DeviceID:    "",
			Description: "USB30Bus AppleUSBXHCIPPT",
		},
	}

	assert.EqualValues(t, expectedUSBDevicesList, usbDevicesList)
}

func TestParseOutputToListOfDisplays(t *testing.T) {
	xml := helperLoadSystemProfilerXML(t, "displays.xml")

	displayList, err := parseOutputToListOfDisplays(bytes.NewReader(xml))
	assert.NoError(t, err)

	expectedDisplayList := []*monitorInfo{
		{
			ID:          "Thunderbolt Display",
			Description: "Display Type: LCD, Connection Type: displayport_dongletype_dp",
			VendorName:  "610",
			Size:        "",
			Resolution:  "2560 x 1440",
		},
	}

	assert.EqualValues(t, expectedDisplayList, displayList)
}
