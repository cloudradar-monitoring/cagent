package cagent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/hwinfo"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/docker"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/networking"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/sensors"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/services"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat"
	vmstatTypes "github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat/types"
)

func (ca *Cagent) initHubClientOnce() {
	ca.hubClientOnce.Do(func() {
		transport := &http.Transport{
			ResponseHeaderTimeout: 15 * time.Second,
		}
		if ca.rootCAs != nil {
			transport.TLSClientConfig = &tls.Config{
				RootCAs: ca.rootCAs,
			}
		}
		if len(ca.Config.HubProxy) > 0 {
			if !strings.HasPrefix(ca.Config.HubProxy, "http://") {
				ca.Config.HubProxy = "http://" + ca.Config.HubProxy
			}
			proxyURL, err := url.Parse(ca.Config.HubProxy)
			if err != nil {
				log.WithFields(log.Fields{
					"url": ca.Config.HubProxy,
				}).Warningln("failed to parse hub_proxy URL")
			} else {
				if len(ca.Config.HubProxyUser) > 0 {
					proxyURL.User = url.UserPassword(ca.Config.HubProxyUser, ca.Config.HubProxyPassword)
				}
				transport.Proxy = func(_ *http.Request) (*url.URL, error) {
					return proxyURL, nil
				}
			}
		}
		ca.hubClient = &http.Client{
			Timeout:   time.Duration(ca.Config.HubRequestTimeout) * time.Second,
			Transport: transport,
		}
	})
}

// validateHubURL performs Hub URL validation, that reference field name as in source config.
func (ca *Cagent) validateHubURL(fieldHubURL string) error {
	if len(ca.Config.HubURL) == 0 {
		return newEmptyFieldError(fieldHubURL)
	} else if u, err := url.Parse(ca.Config.HubURL); err != nil {
		err = errors.WithStack(err)
		return newFieldError(fieldHubURL, err)
	} else if u.Scheme != "http" && u.Scheme != "https" {
		err := errors.Errorf("wrong scheme '%s', URL must start with http:// or https://", u.Scheme)
		return newFieldError(fieldHubURL, err)
	}
	return nil
}

// CheckHubCredentials performs credentials check for a Hub config, returning errors that reference
// field names as in source config. Since config may be filled from file or UI, the field names can be different.
// Consider also localization of UI, we want to decouple credential checking logic from their actual view in UI.
//
// Examples:
// * for TOML: CheckHubCredentials(ctx, "hub_url", "hub_user", "hub_password")
// * for WinUI: CheckHubCredentials(ctx, "URL", "User", "Password")
func (ca *Cagent) CheckHubCredentials(ctx context.Context, fieldHubURL, fieldHubUser, fieldHubPassword string) error {
	ca.initHubClientOnce()
	err := ca.validateHubURL(fieldHubURL)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("HEAD", ca.Config.HubURL, nil)
	req.Header.Add("User-Agent", ca.userAgent())
	if len(ca.Config.HubUser) > 0 {
		req.SetBasicAuth(ca.Config.HubUser, ca.Config.HubPassword)
	}

	ctx, cancelFn := context.WithTimeout(ctx, time.Minute)
	req = req.WithContext(ctx)
	resp, err := ca.hubClient.Do(req)
	cancelFn()
	if err = ca.checkClientError(resp, err, fieldHubUser, fieldHubPassword); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (ca *Cagent) checkClientError(resp *http.Response, err error, fieldHubUser, fieldHubPassword string) error {
	if err != nil {
		if errors.Cause(err) == context.DeadlineExceeded {
			err = errors.New("connection timeout, please check your proxy or firewall settings")
			return err
		}
		return err
	}

	var responseBody string
	responseBodyBytes, readBodyErr := ioutil.ReadAll(resp.Body)
	if readBodyErr == nil {
		responseBody = string(responseBodyBytes)
	}

	_ = resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		if len(ca.Config.HubUser) == 0 {
			return newEmptyFieldError(fieldHubUser)
		} else if len(ca.Config.HubPassword) == 0 {
			return newEmptyFieldError(fieldHubPassword)
		}
		return errors.Errorf("unable to authorize with provided Hub credentials (HTTP %d). %s", resp.StatusCode, responseBody)
	} else if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return errors.Errorf("got unexpected response from server (HTTP %d). %s", resp.StatusCode, responseBody)
	}
	return nil
}

