package storcli

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring"
)

const cacheExpirationDuration = 30 * time.Minute

type StorCLI struct {
	binaryPath     string
	lastRunAt      time.Time
	lastRunReports []*monitoring.ModuleReport
}

func CreateModule(binaryPath string) monitoring.Module {
	return &StorCLI{binaryPath: binaryPath}
}

func (s *StorCLI) IsEnabled() bool {
	return s.binaryPath != ""
}

func (s *StorCLI) Run() ([]*monitoring.ModuleReport, error) {
	if s.binaryPath == "" {
		return nil, nil
	}

	now := time.Now()
	if !s.lastRunAt.IsZero() && s.lastRunAt.Sub(now) < cacheExpirationDuration {
		return s.lastRunReports, nil
	}

	reports := make([]*monitoring.ModuleReport, 0)
	cmdLineStr := s.getCommandLineCombined()
	cmdExecReport := monitoring.NewReport("storecli execution for hardware raid health", now, cmdLineStr)
	reports = append(reports, &cmdExecReport)

	cmdLine := s.getCommandLine()
	showAllCmd := exec.Command(cmdLine[0], cmdLine[1:]...)
	stderrBuffer := bytes.Buffer{}
	showAllCmd.Stderr = bufio.NewWriter(&stderrBuffer)
	outBytes, err := showAllCmd.Output()
	if err != nil {
		stderrBytes, _ := ioutil.ReadAll(bufio.NewReader(&stderrBuffer))
		stderr := string(stderrBytes)
		errMsg := fmt.Sprintf("Error while invoking storcli command: %s. %s", err.Error(), stderr)
		logrus.Error(errMsg)

		cmdExecReport.Message = errMsg
		cmdExecReport.AddAlert(errMsg)
		return reports, nil
	}

	parsedOutput, err := tryParseCmdOutput(&outBytes)
	if err != nil {
		logrus.WithError(err).Error()
		cmdExecReport.Message = err.Error()
		cmdExecReport.AddAlert(err.Error())
		return reports, nil
	}

	if len(parsedOutput.Controllers) < 1 {
		return reports, fmt.Errorf("unexpected storcli JSON: no controller objects present")
	}

	// there is always at least one controller object even if there is no RAID installed
	firstController := parsedOutput.Controllers[0]

	cmdExecReport.Measurements = map[string]interface{}{
		"Command Status": firstController.CommandStatus,
	}

	if firstController.CommandStatus.Status != "Failure" {
		cmdExecReport.Measurements["Detected Controllers"] = len(parsedOutput.Controllers)

		for _, c := range parsedOutput.Controllers {
			cBasicInfo := &c.ResponseData.Basics
			cid := cBasicInfo.ControllerID
			cmdExecReport.Measurements[fmt.Sprintf("Controller %d", cid)] = cBasicInfo.GetDisplayName()

			measurements, alerts, warnings, err := getReportData(&c.ResponseData)
			if err != nil {
				logrus.WithError(err).Error()
				continue
			}

			r := monitoring.NewReport(getModuleReportName(cid), now, cmdLineStr)
			r.Measurements = measurements
			r.Alerts = append(r.Alerts, alerts...)
			r.Warnings = append(r.Warnings, warnings...)
			reports = append(reports, &r)
		}
	}

	s.lastRunReports = reports
	s.lastRunAt = now

	return reports, nil
}

func (s *StorCLI) GetDescription() string {
	return "storcli monitoring"
}

func getModuleReportName(controllerID int) string {
	return fmt.Sprintf("storecli hardware raid health controller c%d", controllerID)
}

func (s *StorCLI) getCommandLineCombined() string {
	return strings.Join(s.getCommandLine(), " ")
}
