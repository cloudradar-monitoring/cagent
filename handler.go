package cagent

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/hwinfo"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/docker"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/services"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat"
	vmstattypes "github.com/cloudradar-monitoring/cagent/pkg/monitoring/vmstat/types"
)

var ErrorTestWinUISettingsAreEmpty = errors.New("Please fill 'HUB URL', 'HUB USER' and 'HUB PASSWORD' from your Cloudradar account")

func (ca *Cagent) initHubHTTPClient() {
	if ca.hubHTTPClient == nil {
		tr := *(http.DefaultTransport.(*http.Transport))
		if ca.rootCAs != nil {
			tr.TLSClientConfig = &tls.Config{RootCAs: ca.rootCAs}
		}
		if ca.Config.HubProxy != "" {
			if !strings.HasPrefix(ca.Config.HubProxy, "http://") {
				ca.Config.HubProxy = "http://" + ca.Config.HubProxy
			}

			u, err := url.Parse(ca.Config.HubProxy)

			if err != nil {
				log.Errorf("Failed to parse 'hub_proxy' URL")
			} else {
				if ca.Config.HubProxyUser != "" {
					u.User = url.UserPassword(ca.Config.HubProxyUser, ca.Config.HubProxyPassword)
				}
				tr.Proxy = func(_ *http.Request) (*url.URL, error) {
					return u, nil
				}
			}
		}

		ca.hubHTTPClient = &http.Client{
			Timeout:   time.Second * 30,
			Transport: &tr,
		}
	}
}

func (ca *Cagent) TestHub() error {
	if ca.Config.HubURL == "" {
		return fmt.Errorf("please fill config with hub_url, hub_user and hub_password from your Cloudradar account")
	}

	var u *url.URL
	var err error
	if u, err = url.Parse(ca.Config.HubURL); err != nil {
		return fmt.Errorf("can't parse hub_url: %s. Make sure to put the right params from your Cloudradar account", err.Error())
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("wrong scheme: hub_url must start with 'https' or 'http'")
	}

	ca.initHubHTTPClient()
	req, err := http.NewRequest("HEAD", ca.Config.HubURL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", ca.userAgent())
	if ca.Config.HubUser != "" {
		req.SetBasicAuth(ca.Config.HubUser, ca.Config.HubPassword)
	}

	resp, err := ca.hubHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to connect. %s. If you have a proxy or firewall, it may be blocking the connection", err.Error())
	}

	if resp.StatusCode == 401 && ca.Config.HubUser == "" {
		return fmt.Errorf("unable to authorise without credentials. Please set hub_user & hub_password in the сonfig according to your Cloudradar account")
	} else if resp.StatusCode == 401 && ca.Config.HubUser != "" {
		return fmt.Errorf("unable to authorise with provided credentials. Please correct the hub_user & hub_password in the сonfig according to your Cloudradar account")
	} else if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("got bad response status: %d, %s. If you have a proxy or firewall it may be blocking the connection", resp.StatusCode, resp.Status)
	}

	return nil
}

func (ca *Cagent) TestHubWinUI() error {
	if ca.Config.HubURL == "" {
		return ErrorTestWinUISettingsAreEmpty
	}

	var u *url.URL
	var err error
	if u, err = url.Parse(ca.Config.HubURL); err != nil {
		return fmt.Errorf("Can't parse 'HUB URL': %s. Make sure to put the right params from your Cloudradar account", err.Error())
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("'HUB URL' must start with 'https://' or 'http://'")
	}

	ca.initHubHTTPClient()
	req, err := http.NewRequest("HEAD", ca.Config.HubURL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", ca.userAgent())
	if ca.Config.HubUser != "" {
		req.SetBasicAuth(ca.Config.HubUser, ca.Config.HubPassword)
	}

	resp, err := ca.hubHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("Unable to connect. %s. If you have a proxy or firewall, it may be blocking the connection", err.Error())
	}

	if resp.StatusCode == 401 && ca.Config.HubUser == "" {
		return fmt.Errorf("Unable to authorise without credentials. Please set 'HUB USER' and 'HUB PASSWORD' from your Cloudradar account")
	} else if resp.StatusCode == 401 && ca.Config.HubUser != "" {
		return fmt.Errorf("Unable to authorise with provided credentials. Please correct 'HUB USER' and 'HUB PASSWORD' according to your Cloudradar account")
	} else if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("Got bad response status: %d, %s. If you have a proxy or firewall it may be blocking the connection", resp.StatusCode, resp.Status)
	}

	return nil
}

