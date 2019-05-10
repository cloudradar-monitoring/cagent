package smart

import (
	"github.com/pkg/errors"
)

// Option callback for connection option
type Option func(*SMART) error

func Executable(val string, tryDefaultOnFail bool) Option {
	return func(sm *SMART) error {
		var err error
		if val == "" && !tryDefaultOnFail {
			return errors.New("smartctl: invalid path to smartctl executable")
		}

		if val != "" {
			if err = sm.detectTools(val); err == nil {
				return nil
			} else if !tryDefaultOnFail {
				return err
			}
		}

		return sm.detectTools(defaultSmartctlPath)
	}
}
