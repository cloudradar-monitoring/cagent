// +build !windows

package processes

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/process"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/docker"
)

var errorProcessTerminated = fmt.Errorf("Process was terminated")

type procStatus struct {
	PPID  int
	State string
}

var dockerContainerIDRE = regexp.MustCompile(`(?m)/docker/([a-f0-9]*)$`)
var monitoredProcessCache = make(map[int]*process.Process)

func processes(systemMemorySize uint64) ([]*ProcStat, error) {
	if runtime.GOOS == "linux" {
		return processesFromProc(systemMemorySize)
	}
	return processesFromPS(systemMemorySize)
}

func getProcLongState(shortState byte) string {
	switch shortState {
	case 'R':
		return "running"
	case 'S':
		return "sleeping"
	case 'D':
		return "blocked"
	case 'Z':
		return "zombie"
	case 'X':
		return "dead"
	case 'T', 't':
		return "stopped"
	case 'W':
		return "paging"
	case 'I':
		return "idle"
	default:
		return fmt.Sprintf("unknown(%s)", string(shortState))
	}
}

// get process states from /proc/(pid)/stat
func processesFromProc(systemMemorySize uint64) ([]*ProcStat, error) {
	filepaths, err := filepath.Glob(common.HostProc() + "/[0-9]*/status")
	if err != nil {
		return nil, err
	}

	var procs []*ProcStat
	var updatedProcessCache = make(map[int]*process.Process)

	for _, statusFilepath := range filepaths {
		statusFile, err := readProcFile(statusFilepath)
		if err != nil {
			if err != errorProcessTerminated {
				log.WithError(err).Error("readProcFile error")
			}
			continue
		}

		parsedProcStatus := parseProcStatusFile(statusFile)
		stat := &ProcStat{ParentPID: parsedProcStatus.PPID, State: parsedProcStatus.State}
		// get the PID from the filepath(/proc/<pid>/status) itself
		pathParts := strings.Split(statusFilepath, string(filepath.Separator))
		pidString := pathParts[len(pathParts)-2]
		stat.PID, err = strconv.Atoi(pidString)
		if err != nil {
			log.WithError(err).Errorf("proc/status: failed to convert PID(%s) to int", pidString)
		}

		commFilepath := common.HostProc() + "/" + pidString + "/comm"
		comm, err := readProcFile(commFilepath)
		if err != nil && err != errorProcessTerminated {
			log.WithError(err).Errorf("failed to read comm(%s)", commFilepath)
		} else if err == nil {
			stat.Name = string(bytes.TrimRight(comm, "\n"))
		}

		cmdLineFilepath := common.HostProc() + "/" + pidString + "/cmdline"
		cmdline, err := readProcFile(cmdLineFilepath)
		if err != nil && err != errorProcessTerminated {
			log.WithError(err).Errorf("failed to read cmdline(%s)", cmdLineFilepath)
		} else if err == nil {
			stat.Cmdline = strings.Replace(string(bytes.TrimRight(cmdline, "\x00")), "\x00", " ", -1)
		}

		cgroupFilepath := common.HostProc() + "/" + pidString + "/cgroup"
		cgroup, err := readProcFile(cgroupFilepath)
		if err != nil && err != errorProcessTerminated {
			log.WithError(err).Errorf("failed to read cgroup(%s)", cgroupFilepath)
		} else if err == nil {
			reParts := dockerContainerIDRE.FindStringSubmatch(string(cgroup))
			if len(reParts) > 0 {
				containerID := reParts[1]
				containerName, err := docker.ContainerNameByID(containerID)
				if err != nil {
					if err != docker.ErrorNotImplementedForOS && err != docker.ErrorDockerNotAvailable {
						log.WithError(err).Errorf("failed to read docker container name by id(%s)", containerID)
					}
				} else {
					stat.Container = containerName
				}
			}
		}

		statFilepath := common.HostProc() + "/" + pidString + "/stat"
		statFileContent, err := readProcFile(statFilepath)
		if err != nil && err != errorProcessTerminated {
			log.WithError(err).Errorf("failed to read stat (%s)", statFilepath)
		} else if err == nil {
			stat.ProcessGID = parseStatFileContent(statFileContent)
		}

		if stat.PID > 0 {
			p := getProcessByPID(stat.PID)
			stat.RSS, stat.VMS, stat.MemoryUsagePercent, stat.CPUAverageUsagePercent = gatherProcessResourceUsage(p, systemMemorySize)
			updatedProcessCache[stat.PID] = p
		}

		procs = append(procs, stat)
	}

	monitoredProcessCache = updatedProcessCache

	return procs, nil
}

func parseProcStatusFile(b []byte) procStatus {
	// fill default value
	// we need non-zero values in order to check if we set them, because PPID can be 0
	status := procStatus{
		PPID: -1,
	}

	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())

		switch strings.ToLower(fields[0]) {
		case "ppid:":
			var err error
			status.PPID, err = strconv.Atoi(fields[1])
			if err != nil {
				log.WithError(err).Errorf("proc/status: failed to convert PPID(%s) to int", fields[1])
			}
		case "state:":
			// extract raw long state
			// eg "State:	S (sleeping)"
			if len(fields) >= 3 {
				status.State = strings.ToLower(strings.Trim(fields[2], "()"))
				break
			}

			if len(fields) < 2 {
				break
			}
			// determine long state from the short one in case long one is not available
			// eg "State:	S"
			status.State = getProcLongState(fields[1][0])
		}

		if status.PPID >= 0 && status.State != "" {
			// we found all fields we want to
			// we can break and return
			break
		}
	}

	return status
}

