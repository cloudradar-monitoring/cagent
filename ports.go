package cagent

import (
	"fmt"
	"syscall"

	"github.com/shirou/gopsutil/net"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/types"
)

type PortStat struct {
	Protocol     string `json:"proto"`
	LocalAddress string `json:"addr"`
	PID          int32  `json:"pid,omitempty"`
	ProgramName  string `json:"program,omitempty"`
}

// PortsResult lists all active connections
func (ca *Cagent) PortsResult(processList []ProcStat) (types.MeasurementsMap, error) {
	connections, err := net.Connections("inet")
	if err != nil {
		log.Error("[PORTS] could not list connections: ", err.Error())
		return nil, err
	}

	var ports []PortStat
	for _, conn := range connections {
		state := conn.Status
		// some connection types do not have 'state'. They are active by default.
		isActiveConnection := state == "LISTEN" || state == "NONE" || state == ""
		if !isActiveConnection {
			continue
		}

		var programName string
		if conn.Pid != 0 {
			for _, proc := range processList {
				if int32(proc.PID) == conn.Pid {
					programName = proc.Name
					break
				}
			}
		}

		ports = append(ports, PortStat{
			Protocol:     formatConnectionProtocol(conn.Family, conn.Type),
			LocalAddress: formatNetAddr(&conn.Laddr),
			PID:          conn.Pid,
			ProgramName:  programName,
		})
	}

	log.Debugf("[PORTS] results: %d", len(ports))

	return types.MeasurementsMap{"list": ports}, nil
}

func formatNetAddr(addr *net.Addr) string {
	return fmt.Sprintf("%s:%d", addr.IP, addr.Port)
}

func formatConnectionProtocol(family, socketType uint32) string {
	var version string
	if family == syscall.AF_INET6 {
		version = "6"
	}

	var baseProto string
	if socketType == syscall.SOCK_STREAM {
		baseProto = "tcp"
	} else {
		baseProto = "udp"
	}

	return baseProto + version
}
