package proxydetect

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

var ErrNotFound = fmt.Errorf("not found")
var log = logrus.WithField("package", "proxydetect")
var UserAgent = "proxydetect" // used for autoconfig requests

func GetProxyForURL(u *url.URL) (*url.URL, error) {
	// priority for the proxy set in the environment vars
	req := http.Request{URL: u}
	proxy, err := http.ProxyFromEnvironment(&req)
	if err != nil {
		return nil, err
	} else if proxy != nil {
		return proxy, nil
	}

	// then fallback to the system-specific proxy detection(only windows is implemented right now)
	return getProxyForURL(u)
}

func GetProxyForRequest(r *http.Request) (*url.URL, error) {
	// to be compatible with http.Transport's Proxy field
	//  we need to return nil, nil in case Proxy is not set

	if r == nil {
		return nil, nil
	}

	if r.URL == nil {
		return nil, nil
	}

	proxy, err := GetProxyForURL(r.URL)
	if err == ErrNotFound {
		return nil, nil
	}

	return proxy, err
}
