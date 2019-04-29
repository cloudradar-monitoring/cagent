package smart

import (
	"bytes"
	"regexp"
)

var diskScanRegexp = regexp.MustCompile(`^(\/dev\/(\w|\d)+)(\s?(\w|\W)+)?.?$`)

func parseDisks(buf *bytes.Buffer) ([]string, error) {
	var disks []string

	for {
		s, err := buf.ReadString('\n')
		if s != "" {
			disks = append(disks, s)
		}

		if err != nil {
			break
		}
	}

	for i, d := range disks {
		if diskScanRegexp.Match([]byte(d)) {
			res := diskScanRegexp.FindAllStringSubmatch(d, -1)
			disks[i] = res[0][1]
		}
	}

	if len(disks) == 0 {
		return nil, ErrNoDisksFound
	}

	return disks, nil
}
