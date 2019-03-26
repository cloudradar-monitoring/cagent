package windevice

// DeviceSelector is an interface capable of selecting devices in a device list.
type DeviceSelector interface {
	Select(Device) (bool, error)
}
