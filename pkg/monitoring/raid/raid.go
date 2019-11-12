package raid

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring"
)

const (
	raidStatusOptimal    = "optimal"
	raidStatusDegraded   = "degraded"
	raidStatusRebuilding = "rebuilding"
)

var log = logrus.WithField("package", "raid")

type RAID struct {
	mdstatFilePath string
}

func CreateModule() monitoring.Module {
	return &RAID{
		mdstatFilePath: common.HostProc("mdstat"),
	}
}

func (r *RAID) GetDescription() string {
	return fmt.Sprintf("software RAID monitoring using %s file", r.mdstatFilePath)
}

func (r *RAID) IsEnabled() bool {
	return runtime.GOOS == "linux"
}

func (r *RAID) readAndParseMdstat() raidArrays {
	buf, err := ioutil.ReadFile(r.mdstatFilePath)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		log.WithError(err).Errorf("could not read %s file", r.mdstatFilePath)
		return nil
	}

	raidArrays := parseMdstat(string(buf))
	return raidArrays
}

func (r *RAID) Run() ([]*monitoring.ModuleReport, error) {
	raidArrays := r.readAndParseMdstat()
	if len(raidArrays) == 0 {
		return nil, nil
	}

	report := monitoring.NewReport(
		fmt.Sprintf("software raid health according to %s", r.mdstatFilePath),
		time.Now(),
		"",
	)

	virtualDrives := make(map[string]int)
	raidStatuses := make(map[string]string)
	atLeastOneDegraded := false
	for _, raidInfo := range ([]raidInfo)(raidArrays) {
		var status string
		if raidInfo.State == "active" || raidInfo.State == "started" {
			status = raidStatusOptimal
		} else {
			status = raidInfo.State
		}
		raidName := raidInfo.Name
		virtualDrives[fmt.Sprintf("%s raid level", raidName)] = raidInfo.RaidLevel

		if raidInfo.IsRebuilding {
			status = raidStatusRebuilding
			report.AddWarning(fmt.Sprintf("Raid %s rebuilding.", raidName))
		}

		failedDevs := raidInfo.GetFailedDevices()
		if len(failedDevs) > 0 {
			report.AddAlert(fmt.Sprintf(
				"Raid %s degraded. Devices failing: %s.",
				raidName,
				strings.Join(failedDevs, ", "),
			))
			status = raidStatusDegraded
		}

		detectedDeviceStatusesCount := len(raidInfo.Active) + len(raidInfo.Inactive)
		numberOfMissingDevs := detectedDeviceStatusesCount - len(raidInfo.Devices)
		if numberOfMissingDevs > 0 {
			report.AddAlert(fmt.Sprintf("Raid %s degraded. Missing %d devices.", raidName, numberOfMissingDevs))
			status = raidStatusDegraded
		}

		if status == raidStatusDegraded {
			atLeastOneDegraded = true
		}
		raidStatuses[raidName] = status
	}

	if atLeastOneDegraded {
		report.AddAlert("Raid status not optimal (Needs Attention)")
	}

	report.Measurements = map[string]interface{}{
		"Virtual Drives": virtualDrives,
		"Status":         raidStatuses,
	}

	return []*monitoring.ModuleReport{&report}, nil
}
