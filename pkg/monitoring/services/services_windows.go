// +build windows

package services

import (
	"context"
	"github.com/StackExchange/wmi"
	"time"
)

const serviceListTimeout = time.Second * 5

// ErrorNotImplementedForOS exists here just for cross-platform building (cause it presented in services.go)
var ErrorNotImplementedForOS error

type Win32_Service struct {
	Name        string
	DisplayName string
	StartMode   string
	State       string
	Status      string
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
func ListServices() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), serviceListTimeout)
	defer cancel()

	var wmiServices []Win32_Service
	err := wmiQueryWithContext(ctx, "Select Name,DisplayName,StartMode,State,Status from Win32_Service", &wmiServices)

	if err != nil {
		return nil, err
	}

	var servicesList []map[string]interface{}
	for _, wmiService := range wmiServices {
		servicesList = append(servicesList, map[string]interface{}{
			"name":         wmiService.Name,
			"display_name": wmiService.DisplayName,
			"start":        wmiService.StartMode,
			"state":        wmiService.State,
			"status":       wmiService.Status,
			"manager":      "windows",
		})
	}

	return map[string]interface{}{"list": servicesList}, nil
}
