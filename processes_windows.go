// +build windows

package cagent

import (
	"context"
	"errors"
	"time"

	"github.com/StackExchange/wmi"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/docker"
)

const processListTimeout = time.Second * 10

type Win32_Process struct {
	Name            *string
	CommandLine     *string
	ProcessID       *uint32
	ParentProcessId *uint32
	ExecutionState  *uint16
}

func WMIQueryWithContext(ctx context.Context, query string, dst interface{}, connectServerArgs ...interface{}) error {
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

func processes(_ *docker.Watcher) ([]ProcStat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), processListTimeout)
	defer cancel()

	wmiProcs := []Win32_Process{}

	err := WMIQueryWithContext(ctx, `SELECT Name, CommandLine, ProcessID, ParentProcessId FROM Win32_Process`, &wmiProcs)
	if err != nil {
		if err == context.DeadlineExceeded {
			return nil, TimeoutError{"WMI query", processListTimeout}
		}
		return nil, errors.New("WMI query error: " + err.Error())
	}

	var procs []ProcStat
	for _, proc := range wmiProcs {
		if proc.ProcessID == nil {
			continue
		}

		ps := ProcStat{
			PID:   int(*proc.ProcessID),
			State: "running",
		}

		if proc.Name != nil {
			ps.Name = *proc.Name
		}

		if proc.ParentProcessId != nil {
			ps.ParentPID = int(*proc.ParentProcessId)
		}

		if proc.CommandLine != nil {
			ps.Cmdline = *proc.CommandLine
		}

		procs = append(procs, ps)
	}

	return procs, nil
}
