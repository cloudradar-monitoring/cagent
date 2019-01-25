package docker

import (
	"errors"
	"runtime"
)

var ErrorNotImplementedForOS = errors.New("Docker support not implemented for " + runtime.GOOS)
