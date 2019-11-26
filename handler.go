package cagent

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/hwinfo"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/docker"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/networking"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/sensors"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/services"
)

func (ca *Cagent) Run(outputFile *os.File, interrupt chan struct{}) {
	for {
		err := ca.RunOnce(outputFile, ca.Config.OperationMode == OperationModeFull)
		if err != nil {
			log.Error(err)
		}

		select {
		case <-interrupt:
			return
		case <-time.After(secToDuration(ca.Config.Interval)):
			continue
		}
	}
}

func (ca *Cagent) RunOnce(outputFile *os.File, fullMode bool) error {
	measurements := ca.collectMeasurements(fullMode)
	return ca.reportMeasurements(measurements, outputFile)
}

func (ca *Cagent) collectMeasurements(fullMode bool) common.MeasurementsMap {
	var errCollector = common.ErrorCollector{}
	var measurements = make(common.MeasurementsMap)

	cpum, err := ca.CPUWatcher().Results()
	errCollector.Add(err)
	measurements = measurements.AddWithPrefix("cpu.", cpum)

	fsResults, err := ca.GetFileSystemWatcher().Results()
	errCollector.Add(err)
	measurements = measurements.AddWithPrefix("fs.", fsResults)

	mem, memStat, err := ca.MemResults()
	errCollector.Add(err)
	measurements = measurements.AddWithPrefix("mem.", mem)

	cpuUtilisationAnalysisResult, cpuUtilisationAnalysisIsActive, err := ca.CPUUtilisationAnalyser().Results()
	errCollector.Add(err)
	measurements = measurements.AddWithPrefix("cpu_utilisation_analysis.", cpuUtilisationAnalysisResult)
	if cpuUtilisationAnalysisIsActive {
		measurements = measurements.AddWithPrefix(
			"cpu_utilisation_analysis.",
			common.MeasurementsMap{"settings": ca.Config.CPUUtilisationAnalysis},
		)
	}

	if fullMode {
		info, err := ca.HostInfoResults()
		errCollector.Add(err)
		measurements = measurements.AddWithPrefix("system.", info)

		ipResults, err := networking.IPAddresses()
		errCollector.Add(err)
		measurements = measurements.AddWithPrefix("system.", ipResults)

		netResults, err := ca.GetNetworkWatcher().Results()
		errCollector.Add(err)
		measurements = measurements.AddWithPrefix("net.", netResults)

		proc, processList, err := ca.ProcessesResult(memStat)
		errCollector.Add(err)
		measurements = measurements.AddWithPrefix("proc.", proc)

		ports, err := ca.PortsResult(processList)
		errCollector.Add(err)
		measurements = measurements.AddWithPrefix("listeningports.", ports)

		swap, err := ca.SwapResults()
		errCollector.Add(err)
		measurements = measurements.AddWithPrefix("swap.", swap)

		ca.getVMStatMeasurements(func(name string, meas common.MeasurementsMap, err error) {
			if err == nil {
				measurements = measurements.AddWithPrefix("virt."+name+".", meas)
			}
			errCollector.Add(err)
		})

		ca.hwInventory.Do(func() {
			hwInfo, err := hwinfo.Inventory()
			errCollector.Add(err)
			if hwInfo != nil {
				measurements = measurements.AddInnerWithPrefix("hw.inventory", hwInfo)
			}
		})

		if runtime.GOOS == "windows" {
			wu, err := ca.WindowsUpdatesWatcher().WindowsUpdates()
			errCollector.Add(err)
			measurements = measurements.AddWithPrefix("windows_update.", wu)
		}

		servicesList, err := services.ListServices(ca.Config.DiscoverAutostartingServicesOnly)
		if err != services.ErrorNotImplementedForOS {
			errCollector.Add(err)
		}
		measurements = measurements.AddWithPrefix("services.", servicesList)

		containersList, err := docker.ListContainers()
		if err != docker.ErrorNotImplementedForOS && err != docker.ErrorDockerNotAvailable {
			errCollector.Add(err)
		}
		measurements = measurements.AddWithPrefix("docker.", containersList)

		if ca.Config.TemperatureMonitoring {
			temperatures, err := sensors.ReadTemperatureSensors()
			errCollector.Add(err)
			measurements = measurements.AddWithPrefix("temperatures.", common.MeasurementsMap{"list": temperatures})
		}

		moduleReports, err := ca.collectModulesMeasurements()
		errCollector.Add(err)
		measurements = measurements.AddWithPrefix("", common.MeasurementsMap{"modules": moduleReports})

		smartMeas := ca.getSMARTMeasurements()
		if len(smartMeas) > 0 {
			measurements = measurements.AddInnerWithPrefix("smartmon", smartMeas)
		}
	}

	measurements["operation_mode"] = ca.Config.OperationMode

	if errCollector.HasErrors() {
		measurements["message"] = errCollector.Combine()
		measurements["cagent.success"] = 0
	} else {
		measurements["cagent.success"] = 1
	}

	return measurements
}

