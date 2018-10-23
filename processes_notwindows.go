// +build !windows

package cagent

import (
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

func processes() ([]ProcStat, error) {
	if runtime.GOOS == "linux" {
		return processesFromProc()
	}
	return processesFromPS()
}

func getHostProc() string {
	procPath := "/proc"

	if os.Getenv("HOST_PROC") != "" {
		procPath = os.Getenv("HOST_PROC")
	}

	return procPath
}

// get process states from /proc/(pid)/stat
func processesFromProc() ([]ProcStat, error) {
	filenames, err := filepath.Glob(getHostProc() + "/[0-9]*/stat")
	if err != nil {
		return nil, err
	}

	var procs []ProcStat

	for _, filename := range filenames {
		data, err := readProcFile(filename)
		if err != nil {
			return nil, err
		}

		if data == nil {
			continue
		}

		stats := bytes.Fields(data)

		if len(stats) < 3 {
			return nil, fmt.Errorf("Something is terribly wrong with %s", filename)
		}

		pid, err := strconv.Atoi(string(stats[0]))
		if err != nil {
			log.Errorf("Failed to convert PID(%s) to int: %s", stats[0], err.Error())
		}

		comm, err := readProcFile(getHostProc() + "/" + string(stats[0]) + "/comm")
		if err != nil {
			log.Errorf("Failed to read comm(%s): %s", stats[0], err.Error())
		}

		cmdline, err := readProcFile(getHostProc() + "/" + string(stats[0]) + "/cmdline")
		if err != nil {
			log.Errorf("Failed to read cmdline(%s): %s", stats[0], err.Error())
		}

		ppid, err := strconv.Atoi(string(stats[4]))
		if err != nil {
			log.Errorf("Failed to convert PPID(%s) to int: %s", stats[4], err.Error())
		}

		stat := ProcStat{
			PID:       pid,
			ParentPID: ppid,
			Name:      string(bytes.TrimRight(comm, "\n")),
			Cmdline:   strings.Replace(string(bytes.TrimRight(cmdline, "\x00")), "\x00", " ", -1),
		}

		switch stats[2][0] {
		case 'R':
			stat.State = "running"
		case 'S':
			stat.State = "sleeping"
		case 'D':
			stat.State = "blocked"
		case 'Z':
			stat.State = "zombie"
		case 'X':
			stat.State = "dead"
		case 'T', 't':
			stat.State = "stopped"
		case 'W':
			stat.State = "paging"
		case 'I':
			stat.State = "idle"
		default:
			stat.State = "unknown"
		}
		procs = append(procs, stat)
	}

	return procs, nil
}

func readProcFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// if file doesn't exists
			return nil, nil
		}

		// Reading from /proc/<PID> fails with ESRCH if the process has
		// been terminated between open() and read().
		if perr, ok := err.(*os.PathError); ok && perr.Err == syscall.ESRCH {
			return nil, nil
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

	for i, line := range lines {
		if i == 0 {
			// skip the header
			continue
		}

		parts := strings.Fields(line)

		if len(parts) < 3 {
			continue
		}

		pid, err := strconv.Atoi(string(parts[0]))
		if err != nil {
			log.Errorf("Failed to convert PID(%s) to int: %s", parts[0], err.Error())
		}

		last := strings.Join(parts[3:], " ")
		fileBaseWithArgs := filepath.Base(last)
		fileBaseParts := strings.Fields(fileBaseWithArgs)
		ppid, err := strconv.Atoi(string(parts[1]))

		if err != nil {
			log.Errorf("Failed to convert PPID(%s) to int: %s", parts[4], err.Error())
		}

		stat := ProcStat{PID: pid, ParentPID: ppid, Name: fileBaseParts[0], Cmdline: last}
		switch parts[2][0] {
		case 'W':
			stat.State = "wait"
		case 'U', 'D', 'L':
			// Also known as uninterruptible sleep or disk sleep
			stat.State = "blocked"
		case 'Z':
			stat.State = "zombie"
		case 'X':
			stat.State = "dead"
		case 'T':
			stat.State = "stopped"
		case 'R':
			stat.State = "running"
		case 'S':
			stat.State = "sleeping"
		case 'I':
			stat.State = "idle"
		default:
			stat.State = "unknown"
		}

		procs = append(procs, stat)
	}

	return procs, nil
}
