package osinfo

import (
	"github.com/pkg/errors"
)

var (
	ErrUnknownOSType = errors.New("osinfo: unknown OS type")
)

func GetOsName() (string, error) {
	return osName()
}
