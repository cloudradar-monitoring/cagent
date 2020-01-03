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

const (
	usageExamples = `Examples:
  jobmon -id my-rsync-job -- rsync -a /etc /var/backups
  jobmon -id my-rsync-job -ro -- rsync -av /etc /var/backups
  jobmon -id my-rsync-job -re=false -s none -- rsync -av /etc /var/backups
  jobmon -id my-robocopy-job -- robocopy C:\Users\nobody\Downloads "C:\My Backups" /MIR
  jobmon -id my-robocopy-job -nr 24h -- robocopy C:\Users\nobody\Downloads "C:\My Backups" /MIR`
)

var logger *logrus.Logger

type plainFormatter struct {
}

func (f *plainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	fieldValuesFormatted := ""
	for key, value := range entry.Data {
		fieldValuesFormatted = fmt.Sprintf("%s %s=%s", fieldValuesFormatted, key, value)
	}
	return []byte(fmt.Sprintf("%s%s\n", entry.Message, fieldValuesFormatted)), nil
}

func init() {
	// jobmon binary should not write any messages to log file
	logger = logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetFormatter(&plainFormatter{})
}

func main() {
	versionPtr := flag.Bool("version", false, "Show the jobmon version")
	cfgPathPtr := flag.String("c", cagent.DefaultCfgPath, "Config file path")

	jobIDPtr := flag.String("id", "", fmt.Sprintf("id of the job, required, maximum %d characters", maxJobIDLength))
	forceRunPtr := flag.Bool("f", false, "Force run of a job even if the job with the same ID is already running or its termination wasn't handled successfully.")
	severityPtr := flag.String("s", "", "alert|warning|none If job fails (exit code != 0) trigger an event with this severity. 'alert' is used by default.\nOverwrites the default severity of cagent.conf. Severity 'none' suppresses all messages.")
	nextRunInPtr := flag.Duration("nr", 0, "<N>h|m indicates when the job should run for the next time. Allows triggering alerts for not run jobs.\nThe shortest interval is 5 minutes.")
	maxExecutionTimePtr := flag.Duration("me", 0, "<N>h|m|s or just <N> for number of seconds. Max execution time for job.")

	recordStdErrPtr := flag.Bool("re", false, "or -re=true|false\nRecord errors from stderr, overwrites the default settings of cagent.conf.\nLimited to the last 4 KB.\nUse '-re=false' to disable the recording")
	recordStdOutPtr := flag.Bool("ro", false, "or -ro=true|false\nRecord errors from stdout, overwrites the default settings of cagent.conf\nLimited to the last 4 KB.")

	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -id <JOB_ID> {ARGS} -- <COMMAND_TO_EXECUTE> {COMMAND_ARGS}\n", os.Args[0])
		_, _ = fmt.Fprintln(flag.CommandLine.Output(), "")
		flag.PrintDefaults()
		_, _ = fmt.Fprintln(flag.CommandLine.Output(), "")
		_, _ = fmt.Fprintln(flag.CommandLine.Output(), usageExamples)
	}

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
