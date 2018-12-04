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
)

func (ca *Cagent) initHubHTTPClient() {
	if ca.hubHTTPClient == nil {
		tr := *(http.DefaultTransport.(*http.Transport))
		if ca.rootCAs != nil {
			tr.TLSClientConfig = &tls.Config{RootCAs: ca.rootCAs}
		}
		if ca.config.HubProxy != "" {
			if !strings.HasPrefix(ca.config.HubProxy, "http://") {
				ca.config.HubProxy = "http://" + ca.config.HubProxy
			}

			u, err := url.Parse(ca.config.HubProxy)

			if err != nil {
				log.Errorf("Failed to parse 'hub_proxy' URL")
			} else {
				if ca.config.HubProxyUser != "" {
					u.User = url.UserPassword(ca.config.HubProxyUser, ca.config.HubProxyPassword)
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
	if ca.config.HubURL == "" {
		return fmt.Errorf("please set the hub_url config param")
	}

	ca.initHubHTTPClient()
	req, err := http.NewRequest("HEAD", ca.config.HubURL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", ca.userAgent())
	if ca.config.HubUser != "" {
		req.SetBasicAuth(ca.config.HubUser, ca.config.HubPassword)
	}

	resp, err := ca.hubHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to connect. %s. If you have a proxy or firewall, it may be blocking the connection", err.Error())
	}

	if resp.StatusCode == 401 && ca.config.HubUser == "" {
		return fmt.Errorf("unable to authorise without credentials. Please set hub_user & hub_password in the config")
	} else if resp.StatusCode == 401 && ca.config.HubUser != "" {
		return fmt.Errorf("unable to authorise with the provided credentials. Please correct the hub_user & hub_password in the config")
	} else if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("got bad response status: %d, %s. If you have a proxy or firewall it may be blocking the connection", resp.StatusCode, resp.Status)
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

	if ca.config.HubGzip {
		var buffer bytes.Buffer
		zw := gzip.NewWriter(&buffer)
		zw.Write(b)
		zw.Close()
		req, err = http.NewRequest("POST", ca.config.HubURL, &buffer)
		req.Header.Set("Content-Encoding", "gzip")
	} else {
		req, err = http.NewRequest("POST", ca.config.HubURL, bytes.NewBuffer(b))
	}
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", ca.userAgent())

	if ca.config.HubUser != "" {
		req.SetBasicAuth(ca.config.HubUser, ca.config.HubPassword)
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

func (ca *Cagent) Run(outputFile *os.File, interrupt chan struct{}) {
	for {
		err := ca.RunOnce(outputFile)
		if err != nil {
			log.Error(err)
		}

		select {
		case <-interrupt:
			return
		case <-time.After(secToDuration(ca.config.Interval)):
			continue
		}
	}
}
