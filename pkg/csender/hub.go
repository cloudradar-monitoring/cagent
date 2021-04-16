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
		statusCode, err := cs.Send()
		if err == nil {
			return nil
		}

		if cs.Verbose {
			if statusCode >= 200 && statusCode <= 299 {
				fmt.Fprintln(os.Stdout, "HTTP CODE", statusCode)
			} else {
				fmt.Fprintln(os.Stderr, "HTTP CODE", statusCode)
			}
		}

		if err == cagent.ErrHubTooManyRequests {
			// for error code 429, wait 10 seconds and try again
			retryIn = 10 * time.Second
			log.Infof("csender: HTTP 429, too many requests, retrying in %v", retryIn)
			if cs.Verbose {
				fmt.Fprintf(os.Stdout, "got HTTP %d from %s, retrying in %v\n", statusCode, cs.HubURL, retryIn)
			}
		} else if err == cagent.ErrHubServerError || errors.Is(err, context.DeadlineExceeded) {
			// for error codes 5xx, wait for 1 seconds and try again, increase by 1 second each retry
			retries++
			retryIn = time.Duration(retries) * time.Second

			if retries > cs.RetryLimit {
				log.Errorf("csender: hub connection error, giving up after %d tries", retries)
				if cs.Verbose {
					fmt.Fprintf(os.Stderr, "giving up after %d tries\n", retries)
				}
				return nil
			}
			log.Infof("csender: hub connection error '%s', %d/%d, retrying in %v", err, retries, cs.RetryLimit, retryIn)
			if cs.Verbose {
				fmt.Fprintf(os.Stdout, "got HTTP %d from %s, retrying in %v\n", statusCode, cs.HubURL, retryIn)
			}
		} else {
			return err
		}

		time.Sleep(retryIn)
	}
}

// Send is used by csender. returns status code, error
func (cs *Csender) Send() (int, error) {
	client := cs.httpClient()

	if _, err := url.Parse(cs.HubURL); err != nil {
		return 0, fmt.Errorf("incorrect URL provided with -u (hub URL): %s", err.Error())
	}

	b, err := json.Marshal(cs.result)
	if err != nil {
		return 0, err
	}

	var req *http.Request

	if cs.HubGzip {
		var buffer bytes.Buffer
		zw := gzip.NewWriter(&buffer)
		_, _ = zw.Write(b)
		_ = zw.Close()
		req, err = http.NewRequest("POST", cs.HubURL, &buffer)
		if err != nil {
			return 0, fmt.Errorf("failed to create HTTPS request: %s", err.Error())
		}

		req.Header.Set("Content-Encoding", "gzip")
	} else {
		req, err = http.NewRequest("POST", cs.HubURL, bytes.NewBuffer(b))
	}

	if err != nil {
		return 0, err
	}

	req.Header.Add("User-Agent", cs.userAgent())
	req.Header.Add("X-CustomCheck-Token", cs.HubToken)

	resp, err := client.Do(req)
	if err != nil {
		return 0, clientError(resp, err)
	}

	defer resp.Body.Close()

	if resp != nil {
		if resp.StatusCode == http.StatusTooManyRequests {
			return resp.StatusCode, cagent.ErrHubTooManyRequests
		}
		if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
			return resp.StatusCode, cagent.ErrHubServerError
		}
	}

	if err := clientError(resp, err); err != nil {
		return resp.StatusCode, err
	}

	return resp.StatusCode, nil
}

func clientError(resp *http.Response, err error) error {
	if err != nil {
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
