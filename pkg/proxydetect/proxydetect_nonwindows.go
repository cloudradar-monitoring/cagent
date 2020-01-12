// +build !windows

package proxydetect

import "net/url"

// todo: darwin proxy detection

func getProxyForURL(u *url.URL) (*url.URL, error) {
	// workaround non-used warning on windows
	_ = log
	return nil, ErrNotFound
}
