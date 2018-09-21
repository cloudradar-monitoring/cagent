package cagent

import (
	"runtime"

	log "github.com/sirupsen/logrus"
)

type ProcStat struct {
	PID     int
	Name    string
	Cmdline string
}

// Gets empty fields of metrics based on the OS
func getEmptyFields() map[string][]ProcStat {
	fields := map[string][]ProcStat{
		"blocked":  {},
		"zombies":  {},
		"stopped":  {},
		"running":  {},
		"sleeping": {},
		"total":    {},
	}
	switch runtime.GOOS {
	case "windows":
		fields = map[string][]ProcStat{"running": {}}
	case "freebsd":
		fields["idle"] = []ProcStat{}
		fields["wait"] = []ProcStat{}
	case "darwin":
		fields["idle"] = []ProcStat{}
	case "openbsd":
		fields["idle"] = []ProcStat{}
	case "linux":
		fields["dead"] = []ProcStat{}
		fields["paging"] = []ProcStat{}
		fields["idle"] = []ProcStat{}
	}
	return fields
}

func (ca *Cagent) ProcessesResult() (m MeasurementsMap, err error) {
	fields := getEmptyFields()
	m = make(MeasurementsMap)
	err = processes(fields)

	if err != nil {
		log.Error("[PROC] error: ", err.Error())
		return nil, err
	}
	log.Info("[PROC] results: ", len(fields))

	for key, val := range fields {
		m[key+".list"] = val
		m[key+".num"] = len(val)
	}
	return
}
