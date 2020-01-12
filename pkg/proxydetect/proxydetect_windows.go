// +build windows

package proxydetect

import (
	"net"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cloudradar-monitoring/cagent/pkg/winapi"
	"golang.org/x/sys/windows"
)

const (
	resolveTimeout = 5000
	connectTimeout = 5000
	sendTimeout    = 10000
	receiveTimeout = 10000
)

func getProxyForUrl(u *url.URL) (proxy *url.URL, err error) {
	ieProxy, err := winapi.HttpGetIEProxyConfigForCurrentUser()
	if err != nil {
		log.Errorf("HttpGetIEProxyConfigForCurrentUser error: %s", err.Error())
	} else {
		defer func() {
			if err := ieProxy.Free(); err != nil {
				log.Errorf("failed to free resources after winapi call: %s", err.Error())
			}
		}()

		if ieProxy.FAutoDetect {
			proxyInfo, err := getProxyInfoForUrl(u,
				&winapi.HttpAutoProxyOptions{
					DwFlags:                winapi.WINHTTP_AUTOPROXY_AUTO_DETECT,
					DwAutoDetectFlags:      winapi.WINHTTP_AUTO_DETECT_TYPE_DHCP | winapi.WINHTTP_AUTO_DETECT_TYPE_DNS_A,
					FAutoLogonIfChallenged: true,
				})
			if err == nil {
				defer func() {
					if err := proxyInfo.Free(); err != nil {
						log.Errorf("WINHTTP_AUTOPROXY_AUTO_DETECT failed to free resources: %s", err.Error())
					}
				}()

				proxy, err = matchProxyForURLAndBypassList(
					u,
					strings.Split(proxyInfo.LpszProxy.String(), ";"),
					strings.Split(proxyInfo.LpszProxyBypass.String(), ";"),
				)
				if err == nil {
					return proxy, nil
				}
			}
			if err != ErrNotFound {
				log.Errorf("WINHTTP_AUTOPROXY_AUTO_DETECT got error: %s", err.Error())
			}
		}

		if ieProxy.LpszAutoConfigUrl.String() != "" {
			proxyInfo, err := getProxyInfoForUrl(u,
				&winapi.HttpAutoProxyOptions{
					DwFlags:                winapi.WINHTTP_AUTOPROXY_CONFIG_URL,
					LpszAutoConfigUrl:      ieProxy.LpszAutoConfigUrl,
					FAutoLogonIfChallenged: true,
				})
			if err == nil {
				defer func() {
					if err := proxyInfo.Free(); err != nil {
						log.Errorf("WINHTTP_AUTOPROXY_CONFIG_URL failed to free resources: %s", err.Error())
					}
				}()
				proxy, err = matchProxyForURLAndBypassList(
					u,
					strings.Split(proxyInfo.LpszProxy.String(), ";"),
					strings.Split(proxyInfo.LpszProxyBypass.String(), ";"),
				)
				if err == nil {
					return proxy, nil
				}
			}

			if err != ErrNotFound {
				log.Errorf("WINHTTP_AUTOPROXY_CONFIG_URL got error: %s", err.Error())
			}
		}

		proxy, err = matchProxyForURLAndBypassList(
			u,
			strings.Split(ieProxy.LpszProxy.String(), ";"),
			strings.Split(ieProxy.LpszProxyBypass.String(), ";"),
		)
		if err == nil {
			return proxy, nil
		} else if err != ErrNotFound {
			log.Errorf("WINHTTP_CURRENT_USER_IE_PROXY_CONFIG named proxy got error: %s", err.Error())
		}
	}

	proxyInfo, err := winapi.HttpGetDefaultProxyConfiguration()
	if err == nil {
		defer func() {
			if err := proxyInfo.Free(); err != nil {
				log.Errorf("HttpGetDefaultProxyConfiguration failed to free resources: %s", err.Error())
			}
		}()

		return matchProxyForURLAndBypassList(
			u,
			strings.Split(proxyInfo.LpszProxy.String(), ";"),
			strings.Split(proxyInfo.LpszProxyBypass.String(), ";"),
		)
	} else if err != ErrNotFound {
		log.Errorf("WinHttpGetDefaultProxyConfiguration got error: %s", err.Error())
	}

	return nil, ErrNotFound
}

