package smart

import (
	"github.com/pkg/errors"
)

var (
	ErrNoDisksFound             = errors.New("smart: no physical disks found in the system")
	ErrSmartctlNotFound         = errors.New("smart: couldn't detect path to smartctl executable")
	ErrSmartctlInfoNotAvailable = errors.New("smart: cannot get version string from smartctl executable")
	ErrUnderlyingToolNotFound   = errors.New("smart: underlying tool not found")
	ErrParseSmartctlVersion     = errors.New("smart: unable parse smartctl version")
)
