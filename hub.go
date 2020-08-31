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
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/proxydetect"
)

func (ca *Cagent) initHubClientOnce() {
	ca.hubClientOnce.Do(func() {
		// copy the default transport struct
		transport := *(http.DefaultTransport.(*http.Transport))
		transport.ResponseHeaderTimeout = 15 * time.Second

		rootCAs, err := common.CustomRootCertPool()
		if err != nil {
			if err != common.ErrorCustomRootCertPoolNotImplementedForOS {
				logrus.Errorf("failed to add root certs: %s", err.Error())
			}
		} else if rootCAs != nil {
			transport.TLSClientConfig = &tls.Config{
				RootCAs: rootCAs,
			}
		}

		transport.Proxy = proxydetect.GetProxyForRequest
		proxydetect.UserAgent = ca.userAgent()

		if len(ca.Config.HubProxy) > 0 {
			// in case we have proxy set in the config
			// it will override the proxy from the system
			if !strings.HasPrefix(ca.Config.HubProxy, "http://") {
				ca.Config.HubProxy = "http://" + ca.Config.HubProxy
			}
			proxyURL, err := url.Parse(ca.Config.HubProxy)
			if err != nil {
				logrus.WithFields(logrus.Fields{
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
			Transport: &transport,
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
		if req != nil {
			req.Header.Set("Content-Encoding", "gzip")
		}
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

	if resp != nil {
		if resp.StatusCode == http.StatusTooManyRequests {
			return ErrHubTooManyRequests
		}
		if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
			return ErrHubServerError
		}
	}
	if err = ca.checkClientError(resp, err, "hub_user", "hub_password"); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