func getProxyInfoForUrl(targetUrl *url.URL, autoProxyOptions *winapi.HttpAutoProxyOptions) (*winapi.HttpProxyInfo, error) {
	userAgent, _ := windows.UTF16PtrFromString("cagent")

	h, err := winapi.HttpOpen(
		userAgent,
		winapi.WINHTTP_ACCESS_TYPE_NO_PROXY,
		nil,
		nil,
		0)
	if err != nil {
		return nil, err
	}
	defer winapi.HttpCloseHandle(h)

	err = winapi.SetTimeouts(h, resolveTimeout, connectTimeout, sendTimeout, receiveTimeout)
	if err != nil {
		return nil, err
	}

	proxyInfo, err := winapi.HttpGetProxyForUrl(h, targetUrl.String(), autoProxyOptions)
	if err != nil {
		return nil, err
	}

	return proxyInfo, nil
}

func matchProxyForURLAndBypassList(targetUrl *url.URL, proxies []string, bypassList []string) (*url.URL, error) {
	var proxyUrlStr string
	// lpszProxy can contains multiple proxies for different protocols
	// "10.0.0.1" or
	// "http=10.0.0.1;socks=10.0.0.2"

	// lets try to find the one that match the requested protocol, otherwise choose the one without specified protocol
	for _, s := range proxies {
		parts := strings.SplitN(s, "=", 2)
		//
		if len(parts) < 2 {
			// assign a match, but keep looking in case we have a protocol specific match
			proxyUrlStr = s
		} else if strings.EqualFold(strings.TrimSpace(parts[0]), targetUrl.Scheme) {
			// note that unlike UNIX, HTTP proxy will not be chosen for HTTPS requests
			proxyUrlStr = targetUrl.Scheme + "://" + parts[1]
			break
		} else if strings.EqualFold(strings.TrimSpace(parts[0]), "socks") {
			// SOCKS is universal protocol and can handle HTTP, HTTPS, FTP and others underneath
			// golang only supports SOCKS version 5, so assume that socks version is 5
			proxyUrlStr = "socks5://" + parts[1]
		}
	}

	if proxyUrlStr == "" {
		return nil, ErrNotFound
	}

	if !strings.HasPrefix(proxyUrlStr, "//") && !strings.Contains(proxyUrlStr, "://") {
		proxyUrlStr = "//" + proxyUrlStr
	}

	proxyUrl, err := url.Parse(proxyUrlStr)
	if err != nil {
		return nil, err
	}

	if isUrlBypassesProxy(targetUrl, bypassList) {
		return nil, ErrNotFound
	}

	return proxyUrl, nil
}

func isUrlBypassesProxy(targetUrl *url.URL, bypassList []string) bool {
	targetHost := targetUrl.Hostname()
	for _, bypass := range bypassList {
		bypass = strings.TrimSpace(bypass)
		if bypass == "" {
			continue
		} else if bypass == "<local>" {
			// local used to bypass for all loopback addresses
			addr, err := net.LookupIP(targetUrl.Hostname())
			if err == nil && len(addr) > 0 && addr[0].IsLoopback() {
				return true
			}
			continue
		}
		// use bypass as an exact pattern
		if shouldBypass, err := filepath.Match(bypass, targetHost); err != nil {
			return false
		} else if shouldBypass {
			return true
		}

		// add "*" to match subdomains (domain.com -> *.domain.com)
		if !strings.HasPrefix(bypass, "*") {
			if !strings.HasPrefix(bypass, ".") {
				bypass = "." + bypass
			}
			bypass = "*" + bypass
		}

		if shouldBypass, err := filepath.Match(bypass, targetHost); err != nil {
			return false
		} else if shouldBypass {
			return true
		}
	}
	return false
}
