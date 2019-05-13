// +build linux

package osinfo

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

var releaseFiles = map[string][]string{
	"Novell SUSE":   {"/etc/SUSE-release"},
	"Red Hat":       {"/etc/redhat-release", "/etc/redhat_version"},
	"Fedora":        {"/etc/fedora-release"},
	"Slackware":     {"/etc/slackware-release", "/etc/slackware-version"},
	"Debian":        {"/etc/debian_release", "/etc/debian_version"},
	"Mandrake":      {"/etc/mandrake-release"},
	"Yellow dog":    {"/etc/yellowdog-release"},
	"Sun JDS":       {"/etc/sun-release"},
	"Solaris/Sparc": {"/etc/release"},
	"Gentoo":        {"/etc/gentoo-release"},
	"UnitedLinux":   {"/etc/UnitedLinux-release"},
	"Ubuntu":        {"/etc/lsb-release"},
}

func osName() (string, error) {
	for dist, files := range releaseFiles {
		if _, err := os.Stat(files[0]); err == nil {
			result := dist
			var data []byte
			if data, err = ioutil.ReadFile(files[0]); err != nil {
				return result, errors.Wrapf(err, "osinfo: couldn't read release info from \"%s\"", files[0])
			}

			result += ": " + strings.TrimSuffix(string(data), "\n")
			result = strings.TrimSuffix(result, " ")
			if len(files) == 2 {
				if _, err = os.Stat(files[1]); err == nil {
					if data, err = ioutil.ReadFile(files[1]); err != nil {
						return result, errors.Wrapf(err, "osinfo: couldn't read version info from \"%s\"", files[0])
					}
					result += ": " + strings.TrimSuffix(string(data), "\n")
				}
			}

			return result, nil
		}
	}

	return "", ErrUnknownOSType
}
