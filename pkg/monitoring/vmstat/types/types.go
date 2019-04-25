package types

import (
	"errors"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

var (
	ErrAlreadyExists = errors.New("vmstat: provider already registered")
	ErrNotAvailable  = errors.New("vmstat: provider not available")
	ErrCheck         = errors.New("vmstat: check provider availability")
	ErrNotRegistered = errors.New("vmstat: provider not registered")
)

type Provider interface {
	Run() error
	Shutdown() error

	Name() string
	IsAvailable() error
	GetMeasurements() (common.MeasurementsMap, error)
}
