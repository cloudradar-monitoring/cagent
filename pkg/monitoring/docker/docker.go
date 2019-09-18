package docker

import (
	"errors"
	"runtime"

	"github.com/sirupsen/logrus"
)

var ErrorNotImplementedForOS = errors.New("docker support not implemented for " + runtime.GOOS)
var ErrorDockerNotAvailable = errors.New("docker executable not found on the system or the service is stopped")

var log = logrus.WithField("package", "docker")
