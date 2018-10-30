package top

import (
	"container/ring"
	"log"
	"sync"
	"time"
)

// Process is used to store aggregated load data about an OS process
type Process struct {
	PID     uint32
	Load    float64
	Load5   *ring.Ring
	Load15  *ring.Ring
	Command string
}

// ProcessInfo is used to store a snapshot of laod data about an OS process
type ProcessInfo struct {
	Identifier string
	PID        uint32
	Command    string
	Load       float64
}

// Top holds a map with information about process loads
type Top struct {
	pList    map[string]*Process
	pListMtx sync.RWMutex
	stop     bool
}

// New returns a new instance of Top struct
func New() *Top {
	t := &Top{
		pList: make(map[string]*Process),
	}

	return t
}

// Run starts measuring process load on the system
func (t *Top) Run() {
	interval := time.Second * 1
	// Call to os agnostic implementation
	t.startCollect(interval)

	for {
		// Check if stop was requested
		if t.stop {
			return
		}

		// Call to os agnostic implementation to fetch procsses
		processes, err := t.GetProcesses()
		if err != nil {
			log.Printf("Failed to get process list: %s", err)
			time.Sleep(interval)
			continue
		}

		// Lock because the map can be accesses from other goroutines as well
		t.pListMtx.Lock()
		var pr *Process
		for _, p := range processes {
			// Check if we already track the process ad if not start tracking it
			if _, ok := t.pList[p.Identifier]; !ok {
				pr = &Process{
					PID:     p.PID,
					Command: p.Command,
					Load5:   ring.New(5 * 60),
					Load15:  ring.New(15 * 60),
				}
				t.pList[p.Identifier] = pr
			} else {
				pr = t.pList[p.Identifier]
			}

			// Add load value to rings for later calculation of load load5 and load15
			pr.Load = p.Load
			pr.Load5.Value = p.Load
			pr.Load5 = pr.Load5.Next()
			pr.Load15.Value = p.Load
			pr.Load15 = pr.Load15.Next()
		}
		t.pListMtx.Unlock()
		time.Sleep(interval)
	}
}

// Stop signals that load measuring should be stopped
func (t *Top) Stop() {
	t.stop = true
}

// HighestLoad returns information about the process causing highest CPU load
func (t *Top) HighestLoad() (string, float64, float64, float64) {
	var pi *Process
	var identifier string

	// Find process with highest load
	t.pListMtx.RLock()
	for k, v := range t.pList {
		if pi == nil {
			pi = v
		}
		if v.Load > pi.Load {
			pi = v
			identifier = k
		}
	}
	t.pListMtx.RUnlock()

	var avg5, avg15 float64
	var wg sync.WaitGroup
	wg.Add(2)

	// Func to calculate 5min average
	avg5f := func() {
		total := float64(0)
		count := 0
		pi.Load5.Do(func(p interface{}) {
			if p == nil {
				return
			}
			count++
			total += p.(float64)
		})
		avg5 = total / float64(count)
		wg.Done()
	}

	// Func to calculate 15min average
	avg15f := func() {
		total := float64(0)
		count := 0
		pi.Load15.Do(func(p interface{}) {
			if p == nil {
				return
			}
			count++
			total += p.(float64)
		})
		avg15 = total / float64(count)
		wg.Done()
	}

	// pi can be nil on first run until thigs are set up
	if pi == nil {
		return "", 0, 0, 0
	}

	// Sanity check in case the Rings aren't initialised yet
	if pi.Load5 != nil && pi.Load15 != nil {
		go avg5f()
		go avg15f()
		wg.Wait()
	}

	return identifier, pi.Load, avg5, avg15
}
