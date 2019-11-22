package csender

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"

	"github.com/pkg/errors"
)

func (cs *Csender) httpClient() *http.Client {
	tr := *(http.DefaultTransport.(*http.Transport))

	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		// inject ca root certs, cause they can be missing on some systems
		rootCertsPath := "/etc/cagent/cacert.pem"
		if _, err := os.Stat(rootCertsPath); err == nil {
			certPool := x509.NewCertPool()

			b, err := ioutil.ReadFile(rootCertsPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to read cacert.pem: "+err.Error())
			} else {
				ok := certPool.AppendCertsFromPEM(b)
				if ok {
					tr.TLSClientConfig = &tls.Config{RootCAs: certPool}
				}
			}
		}
	}

	return &http.Client{
		Timeout:   HubTimeout,
		Transport: &tr,
	}
}

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
