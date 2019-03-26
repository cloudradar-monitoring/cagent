package hwinfo

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type pciDeviceInfo struct {
	Address     string `json:"address"`
	DeviceType  string `json:"device_type,omitempty"`
	VendorName  string `json:"vendor_name,omitempty"`
	ProductName string `json:"product_name"`
	Description string `json:"description,omitempty"`
}

type usbDeviceInfo struct {
	Address     string `json:"address,omitempty"`
	VendorName  string `json:"vendor_name,omitempty"`
	DeviceID    string `json:"id"`
	Description string `json:"description,omitempty"`
}

type monitorInfo struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	VendorName  string `json:"vendor_name,omitempty"`
	Size        string `json:"size,omitempty"`
	Resolution  string `json:"resolution,omitempty"`
}

func Inventory() (map[string]interface{}, error) {
	hw, err := fetchInventory()
	if err != nil {
		err = errors.Wrap(err, "[HWINFO]")
		log.Error(err)
		return hw, err
	}

	return hw, nil
}
