package smart

import (
	"fmt"

	"github.com/pkg/errors"
)

const atLeastMajorVersion = 7

type SMART struct {
	smartctl         string
	smartctlDetected bool
}

func New(opts ...Option) (*SMART, error) {
	sm := &SMART{
		smartctlDetected: false,
	}

	var err error
	for _, opt := range opts {
		if err = opt(sm); err != nil {
			return nil, err
		}
	}

	// if smartctl executable is not specified in the options list try to detect with default
	if !sm.smartctlDetected {
		if err = sm.detectTools(defaultSmartctlPath); err != nil {
			return nil, err
		}
	}

	return sm, nil
}

func (sm *SMART) detectTools(smartctl string) error {
	buildStr, path, err := checkTools(smartctl)
	if err != nil {
		return errors.Wrap(err, "while detecting smartctl")
	}

	var major int
	var minor int

	if major, minor, err = smartctlParseVersion(buildStr); err != nil {
		return errors.Wrap(err, "while checking smartctl version")
	}

	if major < atLeastMajorVersion {
		return fmt.Errorf("smart: unsupported smartctl version. expected minimum [7.0], actual [%d.%d]", major, minor)
	}

	sm.smartctl = path
	sm.smartctlDetected = true

	return nil
}
