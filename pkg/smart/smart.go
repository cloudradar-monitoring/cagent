package smart

import (
	"fmt"

	"github.com/pkg/errors"
)

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

	if _, err = smartctlIsSupportedVersion(buildStr); err != nil {
		return errors.Wrap(err, "while checking smartctl version")
	}

	sm.smartctl = path
	sm.smartctlDetected = true

	fmt.Printf("smartctl path: %s\n", path)
	return nil
}
