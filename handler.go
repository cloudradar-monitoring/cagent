package cagent

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func (ca *Cagent) initHubHttpClient() {
	if ca.hubHttpClient == nil {
		tr := *(http.DefaultTransport.(*http.Transport))
		if ca.rootCAs != nil {
			tr.TLSClientConfig = &tls.Config{RootCAs: ca.rootCAs}
		}
		if ca.HubProxy != "" {
			if !strings.HasPrefix(ca.HubProxy, "http://") {
				ca.HubProxy = "http://" + ca.HubProxy
			}

			u, err := url.Parse(ca.HubProxy)

			if err != nil {
				log.Errorf("Failed to parse 'hub_proxy' URL")
			} else {
				if ca.HubProxyUser != "" {
					u.User = url.UserPassword(ca.HubProxyUser, ca.HubProxyPassword)
				}
				tr.Proxy = func(_ *http.Request) (*url.URL, error) {
					return u, nil
				}
			}
		}

		ca.hubHttpClient = &http.Client{
			Timeout:   time.Second * 30,
			Transport: &tr,
		}
	}
}

func (ca *Cagent) TestHub() error {
	if ca.HubURL == "" {
		return fmt.Errorf("please set the hub_url config param")
	}

	ca.initHubHttpClient()
	req, err := http.NewRequest("HEAD", ca.HubURL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", ca.userAgent())
	if ca.HubUser != "" {
		req.SetBasicAuth(ca.HubUser, ca.HubPassword)
	}
	resp, err := ca.hubHttpClient.Do(req)

	if err != nil {
		return fmt.Errorf("unable to connect. %s. If you have a proxy or firewall, it may be blocking the connection", err.Error())
	}

	if resp.StatusCode == 401 && ca.HubUser == "" {
		return fmt.Errorf("unable to authorise without credentials. Please set hub_user & hub_password in the config")
	} else if resp.StatusCode == 401 && ca.HubUser != "" {
		return fmt.Errorf("unable to authorise with the provided credentials. Please correct the hub_user & hub_password in the config")
	} else if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("got bad response status: %d, %s. If you have a proxy or firewall it may be blocking the connection", resp.StatusCode, resp.Status)
	}

	return nil
}

func (ca *Cagent) PostResultsToHub(result Result) error {
	ca.initHubHttpClient()

	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	var req *http.Request

	if ca.HubGzip {
		var buffer bytes.Buffer
		zw := gzip.NewWriter(&buffer)
		zw.Write(b)
		zw.Close()
		req, err = http.NewRequest("POST", ca.HubURL, &buffer)
		req.Header.Set("Content-Encoding", "gzip")
	} else {
		req, err = http.NewRequest("POST", ca.HubURL, bytes.NewBuffer(b))
	}
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", ca.userAgent())

	if ca.HubUser != "" {
		req.SetBasicAuth(ca.HubUser, ca.HubPassword)
	}

	resp, err := ca.hubHttpClient.Do(req)

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

func (ca *Cagent) Run(outputFile *os.File, interrupt chan struct{}, once bool) {

	var jsonEncoder *json.Encoder
	if ca.PidFile != "" && !once && runtime.GOOS != "windows" {
		err := ioutil.WriteFile(ca.PidFile, []byte(strconv.Itoa(os.Getpid())), 0664)

		if err != nil {
			log.Errorf("Failed to write pid file at: %s", ca.PidFile)
		}
	}

	if outputFile != nil {
		jsonEncoder = json.NewEncoder(outputFile)
	}

	var cpu *CPUWatcher

	if len(ca.CPUUtilTypes) > 0 && len(ca.CPUUtilDataGather) > 0 || len(ca.CPULoadDataGather) > 0 {
		// optimization to prevent CPU watcher to run in case CPU util metrics not are not needed
		cpu = ca.CPUWatcher()
		err := cpu.Once()
		if err != nil {
			log.Error("[CPU] Failed to read utilisation metrics: " + err.Error())
		}
		if !once {
			go cpu.Run()
		}
	}

	fs := ca.FSWatcher()
	net := ca.NetWatcher()
	wuw := ca.WindowsUpdatesWatcher()

	for {
		results := Result{Timestamp: time.Now().Unix(), Measurements: make(MeasurementsMap)}
		errs := []string{}

		if cpu != nil {
			cpum, err := cpu.Results()

			log.Debugf("[CPU] got %d metrics", len(cpum))

			if err != nil {
				// no need to log because already done inside cpu.Results()
				errs = append(errs, err.Error())
			}

			results.Measurements = results.Measurements.AddWithPrefix("cpu.", cpum)
		}

		info, err := ca.HostInfoResults()
		if err != nil {
			// no need to log because already done inside HostInfoResults()
			errs = append(errs, err.Error())
		}

		results.Measurements = results.Measurements.AddWithPrefix("system.", info)

		ipResults, err := IPAddresses()
		if err != nil {
			errs = append(errs, err.Error())
		}

		results.Measurements = results.Measurements.AddWithPrefix("system.", ipResults)

		fsResults, err := fs.Results()
		if err != nil {
			// no need to log because already done inside fs.Results()
			errs = append(errs, err.Error())
		}

		results.Measurements = results.Measurements.AddWithPrefix("fs.", fsResults)

		netResults, err := net.Results()
		if err != nil {
			// no need to log because already done inside net.Results()
			errs = append(errs, err.Error())
		}

		results.Measurements = results.Measurements.AddWithPrefix("net.", netResults)

		proc, err := ca.ProcessesResult()
		if err != nil {
			// no need to log because already done inside ProcessesResult()
			errs = append(errs, err.Error())
		}

		results.Measurements = results.Measurements.AddWithPrefix("proc.", proc)

		mem, err := ca.MemResults()
		if err != nil {
			// no need to log because already done inside MemResults()
			errs = append(errs, err.Error())
		}

		results.Measurements = results.Measurements.AddWithPrefix("mem.", mem)

		swap, err := ca.SwapResults()
		if err != nil {
			// no need to log because already done inside MemResults()
			errs = append(errs, err.Error())
		}

		results.Measurements = results.Measurements.AddWithPrefix("swap.", swap)

		if runtime.GOOS == "linux" {
			raid, err := ca.RaidState()
			if err != nil {
				// no need to log because already done inside RaidState()
				errs = append(errs, err.Error())
			}

			results.Measurements = results.Measurements.AddWithPrefix("raid.", raid)
		}

		if runtime.GOOS == "windows" {
			wu, err := wuw.WindowsUpdates()

			results.Measurements = results.Measurements.AddWithPrefix("windows_update.", wu)

			if err != nil {
				// no need to log because already done inside MemResults()
				errs = append(errs, err.Error())
			}
		}

		if len(errs) == 0 {
			results.Measurements["cagent.success"] = 1
		} else {
			results.Message = strings.Join(errs, "; ")
			results.Measurements["cagent.success"] = 0
		}

		if outputFile != nil {
			err = jsonEncoder.Encode(results)
			if err != nil {
				log.Errorf("Results json encode error: %s", err.Error())
			}
		} else {
			err = ca.PostResultsToHub(results)
			if err != nil {
				log.Errorf("POST to hub error: %s", err.Error())
			}
		}

		if once {
			break
		}

		select {
		case <-interrupt:
			return
		case <-time.After(secToDuration(ca.Interval)):
			continue
		}
	}
}
