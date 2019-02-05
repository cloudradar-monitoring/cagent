package cagent

import (
	"runtime"

	log "github.com/sirupsen/logrus"
)

//todo: make all JSON keys lowercase
type ProcStat struct {
	PID       int
	ParentPID int `json:"parent_PID"`
	Name      string
	Cmdline   string
	State     string
	Container string `json:",omitempty"`
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

func (ca *Cagent) ProcessesResult() (m MeasurementsMap, err error) {
	states := getPossibleProcStates()
	var procs []ProcStat

	procs, err = processes(ca.dockerWatcher)
	if err != nil {
		log.Error("[PROC] error: ", err.Error())
		return nil, err
	}
	log.Info("[PROC] results: ", len(procs))

	m = MeasurementsMap{"list": procs, "possible_states": states}

	return
}
