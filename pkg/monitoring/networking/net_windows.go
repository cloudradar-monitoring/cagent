// +build windows

package networking

import (
	"github.com/pkg/errors"

	"github.com/cloudradar-monitoring/cagent/pkg/winapi"
)

type windowsLinkSpeedProvider struct {
	// interface name -> bytes per second
	cache         map[string]float64
	isInitialized bool
}

func newLinkSpeedProvider() linkSpeedProvider {
	return &windowsLinkSpeedProvider{
		cache: make(map[string]float64),
	}
}

func (p *windowsLinkSpeedProvider) init() error {
	interfacesInfo, err := winapi.GetAdaptersAddresses()
	if err != nil {
		return errors.Wrap(err, "GetAdaptersAddresses failed")
	}

	for _, interfaceInfo := range interfacesInfo {
		p.cache[interfaceInfo.GetInterfaceName()] = float64(interfaceInfo.ReceiveLinkSpeed) / 8
	}
	p.isInitialized = true
	return nil
}

func (p *windowsLinkSpeedProvider) GetMaxAvailableLinkSpeed(ifName string) (float64, error) {
	if !p.isInitialized {
		err := p.init()
		if err != nil {
			return 0, err
		}
	}

	result, exists := p.cache[ifName]
	if !exists {
		return 0, errors.New("no speed information found")
	}
	return result, nil
}
