// +build windows

package osinfo

import "github.com/shirou/gopsutil/host"

func osName() (string, error) {
	info, err := host.Info()

	if err != nil {
		return "", err
	}
	return info.Platform + " " + info.PlatformVersion + " " + info.PlatformFamily, nil
}
