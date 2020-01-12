// +build !windows

package proxydetect

import "net/url"

// todo: darwin proxy detection

func getProxyForUrl(u *url.URL) (*url.URL, error) {
	return nil, ErrNotFound
}
