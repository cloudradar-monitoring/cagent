package services

import (
	"errors"
	"runtime"
)

// ErrorNotImplementedForOS returned in case we don't yet implement service manager parsing or this OS. Should be checked and ignored
var ErrorNotImplementedForOS = errors.New("Services list not implemented for " + runtime.GOOS)
