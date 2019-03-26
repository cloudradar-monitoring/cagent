// Package for parsing the XML output of system_profiler command output (available on OS X)

package hwinfo

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"howett.net/plist"
)

type spPCIDataTypeEntry struct {
	Name       string `plist:"_name"`
	DeviceType string `plist:"sppci_device_type"`
	SlotName   string `plist:"sppci_slot_name"`
	VendorID   string `plist:"sppci_vendor-id"`
	NameExtra  string `plist:"sppci_name"`
}

const spPCIPrefix = "sppci_"

type spPCIDataType struct {
	Items []spPCIDataTypeEntry `plist:"_items"`
}

type spUSBDataTypeEntry struct {
	Items []spUSBDataTypeEntry `plist:"_items"`

	Name           string `plist:"_name"`
	HostController string `plist:"host_controller"`
	PCIDevice      string `plist:"pci_device"`
	LocationID     string `plist:"location_id"`
	ProductID      string `plist:"product_id"`
	VendorID       string `plist:"vendor_id"`
	Manufacturer   string `plist:"manufacturer"`
}

type spUSBDataType struct {
	Items []spUSBDataTypeEntry `plist:"_items"`
}

type spDisplayDataTypeEntry struct {
	Name            string `plist:"_name"`
	ResolutionExtra string `plist:"_spdisplays_resolution"`
	Resolution      string `plist:"spdisplays_resolution"`
	ConnectionType  string `plist:"spdisplays_connection_type"`
	DisplayType     string `plist:"spdisplays_display_type"`
	VendorID        string `plist:"_spdisplays_display-vendor-id"`
}

const spDisplaysPrefix = "spdisplays_"

type spGraphicsCardDataTypeEntry struct {
	Displays []spDisplayDataTypeEntry `plist:"spdisplays_ndrvs"`
}

type spDisplaysDataType struct {
	GraphicCards []spGraphicsCardDataTypeEntry `plist:"_items"`
}

func parseOutputToListOfPCIDevices(r io.ReadSeeker) ([]*pciDeviceInfo, error) {
	decoder := plist.NewDecoder(r)
	var data []spPCIDataType
	err := decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("unexpected XML input: no entries in plist of PCI devices")
	}

	result := make([]*pciDeviceInfo, 0)
	for _, device := range data[0].Items {
		pciInfo := &pciDeviceInfo{
			Address:     device.SlotName,
			DeviceType:  strings.TrimPrefix(device.DeviceType, spPCIPrefix),
			VendorName:  device.VendorID,
			ProductName: device.Name,
			Description: device.NameExtra,
		}
		result = append(result, pciInfo)
	}
	return result, nil
}

func parseOutputToListOfUSBDevices(r io.ReadSeeker) ([]*usbDeviceInfo, error) {
	decoder := plist.NewDecoder(r)
	var data []spUSBDataType
	err := decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("unexpected XML input: no entries in plist of USB devices")
	}

	return getUSBInfoFromHierarchy(data[0].Items), nil
}

func getUSBInfoFromHierarchy(items []spUSBDataTypeEntry) []*usbDeviceInfo {
	list := make([]*usbDeviceInfo, 0)
	for _, item := range items {
		vendorID := item.VendorID
		if vendorID == "apple_vendor_id" {
			vendorID = ""
		}
		vendorName := item.Manufacturer + " " + vendorID
		address := item.LocationID + " " + item.PCIDevice
		description := item.Name + " " + item.HostController
		usbInfo := &usbDeviceInfo{
			VendorName:  strings.TrimSpace(vendorName),
			Address:     strings.TrimSpace(address),
			Description: strings.TrimSpace(description),
			DeviceID:    item.ProductID,
		}
		list = append(list, usbInfo)
		list = append(list, getUSBInfoFromHierarchy(item.Items)...)
	}
	return list
}

func parseOutputToListOfDisplays(r io.ReadSeeker) ([]*monitorInfo, error) {
	decoder := plist.NewDecoder(r)
	var data []spDisplaysDataType
	err := decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("unexpected XML input: no entries in plist of monitors")
	}

	result := make([]*monitorInfo, 0)
	for _, graphicsCard := range data[0].GraphicCards {
		for _, display := range graphicsCard.Displays {
			resolution := display.Resolution
			if resolution == "" {
				resolution = display.ResolutionExtra
			}

			displayType := strings.TrimPrefix(display.DisplayType, spDisplaysPrefix)
			connectionType := strings.TrimPrefix(display.ConnectionType, spDisplaysPrefix)
			description := fmt.Sprintf("Display Type: %s, Connection Type: %s", displayType, connectionType)
			monitorInfo := &monitorInfo{
				ID:          display.Name,
				Description: description,
				VendorName:  display.VendorID,
				Size:        "",
				Resolution:  resolution,
			}
			result = append(result, monitorInfo)
		}
	}
	return result, nil
}
