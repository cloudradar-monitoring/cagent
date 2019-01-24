package hwinfo

import (
	"errors"
)

var ErrNotPresent = errors.New("hwinfo: not present")

func Inventory() (map[string]interface{}, error) {
	hw, err := fetchInventory()
	if err != nil {
		return nil, err
	}

	return hw, nil
}
