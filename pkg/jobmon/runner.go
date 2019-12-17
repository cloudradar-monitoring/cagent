package jobmon

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/user"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

type Runner struct {
	spool *SpoolManager
	cfg   *JobRunConfig
}

func NewRunner(spoolDirPath string, runConfig *JobRunConfig, logger *logrus.Logger) *Runner {
	return &Runner{
		spool: NewSpoolManager(spoolDirPath, logger),
		cfg:   runConfig,
	}
}

func (r *Runner) RunJob(interruptionSignalsChan chan os.Signal, forceRun bool) error {
	var job = newJobRun(r.cfg)
	var cmd = r.createJobCommand()

	stdOutBuffer := newCaptureWriter(os.Stdout, maxStdStreamBufferSize)
	cmd.Stdout = stdOutBuffer

	stdErrBuffer := newCaptureWriter(os.Stderr, maxStdStreamBufferSize)
	cmd.Stderr = stdErrBuffer

	uid, err := r.spool.NewJob(job, forceRun)
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err == nil {
		var t *time.Timer
		if r.cfg.MaxExecutionTime != nil {
			t = r.createMaxExecutionTimeoutHandler(cmd, job)
		}

		go r.waitForInterruptionSignal(interruptionSignalsChan, cmd, job)

		err = cmd.Wait()
		if t != nil {
			t.Stop()
		}
	}

	endedAt := time.Now()
	var exitCode *int
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			exitCode = &code
		} else {
			job.AddError(err.Error())
		}
	} else {
		code := cmd.ProcessState.ExitCode()
		exitCode = &code
	}

	endTimestamp := common.Timestamp(endedAt)
	job.EndedAt = &endTimestamp
	job.Duration = calcRunDuration(job.StartedAt, endedAt)
	job.ExitCode = exitCode

	if r.cfg.RecordStdOut {
		s := stdOutBuffer.String()
		job.StdOut = &s
	}
	if r.cfg.RecordStdErr {
		s := stdErrBuffer.String()
		job.StdErr = &s
	}

	return r.spool.FinishJob(uid, job)
}

func calcRunDuration(startedAt common.Timestamp, endedAt time.Time) *uint64 {
	d := uint64(math.Round(endedAt.Sub(time.Time(startedAt)).Seconds()))
	return &d
}

func (r *Runner) waitForInterruptionSignal(ch chan os.Signal, cmd *exec.Cmd, runResult *JobRun) {
	<-ch
	if isProcessFinished(cmd) {
		return
	}
	osSpecificCommandTermination(cmd)
	msg := fmt.Sprintf("Jobmon has received an interruption signal and all subprocesses have been terminated. This normally means someone has ended jobmon.")
	runResult.AddError(msg)
}

func (r *Runner) createMaxExecutionTimeoutHandler(cmd *exec.Cmd, runResult *JobRun) *time.Timer {
	timeout := r.cfg.MaxExecutionTime
	return time.AfterFunc(*timeout, func() {
		if isProcessFinished(cmd) {
			return
		}
		osSpecificCommandTermination(cmd)
		msg := fmt.Sprintf(
			"Command has been terminated by jobmon because the maximum execution time of %s exceeded.",
			timeout.String(),
		)
		runResult.AddError(msg)
	})
}

func (r *Runner) createJobCommand() *exec.Cmd {
	commandName := r.cfg.Command[0]
	var commandArgs []string
	if len(r.cfg.Command) > 1 {
		commandArgs = r.cfg.Command[1:]
	}
	cmd := exec.Command(commandName, commandArgs...)
	osSpecificCommandConfig(cmd)
	return cmd
}

func getUsername() string {
	u, err := user.Current()
	if err != nil {
		return "<unknown>"
	}
	return u.Username
}

func isProcessFinished(cmd *exec.Cmd) bool {
	return cmd.ProcessState != nil && cmd.ProcessState.Exited()
}