func (ca *Cagent) reportMeasurements(measurements common.MeasurementsMap, outputFile *os.File) error {
	result := &Result{
		Timestamp:    time.Now().Unix(),
		Measurements: measurements,
	}
	if outputFile != nil {
		err := json.NewEncoder(outputFile).Encode(result)
		if err != nil {
			return errors.Wrap(err, "failed to JSON encode measurement result")
		}
		return nil
	}

	if ca.Config.Logs.HubFile != "" {
		ca.prettyPrintMeasurementsToFile(measurements, ca.Config.Logs.HubFile)
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancelFn()

	err := ca.PostResultToHub(ctx, result)
	if err != nil {
		err = errors.Wrap(err, "failed to POST measurement result to Hub")
	}

	return err
}

func (ca *Cagent) RunHeartbeat(interrupt chan struct{}) {
	for {
		err := ca.sendHeartbeat()
		if err != nil {
			log.WithError(err).Error("failed to send heartbeat to Hub")
		}

		select {
		case <-interrupt:
			return
		case <-time.After(secToDuration(ca.Config.HeartbeatInterval)):
			continue
		}
	}
}

func (ca *Cagent) sendHeartbeat() error {
	ca.initHubClientOnce()
	err := ca.validateHubURL("hub_url")
	if err != nil {
		return err
	}

	// no need to wait more than heartbeat interval
	ctx, cancelFn := context.WithTimeout(context.Background(), secToDuration(ca.Config.HeartbeatInterval))
	defer cancelFn()

	req, err := http.NewRequest("GET", ca.Config.HubURL, nil)
	if err != nil {
		return errors.WithStack(err)
	}
	req.Header.Add("User-Agent", ca.userAgent())
	if len(ca.Config.HubUser) > 0 {
		req.SetBasicAuth(ca.Config.HubUser, ca.Config.HubPassword)
	}
	req = req.WithContext(ctx)
	resp, err := ca.hubClient.Do(req)
	if err = ca.checkClientError(resp, err, "hub_user", "hub_password"); err != nil {
		return errors.WithStack(err)
	}
	log.Debugf("Heartbeat sent. Status: %d", resp.StatusCode)
	return err
}

func (ca *Cagent) prettyPrintMeasurementsToFile(measurements common.MeasurementsMap, file string) {
	fl, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.WithError(err).Error("failed to open hub log file")
		return
	}

	defer fl.Close()

	enc := json.NewEncoder(fl)
	enc.SetIndent("", "    ")
	if err = enc.Encode(measurements); err != nil {
		log.WithError(err).Error("failed to JSON encode pretty printed measurement result to log file")
	}
}

func secToDuration(seconds float64) time.Duration {
	return time.Duration(int64(float64(time.Second) * seconds))
}