func newEmptyFieldError(name string) error {
	err := errors.Errorf("unexpected empty field %s", name)
	return errors.Wrap(err, "the field must be filled with details of your Cloudradar account")
}

func newFieldError(name string, err error) error {
	return errors.Wrapf(err, "%s field verification failed", name)
}

func (ca *Cagent) PostResultToHub(ctx context.Context, result *Result) error {
	ca.initHubClientOnce()
	err := ca.validateHubURL("hub_url")
	if err != nil {
		return err
	}

	b, err := json.Marshal(result)
	if err != nil {
		err = errors.Wrap(err, "failed to serialize result")
		return err
	}

	var req *http.Request
	if ca.Config.HubGzip {
		buf := new(bytes.Buffer)
		gzipped := gzip.NewWriter(buf)
		if _, err := gzipped.Write(b); err != nil {
			err = errors.Wrap(err, "failed to write into gzipped buffer")
			return err
		}
		if err := gzipped.Close(); err != nil {
			err = errors.Wrap(err, "failed to finalize gzipped buffer")
			return err
		}
		req, err = http.NewRequest("POST", ca.Config.HubURL, buf)
		req.Header.Set("Content-Encoding", "gzip")
	} else {
		req, err = http.NewRequest("POST", ca.Config.HubURL, bytes.NewBuffer(b))
	}
	if err != nil {
		return errors.WithStack(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("User-Agent", ca.userAgent())
	if len(ca.Config.HubUser) > 0 {
		req.SetBasicAuth(ca.Config.HubUser, ca.Config.HubPassword)
	}
	req = req.WithContext(ctx)
	resp, err := ca.hubClient.Do(req)
	if err = ca.checkClientError(resp, err, "hub_user", "hub_password"); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (ca *Cagent) CollectMeasurements(full bool) (common.MeasurementsMap, error) {
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

	if full {
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

	if errCollector.HasErrors() {
		measurements["cagent.success"] = 0
	} else {
		measurements["cagent.success"] = 1
	}

	measurements["operation_mode"] = ca.Config.OperationMode

	if errCollector.HasErrors() {
		return measurements, errCollector.Combine()
	}

	return measurements, nil
}

func (ca *Cagent) ReportMeasurements(measurements common.MeasurementsMap, outputFile *os.File) error {
	result := &Result{
		Timestamp:    time.Now().Unix(),
		Measurements: measurements,
	}
	if outputFile != nil {
		err := json.NewEncoder(outputFile).Encode(result)
		if err != nil {
			err = errors.Wrap(err, "failed to JSON encode measurement result")
			return err
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

func (ca *Cagent) RunOnce(outputFile *os.File, full bool) error {
	measurements, err := ca.CollectMeasurements(full)
	if err != nil {
		// don't need to log or return it here – just add the message to report
		// it is already logged (down into the CollectMeasurements)
		measurements["message"] = err.Error()
	}

	return ca.ReportMeasurements(measurements, outputFile)
}

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

func (ca *Cagent) getVMStatMeasurements(f func(string, common.MeasurementsMap, error)) {
	ca.vmstatLazyInit.Do(func() {
		if err := vmstat.Init(); err != nil {
			log.Error("vmstat: cannot instantiate virtual machines API: ", err.Error())
			return
		}

		for _, name := range ca.Config.VirtualMachinesStat {
			vm, err := vmstat.Acquire(name)
			if err != nil {
				if err != vmstatTypes.ErrNotAvailable {
					log.Warnf("vmstat: Error while acquiring vm provider \"%s\": %s", name, err.Error())
				}
			} else {
				ca.vmWatchers[name] = vm
			}
		}
	})

	for name, p := range ca.vmWatchers {
		res, err := p.GetMeasurements()
		f(name, res, err)
	}
}

func (ca *Cagent) getSMARTMeasurements() common.MeasurementsMap {
	// measurements fetched below should not affect cagent.success
	if ca.smart != nil {
		res, errs := ca.smart.Parse()

		if len(errs) > 0 {
			var errStr []string
			for _, e := range errs {
				errStr = append(errStr, e.Error())
			}

			if res == nil {
				res = make(common.MeasurementsMap)
			}

			res["messages"] = strings.Join(errStr, "; ")
		}

		return res
	}

	return nil
}
