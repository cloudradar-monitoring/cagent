// +build linux, darwin, !windows

package top

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

func (t *Top) GetProcesses(interval time.Duration) ([]*ProcessInfoSnapshot, error) {
	results := make(chan *ProcessInfoSnapshot)

	// Get all currently active processes
	ps, err := process.Processes()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get processes")
	}

	var wg sync.WaitGroup
	wg.Add(len(ps))
	skipped := uint64(0)
	// Fetch load percentage for every process
	for _, p := range ps {
		// Run in background because the call to Percent blocks for the duration
		go func(p *process.Process) {
			defer wg.Done()

			load, err := p.Percent(interval)
			if err != nil {
				// If we log the error in this place, we get _a lot of_ messages
				atomic.AddUint64(&skipped, 1)
				return
			}

			name, err := p.Name()
			if err != nil {
				// If we log the error in this place, we get _a lot of_ messages
				atomic.AddUint64(&skipped, 1)
				return
			}

			cmd, err := p.Cmdline()
			if err != nil {
				cmd = ""
			}

			parentPID, err := p.Ppid()
			if err != nil {
				parentPID = 0
			}

			// Report result back to main goroutine
			results <- &ProcessInfoSnapshot{
				Name:      name,
				PID:       uint32(p.Pid),
				ParentPID: uint32(parentPID),
				Command:   cmd,
				Load:      load / float64(t.logicalCPUCount),
			}
		}(p)
	}

	// Close channel after processing
	go func() {
		wg.Wait()
		close(results)
	}()

	result := make([]*ProcessInfoSnapshot, 0, len(ps))
	var pi *ProcessInfoSnapshot
	// Read loads collected in the background
	for pi = range results {
		result = append(result, pi)
	}

	return result, nil
}
