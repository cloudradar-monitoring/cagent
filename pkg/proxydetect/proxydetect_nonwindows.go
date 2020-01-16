// +build !windows

package proxydetect

import "net/url"

func getProxyForURL(u *url.URL) (*url.URL, error) {
	// workaround non-used warning on unix
	_ = log
	return nil, ErrNotFound
}
