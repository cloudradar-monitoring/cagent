// +build windows

package cagent

import (
	"context"
	"errors"
	"time"

	"github.com/StackExchange/wmi"
)

type Win32_Process struct {
	Name           string
	CommandLine    *string
	ProcessID      uint32
	ExecutionState *uint16
}

func WMIQueryWithContext(ctx context.Context, query string, dst interface{}, connectServerArgs ...interface{}) error {
	if _, ok := ctx.Deadline(); !ok {
		ctxTimeout, cancel := context.WithTimeout(ctx, Timeout)
		defer cancel()
		ctx = ctxTimeout
	}

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

func processes(fields map[string][]ProcStat) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	procs := []Win32_Process{}

	err := WMIQueryWithContext(ctx, `SELECT Name, CommandLine, ProcessID, ExecutionState FROM Win32_Process`, &procs)
	if err != nil {
		return errors.New("WMI query error: " + err.Error())
	}

	for _, proc := range procs {
		fields["running"] = append(fields["running"], ProcStat{PID: int(proc.ProcessID), Name: proc.Name, Cmdline: *proc.CommandLine})
		fields["total"] = append(fields["total"], ProcStat{PID: int(proc.ProcessID), Name: proc.Name, Cmdline: *proc.CommandLine})
	}

	return nil
}
