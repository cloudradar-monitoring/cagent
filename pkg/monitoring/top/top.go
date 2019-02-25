package top

import (
	"container/ring"
	"math"
	"runtime"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Process is used to store aggregated load data about an OS process
type Process struct {
	PID        uint32
	Load       float64
	Load1      *ring.Ring
	Load5      *ring.Ring
	Load15     *ring.Ring
	Command    string
	Identifier string
}

// ProcessInfoSlice is used for sorting
type ProcessInfoSlice []*ProcessInfo

func (s ProcessInfoSlice) Len() int {
	return len(s)
}
func (s ProcessInfoSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ProcessInfoSlice) Less(i, j int) bool {
	return s[i].Load1 < s[j].Load1
}

// ProcessInfo is used to store a snapshot of load data about an OS process
type ProcessInfo struct {
	Name    string  `json:"name"`
	PID     uint32  `json:"pid"`
	Command string  `json:"command,omitempty"`
	Load    float64 `json:"-"`
	Load1   float64 `json:"load1"`
	Load5   float64 `json:"load5"`
	Load15  float64 `json:"load15"`
}

// Top holds a map with information about process loads
type Top struct {
	pList           map[string]*Process
	pListMtx        sync.RWMutex
	LoadTotal1      float64
	LoadTotal5      float64
	LoadTotal15     float64
	isRunning       bool
	isRunningMtx    sync.Mutex
	stop            bool
	logicalCPUCount uint8
}

// New returns a new instance of Top struct
func New() *Top {
	t := &Top{
		pList:           make(map[string]*Process),
		isRunning:       false,
		stop:            false,
		logicalCPUCount: uint8(runtime.NumCPU()),
	}

	return t
}

func (t *Top) startMeasureProcessLoad(interval time.Duration) {
	var len1 = 60 / int(interval.Seconds())
	var len5 = 5 * 60 / int(interval.Seconds())
	var len15 = 15 * 60 / int(interval.Seconds())

	for {
		// Check if stop was requested
		if t.stop {
			t.isRunning = false
			return
		}

		processes, err := t.GetProcesses(interval)
		if err != nil {
			log.Errorf("Failed to get process list: %s", err)
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
					Load1:      ring.New(len1),
					Load5:      ring.New(len5),
					Load15:     ring.New(len15),
				}
				t.pList[p.Name] = pr
			} else {
				pr = t.pList[p.Name]
			}

			// Add load value to rings for later calculation of load load5 and load15
			pr.Load = p.Load
			pr.Load1.Value = p.Load
			pr.Load1 = pr.Load1.Next()
			pr.Load5.Value = p.Load
			pr.Load5 = pr.Load5.Next()
			pr.Load15.Value = p.Load
			pr.Load15 = pr.Load15.Next()
		}

		t.pListMtx.Unlock()
	}
}

// Run starts measuring process load on the system
func (t *Top) Run() {
	t.isRunningMtx.Lock()
	if !t.isRunning {
		t.isRunning = true
		// Start collecting process info every sec
		go t.startMeasureProcessLoad(time.Second * 5)
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
	t.pListMtx.RUnlock()

	result := make([]*ProcessInfo, 0, len(pl))
	for _, p := range pl {
		pi := &ProcessInfo{
			Name:    p.Identifier,
			Command: p.Command,
			PID:     p.PID,
			Load:    p.Load,
			Load1:   Avg1(p),
			Load5:   Avg5(p),
			Load15:  Avg15(p),
		}
		result = append(result, pi)
	}

	sort.Sort(sort.Reverse(ProcessInfoSlice(result)))

	if n <= len(result) {
		return result[:n]
	}

	return result
}

// Avg1 calculates 1min average
func Avg1(pi *Process) float64 {
	total := float64(0)
	count := 0
	pi.Load1.Do(func(p interface{}) {
		if p == nil {
			return
		}
		count++
		total += p.(float64)
	})

	return math.Round(total/float64(count)*100) / 100
}

// Avg5 calculates 5min average
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
	return math.Round(total/float64(count)*100) / 100
}

// Avg15 calculates 15min average
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
	return math.Round(total/float64(count)*100) / 100
}
