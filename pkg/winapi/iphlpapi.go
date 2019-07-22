// +build windows

package winapi

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modiphlpapi = windows.NewLazySystemDLL("iphlpapi.dll")

	procGetAdaptersAddresses = modiphlpapi.NewProc("GetAdaptersAddresses")
)

// https://docs.microsoft.com/ru-ru/windows/win32/api/iptypes/ns-iptypes-_ip_adapter_addresses_lh
type IPAdapterAddresses struct {
	Length                uint32
	IfIndex               uint32
	Next                  *IPAdapterAddresses
	AdapterName           *byte
	FirstUnicastAddress   *windows.IpAdapterUnicastAddress
	FirstAnycastAddress   *windows.IpAdapterAnycastAddress
	FirstMulticastAddress *windows.IpAdapterMulticastAddress
	FirstDNSServerAddress *windows.IpAdapterDnsServerAdapter
	DNSSuffix             *uint16
	Description           *uint16
	FriendlyName          *uint16
	PhysicalAddress       [syscall.MAX_ADAPTER_ADDRESS_LENGTH]byte
	PhysicalAddressLength uint32
	Flags                 uint32
	Mtu                   uint32
	IfType                uint32
	OperStatus            uint32
	Ipv6IfIndex           uint32
	ZoneIndices           [16]uint32
	FirstPrefix           *windows.IpAdapterPrefix
	TransmitLinkSpeed     uint64
	ReceiveLinkSpeed      uint64
	/* more fields might be present here. */
}

func (a *IPAdapterAddresses) GetInterfaceName() string {
	return syscall.UTF16ToString((*(*[10000]uint16)(unsafe.Pointer(a.FriendlyName)))[:])
}

// GetAdaptersAddresses returns a list of IP adapter and address
// structures. The structure contains an IP adapter and flattened
// multiple IP addresses including unicast, anycast and multicast
// addresses.
func GetAdaptersAddresses() ([]*IPAdapterAddresses, error) {
	var b []byte
	l := uint32(15000) // recommended initial size
	for {
		b = make([]byte, l)
		var err error
		r0, _, _ := syscall.Syscall6(
			procGetAdaptersAddresses.Addr(),
			5,
			uintptr(syscall.AF_UNSPEC),
			uintptr(windows.GAA_FLAG_INCLUDE_PREFIX),
			uintptr(0),
			uintptr(unsafe.Pointer((*IPAdapterAddresses)(unsafe.Pointer(&b[0])))),
			uintptr(unsafe.Pointer(&l)),
			0,
		)

		if r0 == 0 {
			if l == 0 {
				return nil, nil
			}
			break
		} else {
			err = syscall.Errno(r0)
		}

		if err.(syscall.Errno) != syscall.ERROR_BUFFER_OVERFLOW {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}

		if l <= uint32(len(b)) {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
	}
	var aas []*IPAdapterAddresses
	for aa := (*IPAdapterAddresses)(unsafe.Pointer(&b[0])); aa != nil; aa = aa.Next {
		aas = append(aas, aa)
	}
	return aas, nil
}
