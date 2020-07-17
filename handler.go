package cagent

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/cloudradar-monitoring/selfupdate"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/hwinfo"
	"github.com/cloudradar-monitoring/cagent/pkg/jobmon"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/docker"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/networking"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/processes"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/sensors"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/services"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/updates"
)

type Cleaner interface {
	Cleanup() error
}

// cleanupCommand allow to group multiple Cleanup steps into one object
type cleanupCommand struct {
	steps []func() error
}

func (c *cleanupCommand) AddStep(f func() error) {
	c.steps = append(c.steps, f)
}

func (c *cleanupCommand) Cleanup() error {
	errs := common.ErrorCollector{}
	for _, step := range c.steps {
		errs.Add(step())
	}
	return errs.Combine()
}

func (ca *Cagent) Run(outputFile *os.File, interrupt chan struct{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Unexpected error occurred (main routine): %s", err)
			panic(err)
		}
	}()

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
	measurements, cleaner := ca.collectMeasurements(fullMode)
	err := ca.reportMeasurements(measurements, outputFile)
	if err == nil {
		err = cleaner.Cleanup()
	}
	return err
}

func (ca *Cagent) collectMeasurements(fullMode bool) (common.MeasurementsMap, Cleaner) {
	var errCollector = common.ErrorCollector{}
	var cleanupCommand = &cleanupCommand{}
	var measurements = make(common.MeasurementsMap)
	var cfg = ca.Config

	cpum, err := ca.CPUWatcher().Results()
	errCollector.Add(err)
	measurements = measurements.AddWithPrefix("cpu.", cpum)

	fsResults, err := ca.GetFileSystemWatcher().Results()
	errCollector.Add(err)
	measurements = measurements.AddWithPrefix("fs.", fsResults)

	var memStat *mem.VirtualMemoryStat
	if ca.Config.MemMonitoring {
		var mem common.MeasurementsMap
		mem, memStat, err = ca.MemResults()
		errCollector.Add(err)
		measurements = measurements.AddWithPrefix("mem.", mem)
	}

	cpuUtilisationAnalysisResult, cpuUtilisationAnalysisIsActive, err := ca.CPUUtilisationAnalyser().Results()
	errCollector.Add(err)
	measurements = measurements.AddWithPrefix("cpu_utilisation_analysis.", cpuUtilisationAnalysisResult)
	if cpuUtilisationAnalysisIsActive {
		measurements = measurements.AddWithPrefix(
			"cpu_utilisation_analysis.",
			common.MeasurementsMap{"settings": cfg.CPUUtilisationAnalysis},
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

		proc, processList, err := processes.GetMeasurements(memStat, &ca.Config.ProcessMonitoring)
		errCollector.Add(err)
		measurements = measurements.AddWithPrefix("proc.", proc)

		ports, err := ca.PortsResult(processList)
		errCollector.Add(err)
		measurements = measurements.AddWithPrefix("listeningports.", ports)

		if ca.Config.MemMonitoring {
			swap, err := ca.SwapResults()
			errCollector.Add(err)
			measurements = measurements.AddWithPrefix("swap.", swap)
		}

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

		if cfg.SystemUpdatesChecks.Enabled && cfg.SystemUpdatesChecks.CheckInterval > 0 {
			watcher := updates.GetWatcher(cfg.SystemUpdatesChecks.FetchTimeout, cfg.SystemUpdatesChecks.CheckInterval)
			u, err := watcher.GetSystemUpdatesInfo()
			if err != updates.ErrorDisabledOnHost {
				errCollector.Add(err)
				var prefix string
				if runtime.GOOS == "windows" {
					prefix = "windows_update."
				} else {
					prefix = "linux_update."
				}
				measurements = measurements.AddWithPrefix(prefix, u)
			}
		}

		servicesList, err := services.ListServices(cfg.DiscoverAutostartingServicesOnly)
		if err != services.ErrorNotImplementedForOS {
			errCollector.Add(err)
		}
		measurements = measurements.AddWithPrefix("services.", servicesList)

		if cfg.DockerMonitoring.Enabled {
			containersList, err := docker.ListContainers()
			if err != docker.ErrorNotImplementedForOS && err != docker.ErrorDockerNotAvailable {
				errCollector.Add(err)
			}
			measurements = measurements.AddWithPrefix("docker.", containersList)
		}

		if cfg.TemperatureMonitoring {
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

		spool := jobmon.NewSpoolManager(cfg.JobMonitoring.SpoolDirPath, log.StandardLogger())
		ids, jobs, err := spool.GetFinishedJobs()
		errCollector.Add(err)
		measurements = measurements.AddWithPrefix("", common.MeasurementsMap{"jobmon": jobs})
		cleanupCommand.AddStep(func() error {
			return spool.RemoveJobs(ids)
		})
	}

	measurements["operation_mode"] = cfg.OperationMode

	if errCollector.HasErrors() {
		measurements["message"] = errCollector.Combine()
		measurements["cagent.success"] = 0
	} else {
		measurements["cagent.success"] = 1
	}

	return measurements, cleanupCommand
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
	if ca.Config.Updates.Enabled {
		ca.selfUpdater = selfupdate.StartChecking()
	}

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
