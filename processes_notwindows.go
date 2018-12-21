// +build !windows

package cagent

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var errorProcessTerminated = fmt.Errorf("Process was terminated")

type procStatus struct {
	PPID  int
	State string
}

func processes() ([]ProcStat, error) {
	if runtime.GOOS == "linux" {
		return processesFromProc()
	}
	return processesFromPS()
}

func getHostProc() string {
	if hostProc := os.Getenv("HOST_PROC"); hostProc != "" {
		return hostProc
	}

	return "/proc"
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
func processesFromProc() ([]ProcStat, error) {
	filepaths, err := filepath.Glob(getHostProc() + "/[0-9]*/status")
	if err != nil {
		return nil, err
	}

	var procs []ProcStat

	for _, statusFilepath := range filepaths {
		statusFile, err := readProcFile(statusFilepath)
		if err != nil {
			if err != errorProcessTerminated {
				log.Error("[PROC] readProcFile error ", err.Error())
			}
			continue
		}

		procStatus := parseProcStatusFile(statusFile)
		stat := ProcStat{ParentPID: procStatus.PPID, State: procStatus.State}
		// get the PID from the filepath(/proc/<pid>/status) itself
		pathParts := strings.Split(statusFilepath, string(filepath.Separator))
		pidString := pathParts[len(pathParts)-2]
		stat.PID, err = strconv.Atoi(pidString)
		if err != nil {
			log.Errorf("[PROC] proc/status: failed to convert PID(%s) to int: %s", pidString, err.Error())
		}

		commFilepath := getHostProc() + "/" + pidString + "/comm"
		comm, err := readProcFile(commFilepath)
		if err != nil && err != errorProcessTerminated {
			log.Errorf("[PROC] failed to read comm(%s): %s", commFilepath, err.Error())
		} else if err == nil {
			stat.Name = string(bytes.TrimRight(comm, "\n"))
		}

		cmdLineFilepath := getHostProc() + "/" + pidString + "/cmdline"
		cmdline, err := readProcFile(cmdLineFilepath)
		if err != nil && err != errorProcessTerminated {
			log.Errorf("[PROC] failed to read cmdline(%s): %s", cmdLineFilepath, err.Error())
		} else if err == nil {
			stat.Cmdline = strings.Replace(string(bytes.TrimRight(cmdline, "\x00")), "\x00", " ", -1)
		}

		procs = append(procs, stat)
	}

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
				log.Errorf("[PROC] proc/status: failed to convert PPID(%s) to int: %s", fields[1], err.Error())
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

	out, err := exec.Command(bin, "axwwo", "pid,ppid,state,command").Output()
	if err != nil {
		return nil, err
	}

	return out, err
}

func processesFromPS() ([]ProcStat, error) {
	out, err := execPS()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	var procs []ProcStat
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

		stat := ProcStat{}

		if pidIndex, exists := columnsIndex["PID"]; exists {
			pidString := parts[pidIndex]
			stat.PID, err = strconv.Atoi(pidString)
			if err != nil {
				log.Errorf("[PROC] ps: failed to convert PID(%s) to int: %s", pidString, err.Error())
			}
		} else {
			// we can't set PID set to default 0 if it is unavailable for some reason, because 0 PID means the kernel(Swapper) process
			stat.PID = -1
		}

		if ppidIndex, exists := columnsIndex["PPID"]; exists {
			ppidString := parts[ppidIndex]
			stat.ParentPID, err = strconv.Atoi(ppidString)
			if err != nil {
				log.Errorf("[PROC] ps: failed to convert PPID(%s) to int: %s", ppidString, err.Error())
			}
		} else {
			// we can't left ParentPID set to default 0 if it is unavailable for some reason, because 0 PID means the kernel(Swapper) process
			stat.ParentPID = -1
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

		procs = append(procs, stat)
	}

	return procs, nil
}
