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
	binaryPath string
	controller uint
	lastRunAt  *time.Time

	message      string
	alerts       []monitoring.Alert
	warnings     []monitoring.Warning
	measurements map[string]interface{}
}

func CreateModules(binaryPath string, controllerList []uint) []monitoring.Module {
	result := make([]monitoring.Module, 0)
	for _, controller := range controllerList {
		result = append(result, &StorCLI{binaryPath: binaryPath, controller: controller})
	}
	return result
}

func (s *StorCLI) IsEnabled() bool {
	return s.binaryPath != ""
}

func (s *StorCLI) Run() error {
	if s.binaryPath == "" {
		return nil
	}

	now := time.Now()
	if s.lastRunAt != nil {
		if s.lastRunAt.Sub(now) < cacheExpirationDuration {
			return nil
		}
	}

	cmdLine := s.getCommandLine()
	showAllCmd := exec.Command(cmdLine[0], cmdLine[1:]...)
	stderrBuffer := bytes.Buffer{}
	showAllCmd.Stderr = bufio.NewWriter(&stderrBuffer)
	outBytes, err := showAllCmd.Output()
	if err != nil {
		stderrBytes, _ := ioutil.ReadAll(bufio.NewReader(&stderrBuffer))
		stderr := string(stderrBytes)
		s.message = fmt.Sprintf("Error while invoking storcli command: %s. %s", err.Error(), stderr)
		logrus.Error(s.message)
		return nil
	}

	s.lastRunAt = &now

	s.measurements, s.alerts, s.warnings, err = tryParseCmdOutput(&outBytes)
	return err
}

func (s *StorCLI) GetName() string {
	return fmt.Sprintf("storecli hardware raid health controller c%d", s.controller)
}

func (s *StorCLI) GetExecutedCommand() string {
	return s.getCommandLineCombined()
}

func (s *StorCLI) GetAlerts() []monitoring.Alert {
	return s.alerts
}

func (s *StorCLI) GetWarnings() []monitoring.Warning {
	return s.warnings
}

func (s *StorCLI) GetMessage() string {
	return s.message
}

func (s *StorCLI) GetMeasurements() map[string]interface{} {
	return s.measurements
}

func (s *StorCLI) getCommandLineCombined() string {
	return strings.Join(s.getCommandLine(), " ")
}
