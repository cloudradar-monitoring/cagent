package top

import (
	"container/ring"
	"log"
	"sync"
	"time"
)

type Process struct {
	PID     uint32
	Load    float64
	Load5   *ring.Ring
	Load15  *ring.Ring
	Command string
}

type ProcessInfo struct {
	PID     uint32
	Command string
	Load    float64
}

type Top struct {
	pList    map[uint32]*Process
	pListMtx sync.RWMutex
	stop     bool
}

func New() *Top {
	t := &Top{
		pList: make(map[uint32]*Process),
	}

	return t
}

func (t *Top) Run() {
	var tStart time.Time
	for {
		tStart = time.Now()
		// Check if stop was requested
		if t.stop == true {
			return
		}

		// Call to OS dependent implementation to fetch procsses
		processes, err := t.GetProcesses()
		if err != nil {
			log.Printf("Failed to get process list: %s", err)
			continue
		}

		var pr *Process
		for _, p := range processes {
			// Check if we already track the process
			if _, ok := t.pList[p.PID]; !ok {
				pr = &Process{
					PID:     p.PID,
					Command: p.Command,
					Load5:   ring.New(5 * 60),
					Load15:  ring.New(15 * 60),
				}
				t.pList[p.PID] = pr
			} else {
				pr = t.pList[p.PID]
			}

			pr.Load = p.Load
			pr.Load5.Value = p.Load
			pr.Load5 = pr.Load5.Next()
			pr.Load15.Value = p.Load
			pr.Load15 = pr.Load15.Next()
		}

		// Try to get a sample every second
		tRun := time.Since(tStart)
		tWait := 1000000000 - tRun.Nanoseconds()
		if tWait > 0 {
			// log.Printf("Waiting: %fms", float64(tWait)/1000000)
			time.Sleep(time.Nanosecond * time.Duration(tWait))
		} else {
			log.Printf("WARN: Took more than 1s to get cpu load of processes")
		}
	}
}

func (t *Top) Stop() {
	t.stop = true
}

func (t *Top) HighestLoad() {
	var pi *Process
	var pid uint32

	// Find process with highest load
	t.pListMtx.RLock()
	for k, v := range t.pList {
		if pi == nil {
			pi = v
		}
		if v.Load > pi.Load {
			pi = v
			pid = k
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
		return
	}

	// Sanity check in case the Rings aren't initialised yet
	if pi.Load5 != nil && pi.Load15 != nil {
		go avg5f()
		go avg15f()
		wg.Wait()
	}

	log.Printf("Load %d: %f - %f - %f", pid, pi.Load, avg5, avg15)
}
