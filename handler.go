package cagent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/hwinfo"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/docker"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/sensors"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/services"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat"
	vmstatTypes "github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat/types"
	"github.com/cloudradar-monitoring/cagent/types"
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
			Timeout:   30 * time.Second,
			Transport: transport,
		}
	})
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

	if len(ca.Config.HubURL) == 0 {
		return newEmptyFieldError(fieldHubURL)
	} else if u, err := url.Parse(ca.Config.HubURL); err != nil {
		err = errors.WithStack(err)
		return newFieldError(fieldHubURL, err)
	} else if u.Scheme != "http" && u.Scheme != "https" {
		err := errors.Errorf("wrong scheme '%s', URL must start with http:// or https://", u.Scheme)
		return newFieldError(fieldHubURL, err)
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
	_, _ = io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		if len(ca.Config.HubUser) == 0 {
			return newEmptyFieldError(fieldHubUser)
		} else if len(ca.Config.HubPassword) == 0 {
			return newEmptyFieldError(fieldHubPassword)
		}
		err := errors.Errorf("unable to authorize with provided Hub credentials (HTTP %d)", resp.StatusCode)
		return err
	} else if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		err := errors.Errorf("unable to authorize with provided Hub credentials (HTTP %d)", resp.StatusCode)
		return err
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

	if len(ca.Config.HubURL) == 0 {
		return newEmptyFieldError("hub_url")
	} else if u, err := url.Parse(ca.Config.HubURL); err != nil {
		err = errors.WithStack(err)
		return newFieldError("hub_url", err)
	} else if u.Scheme != "http" && u.Scheme != "https" {
		err := errors.Errorf("wrong scheme '%s', URL must start with http:// or https://", u.Scheme)
		return newFieldError("hub_url", err)
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

func (ca *Cagent) GetAllMeasurements() (types.MeasurementsMap, error) {
	var errs []string
	var measurements = make(types.MeasurementsMap)

	cpum, err := ca.CPUWatcher().Results()
	if err != nil {
		// no need to log because already done inside cpu.Results()
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("cpu.", cpum)

	info, err := ca.HostInfoResults()
	if err != nil {
		// no need to log because already done inside HostInfoResults()
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("system.", info)

	ipResults, err := IPAddresses()
	if err != nil {
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("system.", ipResults)

	fsResults, err := ca.FSWatcher().Results()
	if err != nil {
		// no need to log because already done inside fs.Results()
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("fs.", fsResults)

	netResults, err := ca.NetWatcher().Results()
	if err != nil {
		// no need to log because already done inside net.Results()
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("net.", netResults)

	mem, memStat, err := ca.MemResults()
	if err != nil {
		// no need to log because already done inside MemResults()
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("mem.", mem)

	proc, processList, err := ca.ProcessesResult(memStat)
	if err != nil {
		// no need to log because already done inside ProcessesResult()
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("proc.", proc)

	ports, err := ca.PortsResult(processList)
	if err != nil {
		// no need to log because already done inside PortsResult()
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("listeningports.", ports)

	swap, err := ca.SwapResults()
	if err != nil {
		// no need to log because already done inside MemResults()
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("swap.", swap)

	ca.getVMStatMeasurements(func(name string, meas types.MeasurementsMap, err error) {
		if err == nil {
			measurements = measurements.AddWithPrefix("virt."+name+".", meas)
		} else {
			errs = append(errs, err.Error())
		}
	})

	ca.hwInventory.Do(func() {
		hwInfo, err := hwinfo.Inventory()
		if err != nil {
			errs = append(errs, err.Error())
		}
		measurements = measurements.AddInnerWithPrefix("hw.inventory", hwInfo)
	})

	if runtime.GOOS == "linux" {
		raid, err := ca.RaidState()
		if err != nil {
			// no need to log because already done inside RaidState()
			errs = append(errs, err.Error())
		}

		measurements = measurements.AddWithPrefix("raid.", raid)
	}

	if runtime.GOOS == "windows" {
		wu, err := ca.WindowsUpdatesWatcher().WindowsUpdates()
		if err != nil {
			// no need to log because already done inside MemResults()
			errs = append(errs, err.Error())
		}

		measurements = measurements.AddWithPrefix("windows_update.", wu)
	}

	servicesList, err := services.ListServices(ca.Config.DiscoverAutostartingServicesOnly)
	if err != nil && err != services.ErrorNotImplementedForOS {
		// no need to log because already done inside ListServices()
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("services.", servicesList)

	containersList, err := ca.dockerWatcher.ListContainers()
	if err != nil && err != docker.ErrorNotImplementedForOS && err != docker.ErrorDockerNotFound {
		// no need to log because already done inside ListContainers()
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("docker.", containersList)

	temperatures, err := sensors.ReadTemperatureSensors()
	if err != nil {
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("temperatures.", types.MeasurementsMap{"list": temperatures})

	cpuUtilisationAnalysisResult, cpuUtilisationAnalysisIsActive, err := ca.CPUUtilisationAnalyser().Results()
	if err != nil {
		// no need to log because already done inside
		errs = append(errs, err.Error())
	}
	measurements = measurements.AddWithPrefix("cpu_utilisation_analysis.", cpuUtilisationAnalysisResult)
	if cpuUtilisationAnalysisIsActive {
		measurements = measurements.AddWithPrefix(
			"cpu_utilisation_analysis.",
			types.MeasurementsMap{"settings": ca.Config.CPUUtilisationAnalysis},
		)
	}

	if len(errs) == 0 {
		measurements["cagent.success"] = 1

		return measurements, nil
	}

	measurements["cagent.success"] = 0

	return measurements, errors.New(strings.Join(errs, "; "))
}

func (ca *Cagent) ReportMeasurements(measurements types.MeasurementsMap, outputFile *os.File) error {
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
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancelFn()

	err := ca.PostResultToHub(ctx, result)
	if err != nil {
		err = errors.Wrap(err, "failed to POST measurement result to Hub")
	}
	return err
}

func (ca *Cagent) RunOnce(outputFile *os.File) error {
	measurements, err := ca.GetAllMeasurements()
	if err != nil {
		// don't need to log or return it here – just add the message to report
		// it is already logged (down into the GetAllMeasurements)
		measurements["message"] = err.Error()
	}

	return ca.ReportMeasurements(measurements, outputFile)
}

func (ca *Cagent) Run(outputFile *os.File, interrupt chan struct{}, cfg *Config) {
	for {
		err := ca.RunOnce(outputFile)
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

func (ca *Cagent) getVMStatMeasurements(f func(string, types.MeasurementsMap, error)) {
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
