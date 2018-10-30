// build linux, darwin, !windows

package top

import (
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
)

func (t *Top) startCollect(interval time.Duration) {
	// No implementation needed
}

func (t *Top) GetProcesses() ([]*ProcessInfo, error) {
	results := make(chan *ProcessInfo)

	// Get all currently active processes
	ps, err := process.Processes()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get processess")
	}
	// Fetch load percentage for every process
	for _, p := range ps {
		// Run in background because the call to Percent blocks for the duration
		go func(p *process.Process) {
			load, err := p.Percent(time.Second * 1)
			if err != nil {
				log.Printf("Err getting Percent: %s", err)
			}

			name, _ := p.Name()
			cmd, err := p.Cmdline()
			if err != nil {
				log.Printf("Failed to get cmdLine: %s", err)
			}

			// Report result back to main goroutine
			results <- &ProcessInfo{
				Identifier: name,
				PID:        uint32(p.Pid),
				Command:    cmd,
				Load:       load,
			}
		}(p)
	}

	lenTotal := len(ps)
	result := make([]*ProcessInfo, 0, lenTotal)
	count := 0
	var pi *ProcessInfo
	// Read loads collected in the background
	for {
		pi = <-results
		result = append(result, pi)
		count++
		if count == lenTotal {
			break
		}
	}

	return result, nil
}
