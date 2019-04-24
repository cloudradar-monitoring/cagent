package cagent

import (
	"fmt"
	"time"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

type Result struct {
	Timestamp    int64                  `json:"timestamp"`
	Measurements common.MeasurementsMap `json:"measurements"`
	Message      interface{}            `json:"message"`
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
