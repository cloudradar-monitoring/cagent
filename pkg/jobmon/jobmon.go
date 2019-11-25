package jobmon

import (
	"strings"
	"time"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

type Severity string

const (
	SeverityAlert   = "alert"
	SeverityWarning = "warning"

	maxStdStreamBufferSize = 100 // 4 * 1024
)

var ValidSeverities = []Severity{SeverityAlert, SeverityWarning}

func IsValidJobMonitoringSeverity(s Severity) bool {
	for _, validSeverity := range ValidSeverities {
		if validSeverity == s {
			return true
		}
	}
	return false
}

type JobRunConfig struct {
	JobID            string
	Severity         Severity
	NextRunInterval  *time.Duration
	MaxExecutionTime *time.Duration
	RecordStdErr     bool
	RecordStdOut     bool
	Command          []string
}

type JobRun struct {
	ID        string            `json:"id"`
	Command   string            `json:"command"`
	StartedAt common.Timestamp  `json:"job_started"`
	EndedAt   *common.Timestamp `json:"job_ended"`
	Duration  *uint64           `json:"job_duration_s"`
	User      string            `json:"job_user"`
	ExitCode  *int              `json:"exit_code"`
	Severity  Severity          `json:"severity"`
	NextRunIn *int              `json:"next_run_in"`
	StdOut    *string           `json:"stdout"`
	StdErr    *string           `json:"stderr"`
	Errors    []string          `json:"errors,omitempty"`
}

func newJobRun(cfg *JobRunConfig) *JobRun {
	var nextRunInSeconds *int
	if cfg.NextRunInterval != nil {
		s := int(cfg.NextRunInterval.Seconds())
		nextRunInSeconds = &s
	}
	return &JobRun{
		ID:        cfg.JobID,
		Command:   strings.Join(cfg.Command, " "),
		User:      getUsername(),
		Severity:  cfg.Severity,
		NextRunIn: nextRunInSeconds,
		StartedAt: common.Timestamp(time.Now()),
		Errors:    make([]string, 0),
	}
}

func (r *JobRun) AddError(msg string) {
	r.Errors = append(r.Errors, msg)
}
