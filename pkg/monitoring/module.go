package monitoring

import (
	"time"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

type Alert string
type Warning string

type Module interface {
	GetDescription() string
	IsEnabled() bool
	Run() ([]*ModuleReport, error)
}

// ModuleReport provides the results of Module run
// Do not use the struct directly, use NewReport() to initialize it
type ModuleReport struct {
	Name            string                 `json:"name"`
	Timestamp       common.Timestamp       `json:"timestamp"`
	ExecutedCommand string                 `json:"command executed"`
	Alerts          []Alert                `json:"alerts"`
	Warnings        []Warning              `json:"warnings"`
	Message         string                 `json:"message,omitempty"`
	Measurements    map[string]interface{} `json:"measurements,omitempty"`
}

func NewReport(name string, t time.Time, cmd string) ModuleReport {
	return ModuleReport{
		Name:            name,
		Timestamp:       common.Timestamp(t),
		ExecutedCommand: cmd,
		Alerts:          make([]Alert, 0),
		Warnings:        make([]Warning, 0),
		Message:         "",
		Measurements:    nil,
	}
}
