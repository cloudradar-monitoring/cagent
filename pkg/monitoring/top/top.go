package top

import (
	"log"
	"sync"
	"time"

	"github.com/shirou/gopsutil/process"
)

type ProcessInfo struct {
	Command string
	Load1   float64
	Load5   float64
	Load15  float64
}

type Top struct {
	pList    map[int32]*ProcessInfo
	pListMtx sync.RWMutex
	stop     bool
}

func New() *Top {
	t := &Top{
		pList: make(map[int32]*ProcessInfo),
	}

	return t
}

func (t *Top) Run() {
	var tStart, tEnd time.Time
	for {
		// Check if stop was requested
		if t.stop == true {
			return
		}
		tStart = time.Now()
		processes, err := process.Processes()
		if err != nil {
			log.Printf("Error loading processes: %s", err)
			time.Sleep(time.Second * 1)
			continue
		}

		t.pListMtx.Lock()
		for _, p := range processes {
			// Add process to process list if not already there
			_, ok := t.pList[p.Pid]
			if !ok {
				t.pList[p.Pid] = &ProcessInfo{}
			}

			cpuPercent, err := p.CPUPercent()
			if err != nil {
				log.Printf("Error getting CPU percentage for process: %s", err)
				//TODO: what to do?
				continue
			}
			t.pList[p.Pid].Load1 = cpuPercent
		}
		t.pListMtx.Unlock()
		tEnd = time.Now()

		// Try to get a sample every second
		tWait := tEnd.Sub(tStart)
		if tWait.Nanoseconds() < 1000000000 {
			time.Sleep(time.Nanosecond * time.Duration(tWait.Nanoseconds()))
		} else {
			log.Printf("WARN: Took more than 1s to get cpu load of processes")
		}
	}
}

func (t *Top) Stop() {
	t.stop = true
}

func (t *Top) HighestLoad() {
	var pi ProcessInfo
	var pid int32

	t.pListMtx.RLock()
	for k, v := range t.pList {
		if v.Load1 > pi.Load1 {
			pi = *v
			pid = k
		}
	}
	t.pListMtx.RUnlock()

	log.Printf("Load %d: %f.2", pid, pi.Load1)
}
