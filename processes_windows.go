// +build windows

package cagent

import (
	"context"
	"errors"
	"time"

	"github.com/StackExchange/wmi"
)

const processListTimeout = time.Second * 10

type Win32_Process struct {
	Name            string
	CommandLine     *string
	ProcessID       uint32
	ParentProcessId uint32
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

func processes() ([]ProcStat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), processListTimeout)
	defer cancel()

	wmiProcs := []Win32_Process{}

	err := WMIQueryWithContext(ctx, `SELECT Name, CommandLine, ProcessID, ExecutionState FROM Win32_Process`, &wmiProcs)
	if err != nil {
		return nil, errors.New("WMI query error: " + err.Error())
	}

	var procs []ProcStat
	for _, proc := range wmiProcs {
		procs = append(procs,
			ProcStat{
				PID:       int(proc.ProcessID),
				ParentPID: int(proc.ParentProcessId),
				Name:      proc.Name,
				Cmdline:   *proc.CommandLine,
				State:     "running"},
		)
	}

	return procs, nil
}
