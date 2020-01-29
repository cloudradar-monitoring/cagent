package processes

import (
	"runtime"
	"sort"

	"github.com/shirou/gopsutil/mem"
	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

var log = logrus.WithField("package", "processes")

type Config struct {
	Enabled                     bool `toml:"enabled"`
	EnableKernelTaskMonitoring  bool `toml:"enable_kerneltask_monitoring" comment:"Monitor kernel tasks identified by process group 0\nIgnored on Windows."`
	MaxNumberMonitoredProcesses uint `toml:"max_number_monitored_processes" comment:"The process list is sorted by PID descending. Only the top N processes are monitored."`
}

func GetDefaultConfig() Config {
	return Config{
		Enabled:                     true,
		EnableKernelTaskMonitoring:  true,
		MaxNumberMonitoredProcesses: 500,
	}
}

type ProcStat struct {
	PID                    int     `json:"pid"`
	ParentPID              int     `json:"parent_pid"`
	ProcessGID             int     `json:"-"`
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

func GetMeasurements(memStat *mem.VirtualMemoryStat, cfg *Config) (common.MeasurementsMap, []*ProcStat, error) {
	states := getPossibleProcStates()

	var systemMemorySize uint64
	if memStat == nil {
		log.Warn("system memory information in unavailable. Some process stats will not calculated...")
	} else {
		systemMemorySize = memStat.Total
	}
	procs, err := processes(systemMemorySize)
	if err != nil {
		log.WithError(err).Error()
		return nil, nil, err
	}
	log.Debugf("results: %d", len(procs))

	var m common.MeasurementsMap
	if cfg.Enabled {
		m = common.MeasurementsMap{"list": filterProcs(procs, cfg), "possible_states": states}
	}

	return m, procs, nil
}

func filterProcs(procs []*ProcStat, cfg *Config) []*ProcStat {
	// sort by PID descending:
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].PID > procs[j].PID
	})

	result := make([]*ProcStat, 0, cfg.MaxNumberMonitoredProcesses)
	var count uint
	for _, p := range procs {
		if count == cfg.MaxNumberMonitoredProcesses {
			break
		}

		if !cfg.EnableKernelTaskMonitoring && isKernelTask(p) {
			continue
		}

		result = append(result, p)
		count++
	}
	return result
}
