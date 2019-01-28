// +build windows

package services

import (
	"context"
	"strings"
	"time"

	"github.com/StackExchange/wmi"
)

const serviceListTimeout = time.Second * 5

// ErrorNotImplementedForOS exists here just for cross-platform building (cause it presented in services.go)
var ErrorNotImplementedForOS error

type Win32_Service struct {
	Name             string
	DisplayName      string
	Description      string
	StartMode        string
	State            string
	Status           string
	DelayedAutoStart bool
}

// todo: move to the separate package when we will also move processes.go to the separate package
func wmiQueryWithContext(ctx context.Context, query string, dst interface{}, connectServerArgs ...interface{}) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- wmi.Query(query, dst, connectServerArgs...)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

// ListServices parse Windows Service Manager
func ListServices(autostartOnly bool) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), serviceListTimeout)
	defer cancel()

	var wmiServices []Win32_Service
	err := wmiQueryWithContext(ctx, "Select Name,DisplayName,Description,StartMode,State,State,DelayedAutoStart from Win32_Service", &wmiServices)
	if err != nil {
		return nil, err
	}

	var servicesList []map[string]interface{}
	for _, wmiService := range wmiServices {
		var autoStart bool

		if strings.HasPrefix(strings.ToLower(wmiService.StartMode), "auto") {
			autoStart = true
		}

		if autostartOnly && !autoStart {
			continue
		}

		if wmiService.DelayedAutoStart {
			wmiService.StartMode = wmiService.StartMode + "_delayed"
		}

		servicesList = append(servicesList, map[string]interface{}{
			"name":        wmiService.Name,
			"description": wmiService.DisplayName + " " + wmiService.Description,
			"start":       strings.ToLower(wmiService.StartMode),
			"auto_start":  autoStart,
			"state":       strings.ToLower(wmiService.State),
			"status":      strings.ToLower(wmiService.Status),
			"manager":     "windows",
		})
	}

	return map[string]interface{}{"list": servicesList}, nil
}
