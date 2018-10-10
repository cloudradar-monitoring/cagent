package top

import (
	"bytes"
	"container/ring"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
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
	Command string
	Load1   float64
	Load5   float64
	Load15  float64
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
		var buff bytes.Buffer

		// Command to list processes
		cmdPS := exec.Command("ps", "ax", "-o", "pid,%cpu,command")
		// Command to sort processes by cpu load
		cmdSort := exec.Command("sort", "-u", "-k2")

		// List processes and sort them
		r, w := io.Pipe()
		cmdPS.Stdout = w
		cmdSort.Stdin = r
		cmdSort.Stdout = &buff

		cmdPS.Start()
		cmdSort.Start()
		cmdPS.Wait()
		w.Close()
		cmdSort.Wait()

		lines := strings.Split(buff.String(), "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}

			parts1 := strings.Split(line, "   ")

			// If load is >= 10 there are only two spaces
			if len(parts1) != 2 {
				parts1 = strings.Split(line, "  ")
			}

			// Workaround if format is a bit off for some reason.
			// Maybe could make sense to log these and investigate later
			if len(parts1) < 2 {
				continue
			}

			parts2 := strings.SplitN(parts1[1], " ", 2)

			parsedLoad, err := strconv.ParseFloat(parts2[0], 64)
			if err != nil {

			}

			parsedPID, err := strconv.ParseUint(parts1[0], 10, 32)
			if err != nil {

			}

			// Workaround if format is a bit off for some reason.
			// Maybe could make sense to log these and investigate later
			if len(parts2) < 2 {
				log.Printf("Splitting error2:")
				log.Printf("%+v", line)
				log.Printf("%+v", parts1)
				log.Printf("%+v", parts2)
				continue
			}

			var p *Process

			// Check if we already track the process
			if _, ok := t.pList[uint32(parsedPID)]; !ok {
				p = &Process{
					Command: parts2[1],
					Load5:   ring.New(5 * 60),
					Load15:  ring.New(15 * 60),
					PID:     uint32(parsedPID),
				}
				t.pList[uint32(parsedPID)] = p
			} else {
				p = t.pList[uint32(parsedPID)]
			}

			p.Load = parsedLoad
			p.Load5.Value = parsedLoad
			p.Load5 = p.Load5.Next()
			p.Load15.Value = parsedLoad
			p.Load15 = p.Load15.Next()
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
