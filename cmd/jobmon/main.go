package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent"
	"github.com/cloudradar-monitoring/cagent/pkg/jobmon"
)

const (
	minNextRunInterval     = 5 * time.Minute
	minValueForMaxExecTime = 1 * time.Second
	maxJobIDLength         = 100
)

var logger *logrus.Logger

func init() {
	// jobmon binary should not write any messages to log file
	logger = logrus.New()
	logger.SetOutput(os.Stderr)

	tfmt := logrus.TextFormatter{DisableTimestamp: true, DisableColors: true}
	logger.SetFormatter(&tfmt)
}

func main() {
	versionPtr := flag.Bool("version", false, "show the jobmon version")
	cfgPathPtr := flag.String("c", cagent.DefaultCfgPath, "config file path")

	jobIDPtr := flag.String("id", "", fmt.Sprintf("id of the job, required, maximum %d characters", maxJobIDLength))
	forceRunPtr := flag.Bool("f", false, "Force run of a job even if the job with the same ID is already running or its termination wasn't handled successfully.")
	severityPtr := flag.String("s", "", "alert|warning|none process failed job with this severity. Overwrites the default severity of cagent.conf. Severity 'none' suppresses all messages.")
	nextRunInPtr := flag.Duration("nr", 0, "<N>h|m indicates when the job should run for the next time. Allows triggering alters for not run jobs. The shortest interval is 5 minutes.")
	maxExecutionTimePtr := flag.Duration("me", 0, "<N>h|m|s or just <N> for number of seconds. Max execution time for job.")
	recordStdErrPtr := flag.Bool("re", false, "record errors from stderr, overwrites the default settings of cagent.conf, limited to the last 4 KB")
	recordStdOutPtr := flag.Bool("ro", false, "record errors from stdout, overwrites the default settings of cagent.conf, limited to the last 4 KB")

	flag.Parse()

	handleFlagVersion(*versionPtr)

	cfg, err := cagent.HandleAllConfigSetup(*cfgPathPtr)
	if err != nil {
		logger.Fatalf("Failed to handle Cagent configuration: %s", err.Error())
		return
	}

	jobConfig, err := initJobRunConfig(&cfg.JobMonitoring, jobIDPtr, severityPtr, nextRunInPtr, maxExecutionTimePtr, recordStdErrPtr, recordStdOutPtr)
	if err != nil {
		logger.Fatalf("Invalid parameter specified: %s", err.Error())
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	jobMonRunner := jobmon.NewRunner(cfg.JobMonitoring.SpoolDirPath, jobConfig, logger)
	err = jobMonRunner.RunJob(sigChan, *forceRunPtr)
	if err != nil {
		logger.Fatalf("Could not start a job: %s", err.Error())
		return
	}
}

func handleFlagVersion(versionFlag bool) {
	if versionFlag {
		fmt.Printf("jobmon - Job Monitoring tool.\nPart of cagent package v%s %s\n", cagent.Version, cagent.LicenseInfo)
		os.Exit(0)
	}
}

func initJobRunConfig(
	jobMonitoringConfig *cagent.JobMonitoringConfig,
	jobIDPtr, severityPtr *string,
	nextRunInPtr, maxExecutionTimePtr *time.Duration,
	recordStdErrPtr, recordStdOutPtr *bool,
) (*jobmon.JobRunConfig, error) {
	var severity = jobMonitoringConfig.Severity
	var recordStdErr = jobMonitoringConfig.RecordStdErr
	var recordStdOut = jobMonitoringConfig.RecordStdOut

	var jobID = *jobIDPtr
	if jobIDPtr == nil || jobID == "" {
		return nil, errors.New("job ID not specified")
	}
	if len(jobID) > maxJobIDLength {
		return nil, fmt.Errorf("job ID too long. Must be <= %d characters", maxJobIDLength)
	}

	if severityPtr != nil && *severityPtr != "" {
		severity = jobmon.Severity(*severityPtr)
		if !jobmon.IsValidJobMonitoringSeverity(jobmon.Severity(*severityPtr)) {
			return nil, errors.New("specified severity is invalid")
		}
	}

	var nextRun *time.Duration
	if isFlagPassed("nr") && nextRunInPtr != nil {
		if *nextRunInPtr < minNextRunInterval {
			return nil, errors.New("'next run' interval should be more than 5 minutes")
		}
		nextRun = nextRunInPtr
	}

	var maxExecTime *time.Duration
	if isFlagPassed("me") && maxExecutionTimePtr != nil {
		if *maxExecutionTimePtr < minValueForMaxExecTime {
			return nil, errors.New("max execution time should be more than 1 second or unspecified")
		}
		maxExecTime = maxExecutionTimePtr
	}

	if isFlagPassed("re") && recordStdErrPtr != nil {
		recordStdErr = *recordStdErrPtr
	}

	if isFlagPassed("ro") && recordStdOutPtr != nil {
		recordStdOut = *recordStdOutPtr
	}

	restArgs := flag.Args()
	if len(restArgs) > 0 && restArgs[0] == "--" {
		restArgs = restArgs[1:]
	}

	if len(restArgs) == 0 {
		return nil, errors.New("please specify command to execute")
	}

	return &jobmon.JobRunConfig{
		JobID:            jobID,
		Severity:         severity,
		NextRunInterval:  nextRun,
		MaxExecutionTime: maxExecTime,
		RecordStdErr:     recordStdErr,
		RecordStdOut:     recordStdOut,
		Command:          restArgs,
	}, nil
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
