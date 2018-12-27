// +build !windows

package services

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ErrorNotImplementedForOS returned in case we don't yet implement service manager parsing or this OS. Should be checked and ignored
var ErrorNotImplementedForOS = errors.New("Services list not implemented for " + runtime.GOOS)

// SystemdService contains the service's data parsed from systemctl
type SystemdService struct {
	UnitFile    string
	LoadState   string
	ActiveState string
	SubState    string
	Description string
}

// SysVService contains the service's data parsed from service(Sysvinit) or initctl(Upstart)
type SysVService struct {
	UnitFile string
	Status   string
}

// ListSystemdServices list Systemd services via systemctl
func ListSystemdServices() ([]SystemdService, error) {
	cmd := exec.Command("systemctl",
		"--type=service", // show only services(ignore .mount, .target, .path, .socket etc.)
		"--all",          // show loaded but inactive services too
		"--no-pager",     // disable results pagination
		"--plain",        // disable colors and status bullet
		"list-units")
	setPathEnvVar(cmd)

	var outb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &outb
	err := cmd.Run()
	if err != nil {
		errOutput, _ := ioutil.ReadAll(&outb)
		return nil, fmt.Errorf("Systemctl: %s, %s", err.Error(), string(errOutput))
	}

	var services []SystemdService
	scanner := bufio.NewScanner(&outb)
	firstRow := true
	var columns []string
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())

		if firstRow {
			// save columns to use it later
			columns = parts

			// When running `systemctl list-units` here in Golang it omits the last column's header "DESCRIPTION", but still writes the description columns with values
			// Possibly it checks if stdout is terminal
			// Workaround: add DESCRIPTION column if missed
			if columns[len(columns)-1] != "DESCRIPTION" {
				columns = append(columns, "DESCRIPTION")
			}

			firstRow = false
			continue
		}
		if len(parts) == 0 {
			// payload finished
			// need to exit the read loop
			break
		}

		rowColumnsValues := map[string]string{}
		for colIndex, colName := range columns {

			// most likely impossible but still need to check the boundaries
			if colIndex >= len(parts) {
				break
			}

			// if column is the last - join all the data left in the row because it can contain spaces
			if colIndex == len(columns)-1 {
				rowColumnsValues[colName] = strings.Join(parts[colIndex:], " ")
				break
			}

			rowColumnsValues[colName] = parts[colIndex]
		}

		services = append(services,
			SystemdService{UnitFile: rowColumnsValues["UNIT"], LoadState: rowColumnsValues["LOAD"], ActiveState: rowColumnsValues["ACTIVE"], SubState: rowColumnsValues["SUB"], Description: rowColumnsValues["DESCRIPTION"]},
		)
	}

	return services, scanner.Err()
}

var sysVinitServiceRE = regexp.MustCompile(`^\s+\[\s+([\+\-\?]])\s+\]\s+(.*)$`)

// ListSysVinitServices list SysVinit services via `service --status-all`
func ListSysVinitServices() ([]SysVService, error) {
	cmd := exec.Command("service",
		"--status-all",
	)

	setPathEnvVar(cmd)
	var outb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &outb
	err := cmd.Run()
	if err != nil {
		errOutput, _ := ioutil.ReadAll(&outb)
		return nil, fmt.Errorf("service: %s, %s", err.Error(), string(errOutput))
	}

	var services []SysVService
	scanner := bufio.NewScanner(&outb)

	for scanner.Scan() {
		parts := sysVinitServiceRE.FindStringSubmatch(scanner.Text())

		var state string

		switch parts[1] {
		case "+":
			state = "running"
		case "-":
			state = "stopped"
		case "?":
			state = "unknown"
		}

		services = append(services,
			SysVService{UnitFile: parts[2], Status: state},
		)
	}

	return services, scanner.Err()
}

// ListUpstartServices list upstart services via `initctl list`. Returns []SysVService because Upstart is compatible with SysVInit and has the same details
func ListUpstartServices() ([]SysVService, error) {
	cmd := exec.Command("initctl",
		"list",
	)
	setPathEnvVar(cmd)

	var outb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &outb
	err := cmd.Run()
	if err != nil {
		errOutput, _ := ioutil.ReadAll(&outb)
		return nil, fmt.Errorf("Initctl: %s, %s", err.Error(), string(errOutput))
	}

	var services []SysVService
	scanner := bufio.NewScanner(&outb)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())

		if len(parts) < 2 {
			continue
		}

		// filter network interfaces from results e.g.
		// network-interface-security (network-interface/ip6_vti0) start/running
		// network-interface (dummy0) start/running
		if strings.HasPrefix(parts[0], "network-interface") {
			continue
		}

		stateParts := strings.Split(parts[1], "/")

		services = append(services,
			SysVService{UnitFile: parts[0], Status: stateParts[1]},
		)
	}

	return services, scanner.Err()
}

func isSystemd() bool {
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return true
	}
	return false
}

func isUpstart() bool {
	if _, err := os.Stat("/sbin/upstart-udev-bridge"); err == nil {
		return true
	}
	cmd := exec.Command("initctl", "--version")
	setPathEnvVar(cmd)
	if out, err := cmd.Output(); err == nil {
		if strings.Contains(string(out), "initctl (upstart") {
			return true
		}
	}

	return false
}

func setPathEnvVar(cmd *exec.Cmd) {
	// path PATH env var to be able to found right exec location
	cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH"))
}

// ListServices detect the linux system manager and parse/combine results
func ListServices() (map[string]interface{}, error) {
	if runtime.GOOS != "linux" {
		return nil, ErrorNotImplementedForOS
	}

	if isSystemd() {
		services, err := ListSystemdServices()
		if err != nil {
			// in case of error lets try to query other
			log.Errorf("[Services] Systemd appears running but failed to list a services: %s", err.Error())
		} else {
			var servicesList []map[string]interface{}
			for _, service := range services {
				servicesList = append(servicesList,
					map[string]interface{}{
						"name":         service.UnitFile,
						"load_state":   service.LoadState,
						"active_state": service.ActiveState,
						"sub_state":    service.SubState,
						"description":  service.Description,
						"manager":      "systemd",
					})
			}
			// if systemd returns results we can return here cause systemd can't work simultaneously with sysvinit/upstart
			return map[string]interface{}{"list": servicesList}, err
		}
	}

	sysVServices, err := ListSysVinitServices()
	if err != nil {
		// in case of error lets try to query other
		log.Errorf("[Services] SysVinit: failed to list a services: %s", err.Error())
	}

	// create the map to check the services by name
	servicesMap := map[string]map[string]interface{}{}
	for _, service := range sysVServices {
		servicesMap[service.UnitFile] = map[string]interface{}{
			"name":    service.UnitFile,
			"status":  service.Status,
			"manager": "sysvinit",
		}
	}

	if isUpstart() {
		upstartServices, err := ListUpstartServices()
		if err != nil {
			// in case of error lets try to query other
			log.Errorf("[Services] Upstart: failed to list a services via initctl: %s", err.Error())
		}

		for _, service := range upstartServices {
			// overwrite sysVInit services returned by 'service --status-all'
			// because it also contains services run under Upstart but with more accurate status
			servicesMap[service.UnitFile] = map[string]interface{}{
				"name":    service.UnitFile,
				"status":  service.Status,
				"manager": "upstart",
			}
		}
	}

	var servicesList []map[string]interface{}
	for _, service := range servicesMap {
		servicesList = append(servicesList, service)
	}

	return map[string]interface{}{"list": servicesList}, nil
}
