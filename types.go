package cagent

import (
	"fmt"
	"time"
)

type MeasurementsMap map[string]interface{}

func (mm MeasurementsMap) AddWithPrefix(prefix string, m MeasurementsMap) MeasurementsMap {
	for k, v := range m {
		mm[prefix+k] = v
	}
	return mm
}

func (mm MeasurementsMap) AddInnerWithPrefix(prefix string, m MeasurementsMap) MeasurementsMap {
	mm[prefix] = m

	return mm
}

type Result struct {
	Timestamp    int64           `json:"timestamp"`
	Measurements MeasurementsMap `json:"measurements"`
	Message      interface{}     `json:"message"`
}

func floatToIntPercentRoundUP(f float64) int {
	return int(f*100 + 0.5)
}

func floatToIntRoundUP(f float64) int {
	return int(f + 0.5)
}

type TimeoutError struct {
	Origin  string
	Timeout time.Duration
}

func (err TimeoutError) Error() string {
	return fmt.Sprintf("%s timeout after %.1fs", err.Origin, err.Timeout.Seconds())
}
