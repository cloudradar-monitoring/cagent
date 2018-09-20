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

func processes(fields map[string][]ProcStat) error {
	if runtime.GOOS == "linux" {
		return processesFromProc(fields)
	} else {
		return processesFromPS(fields)
	}
}

func getHostProc() string {
	procPath := "/proc"
	if os.Getenv("HOST_PROC") != "" {
		procPath = os.Getenv("HOST_PROC")
	}
	return procPath
}

// get process states from /proc/(pid)/stat
func processesFromProc(fields map[string][]ProcStat) error {
	filenames, err := filepath.Glob(getHostProc() + "/[0-9]*/stat")

	if err != nil {
		return err
	}

	for _, filename := range filenames {
		_, err := os.Stat(filename)
		data, err := readProcFile(filename)
		if err != nil {
			return err
		}
		if data == nil {
			continue
		}

		stats := bytes.Fields(data)
		if len(stats) < 3 {
			return fmt.Errorf("Something is terribly wrong with %s", filename)
		}

		pid, err := strconv.Atoi(string(stats[0]))

		if err != nil {
			log.Errorf("Failed to convert PID(%s) to int: %s", stats[0], err.Error())
		}

		comm, err := readProcFile(getHostProc() + "/" + string(stats[0]) + "/comm")

		if err != nil {
			log.Errorf("Failed to read comm: %s", stats[0], err.Error())
		}

		cmdline, err := readProcFile(getHostProc() + "/" + string(stats[0]) + "/cmdline")

		if err != nil {
			log.Errorf("Failed to read cmdline: %s", stats[0], err.Error())
		}

		stat := ProcStat{PID: pid, Name: string(bytes.TrimRight(comm, "\n")), Cmdline: strings.Replace(string(bytes.TrimRight(cmdline, "\x00")), "\x00", " ", -1)}

		switch stats[2][0] {
		case 'R':
			fields["running"] = append(fields["running"], stat)
		case 'S':
			fields["sleeping"] = append(fields["sleeping"], stat)
		case 'D':
			fields["blocked"] = append(fields["blocked"], stat)
		case 'Z':
			fields["zombies"] = append(fields["zombies"], stat)
		case 'X':
			fields["dead"] = append(fields["dead"], stat)
		case 'T', 't':
			fields["stopped"] = append(fields["stopped"], stat)
		case 'W':
			fields["paging"] = append(fields["paging"], stat)
		case 'I':
			fields["idle"] = append(fields["idle"], stat)
		}
		fields["total"] = append(fields["total"], stat)
	}
	return nil
}

func readProcFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
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

	out, err := exec.Command(bin, "axwwo", "pid,state,command").Output()
	if err != nil {
		return nil, err
	}

	return out, err
}

func processesFromPS(fields map[string][]ProcStat) error {
	out, err := execPS()
	if err != nil {
		return err
	}

	lines := strings.Split(string(out), "\n")

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

		last := strings.Join(parts[2:], " ")
		fileBaseWithArgs := filepath.Base(last)
		fileBaseParts := strings.Fields(fileBaseWithArgs)

		stat := ProcStat{PID: pid, Name: fileBaseParts[0], Cmdline: last}
		switch parts[1][0] {
		case 'W':
			fields["wait"] = append(fields["wait"], stat)
		case 'U', 'D', 'L':
			// Also known as uninterruptible sleep or disk sleep
			fields["blocked"] = append(fields["blocked"], stat)
		case 'Z':
			fields["zombies"] = append(fields["zombies"], stat)
		case 'X':
			fields["dead"] = append(fields["dead"], stat)
		case 'T':
			fields["stopped"] = append(fields["stopped"], stat)
		case 'R':
			fields["running"] = append(fields["running"], stat)
		case 'S':
			fields["sleeping"] = append(fields["sleeping"], stat)
		case 'I':
			fields["idle"] = append(fields["idle"], stat)
		case '?':
			fields["unknown"] = append(fields["unknown"], stat)
		}
		fields["total"] = append(fields["total"], stat)
	}
	return nil
}