func parseStatFileContent(b []byte) int {
	fields := strings.Fields(string(b))

	i := 1
	for !strings.HasSuffix(fields[i], ")") {
		i++
	}

	pgrp, err := strconv.ParseInt(fields[i+3], 10, 32)
	if err != nil {
		log.WithError(err).Errorf("proc/stat: failed to convert PGRP (%s) to int", fields[i+3])
		return -1
	}
	return int(pgrp)
}

func readProcFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		// if file doesn't exists it means that process was closed after we got the directory listing
		if os.IsNotExist(err) {
			return nil, errorProcessTerminated
		}

		// Reading from /proc/<PID> fails with ESRCH if the process has
		// been terminated between open() and read().
		if perr, ok := err.(*os.PathError); ok && perr.Err == syscall.ESRCH {
			return nil, errorProcessTerminated
		}

		return nil, err
	}

	return data, nil
}

func execPS() ([]byte, error) {
	bin, err := exec.LookPath("ps")
	if err != nil {
		return nil, err
	}

	out, err := exec.Command(bin, "axwwo", "pid,ppid,pgrp,state,command").Output()
	if err != nil {
		return nil, err
	}

	return out, err
}

func processesFromPS(systemMemorySize uint64) ([]*ProcStat, error) {
	out, err := execPS()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	var procs []*ProcStat
	var updatedProcessCache = make(map[int]*process.Process)
	var columnsIndex = map[string]int{}

	for i, line := range lines {
		parts := strings.Fields(line)

		if i == 0 {
			// parse the header
			for colIndex, colName := range parts {
				columnsIndex[strings.ToUpper(colName)] = colIndex
			}
			continue
		}

		if len(parts) < 3 {
			continue
		}

		stat := &ProcStat{}

		if pidIndex, exists := columnsIndex["PID"]; exists {
			pidString := parts[pidIndex]
			stat.PID, err = strconv.Atoi(pidString)
			if err != nil {
				log.WithError(err).Errorf("ps: failed to convert PID(%s) to int", pidString)
			}
		} else {
			// we can't set PID set to default 0 if it is unavailable for some reason, because 0 PID means the kernel(Swapper) process
			stat.PID = -1
		}

		if ppidIndex, exists := columnsIndex["PPID"]; exists {
			ppidString := parts[ppidIndex]
			stat.ParentPID, err = strconv.Atoi(ppidString)
			if err != nil {
				log.WithError(err).Errorf("ps: failed to convert PPID(%s) to int", ppidString)
			}
		} else {
			// we can't left ParentPID set to default 0 if it is unavailable for some reason, because 0 PID means the kernel task process
			stat.ParentPID = -1
		}

		if pgidIndex, exists := columnsIndex["PGRP"]; exists {
			pgidString := parts[pgidIndex]
			stat.ProcessGID, err = strconv.Atoi(pgidString)
			if err != nil {
				log.WithError(err).Errorf("ps: failed to convert PGID(%s) to int", pgidString)
			}
		} else {
			// we can't left ProcessGID set to default 0 if it is unavailable for some reason, because 0 PGID means the kernel task group
			stat.ProcessGID = -1
		}

		if statIndex, exists := columnsIndex["STAT"]; exists {
			stat.State = getProcLongState(parts[statIndex][0])
		}

		// COMMAND must be the last column otherwise we can't parse it because it can contains spaces
		if commandIndex, exists := columnsIndex["COMMAND"]; exists && commandIndex == (len(columnsIndex)-1) {
			stat.Cmdline = strings.Join(parts[commandIndex:], " ")

			// extract the executable name without the arguments
			fileBaseWithArgs := filepath.Base(stat.Cmdline)
			fileBaseParts := strings.Fields(fileBaseWithArgs)
			stat.Name = fileBaseParts[0]
		}

		if stat.PID > 0 {
			p := getProcessByPID(stat.PID)
			stat.RSS, stat.VMS, stat.MemoryUsagePercent, stat.CPUAverageUsagePercent = gatherProcessResourceUsage(p, systemMemorySize)
			updatedProcessCache[stat.PID] = p
		}

		procs = append(procs, stat)
	}
	monitoredProcessCache = updatedProcessCache

	return procs, nil
}

func getProcessByPID(pid int) *process.Process {
	p, exists := monitoredProcessCache[pid]
	if !exists {
		p = &process.Process{
			Pid: int32(pid),
		}
	}
	return p
}

func gatherProcessResourceUsage(proc *process.Process, systemMemorySize uint64) (uint64, uint64, float32, float32) {
	memoryInfo, err := proc.MemoryInfo()
	if err != nil {
		log.WithError(err).Error("failed to get memory info")
		return 0, 0, 0.0, 0.0
	}
	memUsagePercent := (float64(memoryInfo.RSS) / float64(systemMemorySize)) * 100

	// side effect: p.Percent() call update process internally
	cpuUsagePercent, err := proc.Percent(time.Duration(0))
	if err != nil {
		log.WithError(err).Error("failed to get CPU usage")
	}

	return memoryInfo.RSS, memoryInfo.VMS, float32(common.RoundToTwoDecimalPlaces(memUsagePercent)), float32(common.RoundToTwoDecimalPlaces(cpuUsagePercent))
}

func isKernelTask(procStat *ProcStat) bool {
	return procStat.ParentPID == 0 || procStat.ProcessGID == 0
}
