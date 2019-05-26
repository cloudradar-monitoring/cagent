// +build linux

package osinfo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

var releaseFiles = map[string][]string{
	"Novell SUSE":   {"/etc/SUSE-release"},
	"CentOS":        {"/etc/centos-release"},
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

var osReleaseFile = "/etc/os-release"

var ubuntuDescriptionRegexp = regexp.MustCompile(`(?m)^DISTRIB_DESCRIPTION=\"(.+)\"$`)
var ubuntuCodenameRegexp = regexp.MustCompile(`(?m)^DISTRIB_CODENAME=(\w+)$`)

var osReleaseID = regexp.MustCompile(`(?m)^ID=\"?(.*[^"])\"?$`)
var osReleaseName = regexp.MustCompile(`(?m)^NAME=\"?(.*[^"])\"?$`)
var osReleaseVersion = regexp.MustCompile(`(?m)^VERSION=\"(.+)\"$`)

func uname() (*unix.Utsname, error) {
	uts := &unix.Utsname{}

	if err := unix.Uname(uts); err != nil {
		return nil, err
	}
	return uts, nil
}

func osName() (string, error) {
	uts, err := uname()
	if err != nil {
		log.WithError(err)
	}

	// try /etc/os-release first
	if _, err := os.Stat(osReleaseFile); err == nil {
		var result string

		var data []byte
		if data, err = ioutil.ReadFile(osReleaseFile); err != nil {
			return result, errors.Wrapf(err, "osinfo: couldn't read release info from \"%s\"", osReleaseFile)
		}

		if osReleaseID.Match(data) {
			res := osReleaseID.FindAllStringSubmatch(string(data), -1)
			result = prettyId(res[0][1]) + ": "
		} else {
			log.Error("not matched!!!!!!!")
		}

		if osReleaseName.Match(data) {
			res := osReleaseName.FindAllStringSubmatch(string(data), -1)
			result += res[0][1]
		}

		if osReleaseVersion.Match(data) {
			res := osReleaseVersion.FindAllStringSubmatch(string(data), -1)
			result += " " + res[0][1]
		}

		if uts != nil {
			result += fmt.Sprintf(" %s", string(uts.Release[:bytes.IndexByte(uts.Release[:], 0)]))
		}

		return result, nil
	}

	for dist, files := range releaseFiles {
		if _, err := os.Stat(files[0]); err == nil {
			result := dist
			var data []byte
			if data, err = ioutil.ReadFile(files[0]); err != nil {
				return result, errors.Wrapf(err, "osinfo: couldn't read release info from \"%s\"", files[0])
			}

			switch dist {
			case "Ubuntu":
				fallthrough
			case "Debian":
				if ubuntuDescriptionRegexp.Match(data) {
					res := ubuntuDescriptionRegexp.FindAllStringSubmatch(string(data), -1)
					result += ": " + res[0][1]
				}

				if ubuntuCodenameRegexp.Match(data) {
					res := ubuntuCodenameRegexp.FindAllStringSubmatch(string(data), -1)
					result += " (" + res[0][1] + ")"
				}

				if uts != nil {
					result += fmt.Sprintf(" %s", string(uts.Release[:bytes.IndexByte(uts.Release[:], 0)]))
				}
			default:
				result += ": " + strings.TrimSuffix(string(data), "\n")
				result = strings.TrimSuffix(result, " ")
				result = strings.ReplaceAll(result, "\n", " ")
				if len(files) == 2 {
					if _, err = os.Stat(files[1]); err == nil {
						if data, err = ioutil.ReadFile(files[1]); err != nil {
							return result, errors.Wrapf(err, "osinfo: couldn't read version info from \"%s\"", files[0])
						}
						result += ": " + strings.TrimSuffix(string(data), "\n")
					}
				}

			}

			return result, nil
		}
	}

	return "", ErrUnknownOSType
}

func prettyId(id string) string {
	switch id {
	case "debian":
		return "Debian"
	case "ubuntu":
		return "Ubuntu"
	case "centos":
		return "CentOS"
	case "rhel":
		return "Red Hat"
	case "gentoo":
		return "Gentoo"
	case "slackware":
		return "Slackware"
	case "fedora":
		return "Fedora"
	}

	return id
}