func (ca *Cagent) PostResultsToHub(result Result) error {
	ca.initHubHTTPClient()

	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	var req *http.Request

	if ca.Config.HubGzip {
		var buffer bytes.Buffer
		zw := gzip.NewWriter(&buffer)
		zw.Write(b)
		zw.Close()
		req, err = http.NewRequest("POST", ca.Config.HubURL, &buffer)
		req.Header.Set("Content-Encoding", "gzip")
	} else {
		req, err = http.NewRequest("POST", ca.Config.HubURL, bytes.NewBuffer(b))
	}
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", ca.userAgent())

	if ca.Config.HubUser != "" {
		req.SetBasicAuth(ca.Config.HubUser, ca.Config.HubPassword)
	}

	resp, err := ca.hubHTTPClient.Do(req)

	if err != nil {
		return err
	}

	log.Debugf("Sent to HUB.. Status %d", resp.StatusCode)

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return errors.New(resp.Status)
	}

	return nil
}

func (ca *Cagent) GetAllMeasurements() (MeasurementsMap, error) {
	var errs []string
	var measurements = make(MeasurementsMap)

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

	proc, err := ca.ProcessesResult()
	if err != nil {
		// no need to log because already done inside ProcessesResult()
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("proc.", proc)

	mem, err := ca.MemResults()
	if err != nil {
		// no need to log because already done inside MemResults()
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("mem.", mem)

	swap, err := ca.SwapResults()
	if err != nil {
		// no need to log because already done inside MemResults()
		errs = append(errs, err.Error())
	}

	measurements = measurements.AddWithPrefix("swap.", swap)

	ca.getVMStatMeasurements(func(name string, meas MeasurementsMap, err error) {
		if err == nil {
			measurements = measurements.AddWithPrefix("virt."+name+".", meas)
		} else {
			errs = append(errs, err.Error())
		}
	})

	ca.hwInventory.Do(func() {
		if hwInfo, err := hwinfo.Inventory(); err != nil && err != hwinfo.ErrNotPresent {
			errs = append(errs, err.Error())
		} else if err == nil {
			measurements = measurements.AddInnerWithPrefix("hw.inventory", hwInfo)
		}
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

	cpuUtilisationAnalyser, err := ca.CPUUtilisationAnalyser().Results()
	if err != nil {
		// no need to log because already done inside
		errs = append(errs, err.Error())
	}
	measurements = measurements.AddWithPrefix("cpu_utilisation_analysis.", cpuUtilisationAnalyser)
	measurements = measurements.AddWithPrefix(
		"cpu_utilisation_analysis.",
		MeasurementsMap{"settings": ca.Config.CPUUtilisationAnalysis},
	)

	if len(errs) == 0 {
		measurements["cagent.success"] = 1

		return measurements, nil
	}

	measurements["cagent.success"] = 0

	return measurements, errors.New(strings.Join(errs, "; "))
}

func (ca *Cagent) ReportMeasurements(measurements MeasurementsMap, outputFile *os.File) error {
	result := Result{
		Timestamp:    time.Now().Unix(),
		Measurements: measurements,
	}

	if outputFile != nil {
		jsonEncoder := json.NewEncoder(outputFile)
		err := jsonEncoder.Encode(&result)
		if err != nil {
			return fmt.Errorf("Results json encode error: %s", err.Error())
		}

		return nil
	}

	err := ca.PostResultsToHub(result)
	if err != nil {
		return fmt.Errorf("POST to hub error: %s", err.Error())
	}

	return nil
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

func (ca *Cagent) getVMStatMeasurements(f func(string, MeasurementsMap, error)) {
	ca.vmstatLazyInit.Do(func() {
		if err := vmstat.Init(); err != nil {
			log.Error("vmstat: cannot instantiate virtual machines API: ", err.Error())
			return
		}

		for _, name := range ca.Config.VirtualMachinesStat {
			vm, err := vmstat.Acquire(name)
			if err != nil {
				if err != vmstattypes.ErrNotAvailable {
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
