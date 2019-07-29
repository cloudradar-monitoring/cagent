// +build windows

package services

import (
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/cloudradar-monitoring/cagent/pkg/winapi"
)

type serviceInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	StartMode   string `json:"start"`
	AutoStart   bool   `json:"auto_start"`
	State       string `json:"state"`
	Manager     string `json:"manager"`
}

var serviceStartTypeToStringMap = map[uint32]string{
	windows.SERVICE_BOOT_START:   "boot",
	windows.SERVICE_SYSTEM_START: "system",
	windows.SERVICE_AUTO_START:   "auto",
	windows.SERVICE_DEMAND_START: "manual",
	windows.SERVICE_DISABLED:     "disabled",
}

var serviceStateToStringMap = map[uint32]string{
	windows.SERVICE_STOPPED:          "stopped",
	windows.SERVICE_START_PENDING:    "start pending",
	windows.SERVICE_STOP_PENDING:     "stop pending",
	windows.SERVICE_RUNNING:          "running",
	windows.SERVICE_CONTINUE_PENDING: "continue pending",
	windows.SERVICE_PAUSE_PENDING:    "pause pending",
	windows.SERVICE_PAUSED:           "paused",
}

func ListServices(autoStartOnly bool) (map[string]interface{}, error) {
	svcManager, err := mgr.Connect()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := svcManager.Disconnect()
		if err != nil {
			log.Debugf("could not disconnect from Windows serviceInfo Manager: %s", err)
		}
	}()

	names, err := svcManager.ListServices()
	if err != nil {
		return nil, err
	}

	servicesList := make([]*serviceInfo, 0)
	for _, serviceName := range names {
		serviceInfo, err := tryGetServiceInfo(svcManager, serviceName, autoStartOnly)
		if err != nil {
			log.Debugf("could not get information about service %s: %s", serviceName, err.Error())
			continue
		}

		if serviceInfo != nil {
			servicesList = append(servicesList, serviceInfo)
		}
	}

	return map[string]interface{}{"list": servicesList}, nil
}

func tryGetServiceInfo(svcManager *mgr.Mgr, serviceName string, autoStartOnly bool) (*serviceInfo, error) {
	s, err := openService(svcManager, serviceName)
	if err != nil {
		return nil, errors.Wrap(err, "could not open service")
	}
	defer func() {
		err := s.Close()
		if err != nil {
			log.Debugf("could not close %s service handle: %s", serviceName, err.Error())
		}
	}()

	cfg, err := s.Config()
	if err != nil {
		return nil, errors.Wrap(err, "could not get config for service")
	}

	isAutoStart := cfg.StartType == mgr.StartAutomatic
	if autoStartOnly && !isAutoStart {
		return nil, nil
	}

	description := cfg.DisplayName
	if cfg.Description != "" {
		description += " " + cfg.Description
	}

	serviceState, err := queryState(s.Handle)
	if err != nil {
		return nil, errors.Wrap(err, "could not query state")
	}

	isDelayedAutoStart, err := winapi.GetIsServiceHaveDelayedAutoStartFlag(s.Handle)
	if err != nil {
		return nil, errors.Wrap(err, "could not query if service has DelayedAutoStart option")
	}

	result := serviceInfo{
		Name:        serviceName,
		Description: description,
		State:       formatState(serviceState),
		StartMode:   formatStartMode(cfg.StartType, isDelayedAutoStart),
		AutoStart:   isAutoStart,
		Manager:     "windows",
	}

	return &result, nil
}

// openService is similar to mgr.OpenService(), but asks only for read-only access to services
func openService(svcManager *mgr.Mgr, serviceName string) (*mgr.Service, error) {
	serviceNameAddr, err := syscall.UTF16PtrFromString(serviceName)
	if err != nil {
		return nil, err
	}
	h, err := windows.OpenService(
		svcManager.Handle,
		serviceNameAddr,
		windows.SERVICE_QUERY_CONFIG|windows.SERVICE_QUERY_STATUS|windows.SERVICE_ENUMERATE_DEPENDENTS,
	)
	if err != nil {
		return nil, err
	}
	return &mgr.Service{Name: serviceName, Handle: h}, nil
}

func formatStartMode(startType uint32, isDelayedAutoStart bool) string {
	result := ""
	if str, ok := serviceStartTypeToStringMap[startType]; ok {
		result = str
	} else {
		result = "unknown"
	}

	if isDelayedAutoStart {
		result += "_delayed"
	}
	return result
}

func formatState(serviceState uint32) string {
	result := ""
	if str, ok := serviceStateToStringMap[serviceState]; ok {
		result = str
	} else {
		result = "unknown"
	}
	return result
}

func queryState(handle windows.Handle) (uint32, error) {
	var p *windows.SERVICE_STATUS_PROCESS
	var bytesNeeded uint32
	var buf []byte

	if err := windows.QueryServiceStatusEx(handle, windows.SC_STATUS_PROCESS_INFO, nil, 0, &bytesNeeded); err != windows.ERROR_INSUFFICIENT_BUFFER {
		return 0, err
	}

	buf = make([]byte, bytesNeeded)
	p = (*windows.SERVICE_STATUS_PROCESS)(unsafe.Pointer(&buf[0]))
	if err := windows.QueryServiceStatusEx(handle, windows.SC_STATUS_PROCESS_INFO, &buf[0], uint32(len(buf)), &bytesNeeded); err != nil {
		return 0, err
	}

	return p.CurrentState, nil
}
