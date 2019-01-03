package hwinfo

func Inventory() (map[string]interface{}, error) {
	hw, err := fetchInventory()
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	res[""] = hw

	return res, nil
}
