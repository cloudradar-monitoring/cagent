package top

import (
	"container/ring"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Process is used to store aggregated load data about an OS process
type Process struct {
	PID        uint32
	Load       float64
	Load5      *ring.Ring
	Load15     *ring.Ring
	Command    string
	Identifier string
}

type ProcessSlice []*Process

func (s ProcessSlice) Len() int {
	return len(s)
}
func (s ProcessSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ProcessSlice) Less(i, j int) bool {
	return s[i].Load < s[j].Load
}

// ProcessInfo is used to store a snapshot of load data about an OS process
type ProcessInfo struct {
	Name    string  `json:"name"`
	PID     uint32  `json:"pid"`
	Command string  `json:"command"`
	Load    float64 `json:"load"`
	Load5   float64 `json:"load5"`
	Load15  float64 `json:"load15"`
}

// Top holds a map with information about process loads
type Top struct {
	pList        map[string]*Process
	pListMtx     sync.RWMutex
	LoadTotal1   float64
	LoadTotal5   float64
	LoadTotal15  float64
	isRunning    bool
	isRunningMtx sync.Mutex
	stop         bool
}

// New returns a new instance of Top struct
func New() *Top {
	t := &Top{
		pList:     make(map[string]*Process),
		isRunning: false,
		stop:      false,
	}

	return t
}

func (t *Top) startMearueProcessLoad(interval time.Duration) {
	t.startCollect(interval)
	for {
		// Check if stop was requested
		if t.stop {
			t.isRunning = false
			return
		}

		// Call to os agnostic implementation to fetch processes
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
			if _, ok := t.pList[p.Name]; !ok {
				pr = &Process{
					PID:        p.PID,
					Command:    p.Command,
					Identifier: p.Name,
					Load5:      ring.New(5 * 60),
					Load15:     ring.New(15 * 60),
				}
				t.pList[p.Name] = pr
			} else {
				pr = t.pList[p.Name]
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

// Run starts measuring process load on the system
func (t *Top) Run() {
	t.isRunningMtx.Lock()
	if !t.isRunning {
		t.isRunning = true
		// Start collecting process info every sec
		go t.startMearueProcessLoad(time.Second * 1)
	} else {
		log.Debug("Skipped starting Top because it's already running")
	}
	t.isRunningMtx.Unlock()
}

// Stop signals that load measuring should be stopped
func (t *Top) Stop() {
	log.Printf("Top Stop called")
	t.stop = true
}

func (t *Top) HighestNLoad(n int) []*ProcessInfo {
	pl := make([]*Process, 0, len(t.pList))

	t.pListMtx.RLock()
	for _, v := range t.pList {
		pl = append(pl, v)
	}
	sort.Sort(ProcessSlice(pl))
	t.pListMtx.RUnlock()

	result := make([]*ProcessInfo, 0, len(pl))
	for _, p := range pl {
		pi := &ProcessInfo{
			Name:    p.Identifier,
			Command: p.Command,
			PID:     p.PID,
			Load:    p.Load,
			Load5:   Avg5(p),
			Load15:  Avg15(p),
		}
		result = append(result, pi)
	}

	if n <= len(result) {
		return result[:n]
	}

	return result
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

// Func to calculate 5min average
func Avg5(pi *Process) float64 {
	total := float64(0)
	count := 0
	pi.Load5.Do(func(p interface{}) {
		if p == nil {
			return
		}
		count++
		total += p.(float64)
	})
	return total / float64(count)
}

// Func to calculate 15min average
func Avg15(pi *Process) float64 {
	total := float64(0)
	count := 0
	pi.Load15.Do(func(p interface{}) {
		if p == nil {
			return
		}
		count++
		total += p.(float64)
	})
	return total / float64(count)
}

func LowerThan(value, threshold float64) bool {
	return value < threshold
}
