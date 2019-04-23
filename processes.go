package cagent

import (
	"runtime"

	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/types"
)

type ProcStat struct {
	PID                    int     `json:"pid"`
	ParentPID              int     `json:"parent_pid"`
	Name                   string  `json:"name"`
	Cmdline                string  `json:"cmdline"`
	State                  string  `json:"state"`
	Container              string  `json:"container,omitempty"`
	CPUAverageUsagePercent float32 `json:"cpu_avg_usage_percent,omitempty"`
	RSS                    uint64  `json:"rss"` // Resident Set Size
	VMS                    uint64  `json:"vms"` // Virtual Memory Size
	MemoryUsagePercent     float32 `json:"memory_usage_percent"`
}

// Gets possible process states based on the OS
func getPossibleProcStates() []string {
	fields := []string{
		"blocked",
		"zombie",
		"stopped",
		"running",
		"sleeping",
	}

	switch runtime.GOOS {
	case "windows":
		fields = []string{"running"}
	case "freebsd":
		fields = append(fields, "idle", "wait")
	case "darwin":
		fields = append(fields, "idle")
	case "openbsd":
		fields = append(fields, "idle")
	case "linux":
		fields = append(fields, "dead", "paging", "idle")
	}
	return fields
}

func (ca *Cagent) ProcessesResult(memStat *mem.VirtualMemoryStat) (m types.MeasurementsMap, procs []ProcStat, err error) {
	states := getPossibleProcStates()

	var systemMemorySize uint64
	if memStat == nil {
		log.Warn("[PROC] system memory information in unavailable. Some process stats will not calculated...")
	} else {
		systemMemorySize = memStat.Total
	}
	procs, err = processes(ca.dockerWatcher, systemMemorySize)
	if err != nil {
		log.Error("[PROC] error: ", err.Error())
		return nil, nil, err
	}
	log.Debugf("[PROC] results: %d", len(procs))

	m = types.MeasurementsMap{"list": procs, "possible_states": states}

	return
}
