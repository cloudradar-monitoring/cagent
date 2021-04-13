package csender

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent"
	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/proxydetect"
)

func (cs *Csender) httpClient() *http.Client {
	tr := *(http.DefaultTransport.(*http.Transport))
	rootCAs, err := common.CustomRootCertPool()
	if err != nil {
		if err != common.ErrorCustomRootCertPoolNotImplementedForOS {
			fmt.Fprintln(os.Stderr, "failed to add root certs: "+err.Error())
		}
	} else if rootCAs != nil {
		tr.TLSClientConfig = &tls.Config{
			RootCAs: rootCAs,
		}
	}

	tr.Proxy = proxydetect.GetProxyForRequest
	proxydetect.UserAgent = cs.userAgent()

	return &http.Client{
		Timeout:   HubTimeout,
		Transport: &tr,
	}
}

// GracefulSend sends to hub with retry logic
func (cs *Csender) GracefulSend() error {

	retries := 0
	var retryIn time.Duration

	for {
		err := cs.Send()
		if err == nil {
			return nil
		}

		if err == cagent.ErrHubTooManyRequests {
			// for error code 429, wait 10 seconds and try again
			retryIn = 10 * time.Second
			log.Infof("csender: HTTP 429, too many requests, retrying in %v", retryIn)
		} else if err == cagent.ErrHubServerError {
			// for error codes 5xx, wait for 1 seconds and try again, increase by 1 second each retry
			retries++
			retryIn = time.Duration(retries) * time.Second

			if retries > cs.RetryLimit {
				log.Errorf("csender: hub connection error, giving up")
				return nil
			}
			log.Infof("csender: hub connection error %d/%d, retrying in %v", retries, cs.RetryLimit, retryIn)
		} else {
			return err
		}

		time.Sleep(retryIn)
	}
}

// Send is used by csender
func (cs *Csender) Send() error {
	client := cs.httpClient()

	if _, err := url.Parse(cs.HubURL); err != nil {
		return fmt.Errorf("incorrect URL provided with -u (hub URL): %s", err.Error())
	}

	b, err := json.Marshal(cs.result)
	if err != nil {
		return err
	}

	var req *http.Request

	if cs.HubGzip {
		var buffer bytes.Buffer
		zw := gzip.NewWriter(&buffer)
		_, _ = zw.Write(b)
		_ = zw.Close()
		req, err = http.NewRequest("POST", cs.HubURL, &buffer)
		if err != nil {
			return fmt.Errorf("failed to create HTTPS request: %s", err.Error())
		}

		req.Header.Set("Content-Encoding", "gzip")
	} else {
		req, err = http.NewRequest("POST", cs.HubURL, bytes.NewBuffer(b))
	}

	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", cs.userAgent())
	req.Header.Add("X-CustomCheck-Token", cs.HubToken)

	resp, err := client.Do(req)
	if err != nil {
		return clientError(resp, err)
	}

	defer resp.Body.Close()

	if resp != nil {
		if resp.StatusCode == http.StatusTooManyRequests {
			return cagent.ErrHubTooManyRequests
		}
		if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
			return cagent.ErrHubServerError
		}
	}

	if err := clientError(resp, err); err != nil {
		return err
	}

	return nil
}

func clientError(resp *http.Response, err error) error {
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
		return errors.Errorf("unable to authorize with provided token (HTTP %d). %s", resp.StatusCode, responseBody)
	} else if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return errors.Errorf("got unexpected response from the server (HTTP %d). %s", resp.StatusCode, responseBody)
	}
	return nil
}
