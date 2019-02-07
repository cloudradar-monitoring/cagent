package docker

import (
	"errors"
	"runtime"
)

var ErrorNotImplementedForOS = errors.New("docker support not implemented for " + runtime.GOOS)
var ErrorDockerNotFound = errors.New("docker executable not found on the system")
